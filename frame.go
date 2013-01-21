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

    // NewChild creates a child stack frame similar to this one with
    // this one as the child's parent
    NewChild(Term, Bindings, *ps.List, *ps.List) Frame

    // Parent returns this stack frame's parent stack frame
    Parent() Frame

    // TakeChoicePoint returns the next choice point in line
    // (nil if there isn't one) and a new frame with the remaining
    // choice points intact
    TakeChoicePoint() (ChoicePoint, Frame)
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
    f1.conjs = f1.conjs
    return &f1
}

func (parent *frame) NewChild(goal Term, env Bindings, conjs, disjs *ps.List) Frame {
    child := parent.clone()
    child.parent = parent
    child.goal = goal
    if env != nil {
        child.env = env
    }
    if conjs != nil {
        child.conjs = conjs
    }
    if disjs != nil {
        child.disjs = disjs
    }

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

// is this the bottom-most stack frame?
func IsBottom(f Frame) bool {
    switch fr := f.(type) {
        case *frame:
            return fr == &frameBottom
    }
    panic("Unexpected frame implementation")
}
