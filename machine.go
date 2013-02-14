package golog

import . "github.com/mndrix/golog/term"

import "github.com/mndrix/golog/read"
import "github.com/mndrix/golog/prelude"
import "github.com/mndrix/ps"

import "bytes"
import "fmt"

type Machine interface {
    ForeignReturn

    // these three should be functions that take a Machine rather than methods
    CanProve(interface{}) bool
    Consult(interface{}) Machine
    ProveAll(interface{}) []Bindings

    String() string

    // Bindings returns the machine's most current variable bindings.
    //
    // This method is typically only used by ChoicePoint implementations
    Bindings() Bindings

    // SetBindings returns a new machine like this one but with the given
    // bindings
    SetBindings(Bindings) Machine

    // PushConj returns a machine like this one but with an extra term
    // on front of the conjunction stack
    PushConj(Term) Machine

    // PopConj returns a machine with one fewer items on the conjunction stack
    // along with the term removed.  Returns err = EmptyConjunctions if there
    // are no more conjunctions on that stack
    PopConj() (Term, Machine, error)

    // PushCutBarrier pushes a special marker onto the disjunction stack.
    // This marker can be used to locate which disjunctions came immediately
    // before the marker existed.
    PushCutBarrier() Machine

    // MostRecentCutBarrier returns an opaque value which uniquely
    // identifies the most recent cut barrier in the disjunction stack.
    // Used with CutTo to remove several disjunctions at once.
    // Returns NoBarriers if there are no cut barriers on the disjunctions
    // stack.
    MostRecentCutBarrier() (int64, error)

    // CutTo removes all disjunctions stacked on top of a specific cut barrier.
    // It does not remove the cut barrier itself.
    // A barrier ID is obtained from MostRecentCutBarrier.
    CutTo(int64) Machine

    // PushDisj returns a machine like this one but with an extra ChoicePoint
    // on the disjunctions stack.
    PushDisj(ChoicePoint) Machine

    // PopDisj returns a machine with one fewer choice points on the
    // disjunction stack and the choice point that was removed.  Returns
    // err = EmptyDisjunctions if there are no more disjunctions on
    // that stack
    PopDisj() (ChoicePoint, Machine, error)

    // RegisterForeign registers Go functions to implement Golog predicates.
    // When Golog tries to prove a predicate with one of these predicate
    // indicators, it executes the given function instead.
    // Calling RegisterForeign with a predicate indicator that's already
    // been registered replaces the predicate implementation.
    RegisterForeign(map[string]ForeignPredicate) Machine

    // Step advances the machine one "step" (implementation dependent).
    // It produces a new machine which can take the next step.  It might
    // produce a proof by giving some variable bindings.  When the machine
    // has done as much work as it can do, it returns err=MachineDone
    Step() (Machine, Bindings, error)
}

// ForeignPredicate is the type of functions which implement Golog predicates
// that are defined in Go
type ForeignPredicate func(Machine, []Term) ForeignReturn

type machine struct {
    db      Database
    foreign ps.Map      // predicate indicator => ForeignPredicate
    env     Bindings
    disjs   ps.List     // of ChoicePoint
    conjs   ps.List     // of Term
}
func (*machine) IsaForeignReturn() {}

// NewMachine creates a new Golog machine.  This machine has the standard
// library already loaded and is typically the way one wants to obtain
// a machine.
func NewMachine() Machine {
    return NewBlankMachine().
            Consult(prelude.Prelude).
            RegisterForeign(map[string]ForeignPredicate{
                "!/0" :         BuiltinCut,
                "$cut_to/1" :   BuiltinCutTo,
                ",/2" :         BuiltinComma,
                "->/2" :        BuiltinIfThen,
                ";/2" :         BuiltinSemicolon,
                "=/2" :         BuiltinUnify,
                "atom_codes/2": BuiltinAtomCodes2,
                "call/1" :      BuiltinCall,
                "call/2" :      BuiltinCall,
                "call/3" :      BuiltinCall,
                "call/4" :      BuiltinCall,
                "call/5" :      BuiltinCall,
                "call/6" :      BuiltinCall,
                "downcase_atom/2":  BuiltinDowncaseAtom2,
                "fail/0" :      BuiltinFail,
                "findall/3" :   BuiltinFindall3,
                "listing/0" :   BuiltinListing0,
                "msort/2" :     BuiltinMsort2,
            })
}

