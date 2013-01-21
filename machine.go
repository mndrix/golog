package golog

import . "github.com/mndrix/golog/term"

import "github.com/mndrix/golog/read"
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
    PushGoal(Term, Bindings) Machine
}

type machine struct {
    db      Database
    stack   Frame       // top frame in the call stack
}

func NewMachine() Machine {
    var m machine
    m.db = NewDatabase()
    m.stack = NewFrame()  // an empty stack frame at the bottom
    return &m
}

func (m *machine) clone() *machine {
    var m1 machine
    m1.db = m.db
    m1.stack = m.stack
    return &m1
}

func (m *machine) Consult(text interface{}) Machine {
    terms := read.TermAll_(text)
    m1 := m.clone()
    for _, t := range terms {
        m1.db = m1.db.Assertz(t)
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

    m = m.PushGoal(m.toGoal(goal), nil).(*machine)
    for {
        m, answer, err = m.step()
        if err == MachineDone {
            break
        }
        maybePanic(err)
        if answer != nil {
            answers = append(answers, answer)
        }
    }

    return answers
}

// advance the Golog machine one step closer toward proving the goal at hand
func (m *machine) step() (*machine, Bindings, error) {
    frame := m.peekStack()
    m1 := m.clone()

    // there's no work to do for the bottom stack frame
    if IsBottom(frame) {
        return m, nil, MachineDone
    }

    // handle built ins
    switch frame.Goal().Indicator() {
        case ",/2":
            args := frame.Goal().Arguments()
            conjs := ps.NewList().Cons(args[1])
            disjs := m.candidates(args[0])
            frame1 := frame.NewSibling(args[0], nil, conjs, disjs)
            m1.stack = frame1
            return m1, nil, nil
        case "true/0":
            if frame.HasConjunctions() {  // prove next conjunction
                goal, frame1 := frame.TakeConjunction()
                disjs := m.candidates(goal)
                frame2 := frame1.NewSibling(goal, nil, nil, disjs)
                m1.stack = frame2
                return m1, nil, nil
            } else {  // reached a leaf. emit a solution
                m2 := m.BackTrack().(*machine)
                return m2, frame.Env(), nil
            }
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

func (m *machine) peekStack() Frame {
    return m.stack
}

// pushGoal returns a new machine with this goal added to the call stack.
// it handles adding choice points, if necessary
func (m *machine) PushGoal(goal Term, env Bindings) Machine {
    m1 := m.clone()
    disjs := m.candidates(goal)
    top := m.peekStack()
    m1.stack = top.NewChild(goal, env, nil, disjs)
    return m1
}

func (m *machine) candidates(goal Term) *ps.List {
    candidates := m.db.Candidates(goal)
    disjs := ps.NewList()
    for i := len(candidates) - 1; i>=0; i-- {
        cp := NewSimpleChoicePoint(candidates[i])
        disjs = disjs.Cons(cp)
    }
    return disjs
}

func (m *machine) readTerm(src interface{}) Term {
    return read.Term_(src)
}

func (m *machine) Bindings() Bindings {
    return m.peekStack().Env()
}

func (m *machine) Goal() Term {
    return m.peekStack().Goal()
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
