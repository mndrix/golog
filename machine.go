package golog

import . "github.com/mndrix/golog/term"

import "github.com/mndrix/golog/read"
import "github.com/mndrix/golog/prelude"
import "github.com/mndrix/ps"

import "fmt"

type Machine interface {
    CanProve(interface{}) bool
    Consult(interface{}) Machine
    ProveAll(interface{}) []Bindings
    String() string

    // BackTrack produces a new machine which has back tracked to the most
    // recently created choice point.  Backtracking from the bottom of the
    // call stack is a noop.
    //
    // This method is typically only used by ChoicePoint implementations
    BackTrack() Machine

    // Bindings returns the machine's most current variable bindings.
    //
    // This method is typically only used by ChoicePoint implementations
    Bindings() Bindings

    // Goal returns the goal that this machine is trying to prove.
    // Returns nil if it's not proving anything.
    //
    // This method is typically only used by ChoicePoint implementations
    Goal() Term

    // PushGoal produces a new machine which, on its next step,
    // tries to prove the given goal in the context of the given
    // bindings.  The bindings may be nil to retain the machine's
    // current bindings.  After proving this new goal, the new machine
    // continues proving what the old machine would have proved.
    //
    // This method is typically only used by ChoicePoint implementations
    PushGoal(Term, Bindings) (Machine, error)

    // RegisterForeign registers Go functions to implement Golog predicates.
    // When Golog tries to prove a predicate with one of these predicate
    // indicators, it executes the given function instead.
    // Calling RegisterForeign with a predicate indicator that's already
    // been registered replaces the predicate implementation.
    RegisterForeign(map[string]ForeignPredicate) Machine

    // Stack returns the machine's top stack frame
    Stack() Frame

    // SetStack returns a new machine whose top stack frame is the one given
    SetStack(Frame) Machine
}

// ForeignPredicate is the type of functions which implement Golog predicates
// that are defined in Go
type ForeignPredicate func(Machine, []Term) (bool,Machine)

type machine struct {
    db      Database
    stack   Frame       // top frame in the call stack
    foreign ps.Map      // predicate indicator => ForeignPredicate
}

// NewMachine creates a new Golog machine.  This machine has the standard
// library already loaded and is typically the way one wants to obtain
// a machine.
func NewMachine() Machine {
    return NewBlankMachine().Consult(prelude.Prelude)
}

// NewBlankMachine creates a new Golog machine without loading the
// standard library (prelude)
func NewBlankMachine() Machine {
    var m machine
    m.db = NewDatabase()
    m.stack = NewFrame()  // an empty stack frame at the bottom
    m.foreign = ps.NewMap()
    return &m
}

func (m *machine) clone() *machine {
    var m1 machine
    m1.db = m.db
    m1.stack = m.stack
    m1.foreign = m.foreign
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
    return m.db.String()
}

// IsTrue returns true if goal can be proven from facts and clauses
// in the database
func (m *machine) CanProve(goal interface{}) bool {
    solutions := m.ProveAll(goal)
    return len(solutions) > 0
}