// NewBlankMachine creates a new Golog machine without loading the
// standard library (prelude)
func NewBlankMachine() Machine {
    var m machine
    m.db = NewDatabase()
    m.foreign = ps.NewMap()
    m.env = NewBindings()
    m.disjs = ps.NewList()
    m.conjs = ps.NewList()
    return (&m).PushCutBarrier()
}

func (m *machine) clone() *machine {
    var m1 machine
    m1.db       = m.db
    m1.foreign  = m.foreign
    m1.env      = m.env
    m1.disjs    = m.disjs
    m1.conjs    = m.conjs
    return &m1
}

func (m *machine) Consult(text interface{}) Machine {
    terms := read.TermAll_(text)
    m1 := m.clone()
    for _, t := range terms {
        if IsDirective(t) {
            // ignore all directives, for now
            continue
        }
        m1.db = m1.db.Assertz(t)
    }
    return m1
}

func (m *machine) RegisterForeign(fs map[string]ForeignPredicate) Machine {
    m1 := m.clone()
    for indicator, f := range fs {
        m1.foreign = m1.foreign.Set(indicator, f)
    }
    return m1
}

func (m *machine) String() string {
    var buf bytes.Buffer
    fmt.Fprintf(&buf, "disjs:\n")
    m.disjs.ForEach( func (v ps.Any) {
        fmt.Fprintf(&buf, "  %s\n", v)
    })
    fmt.Fprintf(&buf, "conjs:\n")
    m.conjs.ForEach( func (v ps.Any) {
        fmt.Fprintf(&buf, "  %s\n", v)
    })
    fmt.Fprintf(&buf, "bindings: %s", m.env)
    return buf.String()
}

// CanProve returns true if goal can be proven from facts and clauses
// in the database
func (m *machine) CanProve(goal interface{}) bool {
    solutions := m.ProveAll(goal)
    return len(solutions) > 0
}

var MachineDone = fmt.Errorf("Machine can't step any further")
func (self *machine) ProveAll(goal interface{}) []Bindings {
    var answer Bindings
    var err error
    answers := make([]Bindings, 0)

    goalTerm := self.toGoal(goal)
    vars := Variables(goalTerm)  // preserve incoming human-readable names
    m := self.PushConj(goalTerm)
    for {
        m, answer, err = m.Step()
        if err == MachineDone {
            break
        }
        maybePanic(err)
        if answer != nil {
            answers = append(answers, answer.WithNames(vars))
        }
    }

    return answers
}

// advance the Golog machine one step closer to proving the goal at hand.
// at the end of each invocation, the top item on the conjunctions stack
// is the goal we should next try to prove.
func (self *machine) Step() (Machine, Bindings, error) {
    var m Machine = self
    var goal Term
    var err error
    var cp ChoicePoint

//  fmt.Printf("stepping...\n%s\n", self)

    // find a goal other than true/0 to prove
    indicator := "true/0"
    for indicator == "true/0" {
        var mTmp Machine
        goal, mTmp, err = m.PopConj()
        if err == EmptyConjunctions {   // found an answer
            answer := m.Bindings()
//          fmt.Printf("  emitting answer %s\n", answer)
            m = m.PushConj(NewTerm("fail"))  // backtrack on next Step()
            return m, answer, nil
        }
        maybePanic(err)
        m = mTmp
        indicator = goal.Indicator()
    }

    // are we proving a foreign predicate?
    f, ok := m.(*machine).foreign.Lookup(indicator)
    if ok {     // foreign predicate
//      fmt.Printf("  running foreign predicate %s\n", indicator)
        args := m.(*machine).resolveAllArguments(goal)
        ret := f.(ForeignPredicate)(m, args)
        switch x := ret.(type) {
            case *foreignTrue:
                return m, nil, nil
            case *foreignFail:
                // do nothing. continue to iterate disjunctions below
            case *machine:
                return x, nil, nil
            case *foreignUnify:
                terms := []Term(*x)  // guaranteed even number of elements
                env := m.Bindings()
                for i := 0; i<len(terms); i+=2 {
                    env, err = Unify(env, terms[i], terms[i+1])
                    if err == CantUnify {
                        env = nil
                        break
                    }
                    maybePanic(err)
                }
                if env != nil {
                    return m.SetBindings(env), nil, nil
                }
        }
    } else {    // user-defined predicate, push all its disjunctions
//      fmt.Printf("  running user-defined predicate %s\n", indicator)
        clauses, err := m.(*machine).db.Candidates(goal)
        maybePanic(err)
        m = m.PushCutBarrier()
        for i:=len(clauses)-1; i>=0; i-- {
            clause := clauses[i]
            cp := NewHeadBodyChoicePoint(m, goal, clause)
            m = m.PushDisj(cp)
        }
    }

    // iterate disjunctions looking for one that succeeds
    for {
        cp, m, err = m.PopDisj()
        if err == EmptyDisjunctions {   // nothing left to do
            return nil, nil, MachineDone
        }
        maybePanic(err)

        // follow the next choice point
        mTmp, err := cp.Follow()
        if err == nil {
//          fmt.Printf("  followed CP %s\n", cp)
            return mTmp, nil, nil
        }
    }

    panic("Stepped a machine past the end")
}

