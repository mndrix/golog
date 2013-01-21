package golog

import . "github.com/mndrix/golog/term"

import "github.com/mndrix/ps"

// Frame represents a single frame in a Golog machine's call stack
type Frame interface {
    // Env returns this stack frame's current bindings
    Env() Bindings

    // Goal returns this stack frame's current goal
    Goal() Term

    // HasChoicePoint returns true if this stack frame has unexplored
    // choicepoints
    HasChoicePoint() bool

    // HasConjunctions returns true if this stack frame has unexplored
    // conjunctions (continuations)
    HasConjunctions() bool

    // NewChild creates a child stack frame similar to this one with
    // this one as the child's parent
    NewChild(Term, Bindings, *ps.List, *ps.List) Frame

    // NewSibling is like NewChild but the new stack frame has the same
    // parent as the invocant
    NewSibling(Term, Bindings, *ps.List, *ps.List) Frame

    // Parent returns this stack frame's parent stack frame
    Parent() Frame

    // TakeChoicePoint returns the next choice point in line
    // (nil if there isn't one) and a new frame with the remaining
    // choice points intact
    TakeChoicePoint() (ChoicePoint, Frame)

    // TakeConjunction returns the next conjunction in line
    // (panics if there are no conjunctions available) and a new frame
    // with the remaining conjunctions intact
    TakeConjunction() (Term, Frame)
}

var frameBottom frame
func init() {
    frameBottom.parent = &frameBottom  // infinite parental loop
    frameBottom.env = NewBindings()
    frameBottom.disjs = ps.NewList()
    frameBottom.conjs = ps.NewList()
}

// NewFrame returns a new, empty stack from
func NewFrame() Frame {
    return &frameBottom
}

type frame struct {
    env     Bindings
    parent  *frame
    goal    Term
    disjs   *ps.List     // of ChoicePoint
    conjs   *ps.List     // of Continuation
}

func (f *frame) clone() *frame {
    var f1 frame
    f1.env = f.env
    f1.parent = f.parent
    f1.goal = f.goal
    f1.disjs = f.disjs
    f1.conjs = f.conjs
    return &f1
}

func (f *frame) NewSibling(goal Term, env Bindings, conjs, disjs *ps.List) Frame {
    newb := f.clone()
    newb.goal = goal
    if env != nil {
        newb.env = env
    }
    if conjs != nil {
        newb.conjs = conjs
    }
    if disjs != nil {
        newb.disjs = disjs
    }
    return newb
}

func (f *frame) NewChild(goal Term, env Bindings, conjs, disjs *ps.List) Frame {
    child := f.NewSibling(goal, env, conjs, disjs).(*frame)
    child.parent = f
    return child
}

func (f *frame) TakeChoicePoint() (ChoicePoint, Frame) {
    if f.disjs == nil || f.disjs.IsNil() {
        return nil, nil
    }

    f1 := f.clone()
    f1.disjs = f.disjs.Tail()
    return f.disjs.Head().(ChoicePoint), f1
}

func (f *frame) TakeConjunction() (Term, Frame) {
    goal := f.conjs.Head().(Term)
    f1 := f.clone()
    f1.conjs = f.conjs.Tail()
    return goal, f1
}

func (f *frame) Env() Bindings {
    return f.env
}

func (f *frame) Goal() Term {
    return f.goal
}

func (f *frame) Parent() Frame {
    return f.parent
}

func (f *frame) HasChoicePoint() bool {
    return f.disjs != nil && !f.disjs.IsNil()
}

func (f *frame) HasConjunctions() bool {
    return !f.conjs.IsNil()
}

// is this the bottom-most stack frame?
func IsBottom(f Frame) bool {
    switch fr := f.(type) {
        case *frame:
            return fr == &frameBottom
    }
    panic("Unexpected frame implementation")
}