var MachineDone = fmt.Errorf("Machine can't step any further")
var PleaseBackTrack = fmt.Errorf("please backtrack")  // used by ChoicePoint implementations
func (m *machine) ProveAll(goal interface{}) []Bindings {
    var answer Bindings
    var err error
    answers := make([]Bindings, 0)

    goalTerm := m.toGoal(goal)
    vars := Variables(goalTerm)  // preserve incoming human-readable names
    me, err := m.PushGoal(goalTerm, nil)
    maybePanic(err)
    m = me.(*machine)
    for {
        m, answer, err = m.step()
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

// advance the Golog machine one step closer toward proving the goal at hand
func (m *machine) step() (*machine, Bindings, error) {
    frame := m.Stack()
    m1 := m.clone()

    // there's no work to do for the bottom stack frame
    if IsBottom(frame) {
        return m, nil, MachineDone
    }

    // handle built ins
    indicator := frame.Goal().Indicator()
    v, ok := m.foreign.Lookup(frame.Goal().Indicator())
    if ok {
        f := v.(ForeignPredicate)
        success, foreignM := f(m, nil)
        if success {
            if foreignM != nil {
                m1 = foreignM.(*machine)
            }
            indicator = "true/0"    // lies!
        } else {
            panic("foreign predicates can't fail yet")
        }
    }
    switch indicator {
        case "!/0":
            frame = frame.CutChoicePoints()
            m1.stack = frame
            fallthrough
        case "true/0":
            if frame.HasConjunctions() {  // prove next conjunction
                goal, frame1 := frame.TakeConjunction()
                disjs, err := m.candidates(goal)
                if err != nil { return m, nil, err }
                frame2 := frame1.NewSibling(goal, nil, nil, disjs)
                m1.stack = frame2
                return m1, nil, nil
            } else {  // reached a leaf. emit a solution
                m2 := m1.BackTrack().(*machine)
                return m2, frame.Env(), nil
            }
        case ",/2":
            panic("Should never execute ,/2")
    }

    // have we exhausted all choice points in this frame?
    choicePoint, frame1 := frame.TakeChoicePoint()
    if choicePoint == nil {
        m1 := m.BackTrack().(*machine)
        return m1, nil, nil
    }
    m1.stack = frame1

    // we found a choice point.  try it
    m2, err := choicePoint.Follow(m1)
    if err == PleaseBackTrack {
        m3 := m1.BackTrack().(*machine)
        return m3, nil, nil
    }
    maybePanic(err)
    return m2.(*machine), nil, nil
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

func (m *machine) Stack() Frame {
    return m.stack
}

func (m *machine) SetStack(f Frame) Machine {
    m1 := m.clone()
    m1.stack = f
    return m1
}

// pushGoal returns a new machine with this goal added to the call stack.
// it handles adding choice points, if necessary
func (m *machine) PushGoal(goal Term, env Bindings) (Machine,error) {
    var conjs ps.List
    m1 := m.clone()

    // expand ,/2 into a list of conjunctions
    if goal.Indicator() == ",/2" {
        conjs = commaList(goal.Arguments()[1])
        goal = goal.Arguments()[0]
    }

    var disjs ps.List
    var err error
    if m.IsBuiltin(goal) {
        disjs = ps.NewList()
    } else {
        disjs, err = m.candidates(goal)
        if err != nil { return m, err }
    }
    top := m.Stack()
    m1.stack = top.NewChild(goal, env, conjs, disjs)
    if !isControl(goal) {
        m1.stack = m1.stack.StopCut()
    }
    return m1, nil
}

// converts a nested comma term (like those parsed from a clause body)
// into a list of non-comma terms
func commaList(t Term) ps.List {
    if t.Indicator() != ",/2" {
        return ps.NewList().Cons(t)
    }

    args := t.Arguments()
    return commaList(args[1]).Cons(args[0])
}

// true if goal is a control predicate
func isControl(goal Term) bool {
    switch goal.Indicator() {
        case ",/2": return true
        case "!/0": return true
        case "true/0": return true
    }
    return false
}

func (m *machine) IsBuiltin(goal Term) bool {
    indicator := goal.Indicator()
    switch indicator {
        case "true/0", "listing/0", "!/0":
            return true
    }
    _, ok := m.foreign.Lookup(indicator)
    return ok
}

func (m *machine) candidates(goal Term) (ps.List, error) {
    candidates, err := m.db.Candidates(goal)
    if err != nil { return nil, err }
    disjs := ps.NewList()
    for i := len(candidates) - 1; i>=0; i-- {
        cp := NewSimpleChoicePoint(candidates[i])
        disjs = disjs.Cons(cp)
    }
    return disjs, nil
}

func (m *machine) readTerm(src interface{}) Term {
    return read.Term_(src)
}

func (m *machine) Bindings() Bindings {
    return m.Stack().Env()
}

func (m *machine) Goal() Term {
    return m.Stack().Goal()
}

func (m *machine) BackTrack() Machine {
    m1 := m.clone()

    for !IsBottom(m1.stack) && !m1.stack.HasChoicePoint() {
        m1.stack = m1.stack.Parent()
    }

    return m1
}

func maybePanic(err error) {
    if err != nil {
        panic(err)
    }
}