func (m *machine) toGoal(thing interface{}) Term {
    switch x := thing.(type) {
        case Term:
            return x
        case string:
            return m.readTerm(x)
    }
    msg := fmt.Sprintf("Can't convert %#v to a term", thing)
    panic(msg)
}

func (m *machine) resolveAllArguments(goal Term) []Term {
    env := m.Bindings()
    args := goal.Arguments()
    resolved := make([]Term, len(args))
    for i, arg := range args {
        if IsVariable(arg) {
            a, err := env.Resolve(arg.(*Variable))
            if err == nil {
                resolved[i] = a
                continue
            }
        }
        resolved[i] = arg
    }

    return resolved
}

func (m *machine) readTerm(src interface{}) Term {
    return read.Term_(src)
}

func (m *machine) Bindings() Bindings {
    return m.env
}

func (m *machine) SetBindings(env Bindings) Machine {
    m1 := m.clone()
    m1.env = env
    return m1
}

func (m *machine) PushConj(t Term) Machine {
    // change all !/0 goals into '$cut_to'(RecentBarrierId) goals
    barrierID, err := m.MostRecentCutBarrier()
    if err == nil {
        t = resolveCuts(barrierID, t)
    }

    m1 := m.clone()
    m1.conjs = m.conjs.Cons(t)
    return m1
}

var EmptyConjunctions = fmt.Errorf("Conjunctions list is empty")
func (m *machine) PopConj() (Term, Machine, error) {
    if m.conjs.IsNil() {
        return nil, nil, EmptyConjunctions
    }

    t := m.conjs.Head().(Term)
    m1 := m.clone()
    m1.conjs = m.conjs.Tail()
    return t, m1, nil
}

func (m *machine) PushDisj(cp ChoicePoint) Machine {
    m1 := m.clone()
    m1.disjs = m.disjs.Cons(cp)
    return m1
}

var EmptyDisjunctions = fmt.Errorf("Disjunctions list is empty")
func (m *machine) PopDisj() (ChoicePoint, Machine, error) {
    if m.disjs.IsNil() {
        return nil, nil, EmptyDisjunctions
    }

    cp := m.disjs.Head().(ChoicePoint)
    m1 := m.clone()
    m1.disjs = m.disjs.Tail()
    return cp, m1, nil
}

func (m *machine) PushCutBarrier() Machine {
    barrier := NewCutBarrier(m)
    return m.PushDisj(barrier)
}

var NoBarriers = fmt.Errorf("There are no cut barriers")
func (m *machine) MostRecentCutBarrier() (int64, error) {
    ds := m.disjs
    for {
        if ds.IsNil() {
            return -1, NoBarriers
        }

        id, ok := BarrierId(ds.Head().(ChoicePoint))
        if ok {
            return id, nil
        }

        ds = ds.Tail()
    }
    panic("inconceivable!")
}

func (m *machine) CutTo(want int64) Machine {
    ds := m.disjs
    for {
        if ds.IsNil() {
            msg := fmt.Sprintf("No cut barrier with ID %d", want)
            panic(msg)
        }

        found, ok := BarrierId(ds.Head().(ChoicePoint))
        if ok && found == want {
            m1 := m.clone()
            m1.disjs = ds
            return m1
        }

        ds = ds.Tail()
    }
    panic("inconceivable!")
}

func resolveCuts(id int64, t Term) Term {
    switch t.Indicator() {
        case "!/0":
            return NewTerm("$cut_to", NewInt64(id))
        case ",/2", ";/2":
            args := t.Arguments()
            t0 := resolveCuts(id, args[0])
            t1 := resolveCuts(id, args[1])
            return NewTerm( t.Functor(), t0, t1 )
        case "->/2":
            args := t.Arguments()
            t0 := args[0]   // don't resolve cuts in Condition
            t1 := resolveCuts(id, args[1])
            return NewTerm( t.Functor(), t0, t1 )
    }

    // leave any other cuts unresolved
    return t
}

func maybePanic(err error) {
    if err != nil {
        panic(err)
    }
}
