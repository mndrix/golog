package golog

// All of Golog's builtin, foreign-implemented predicates
// are defined here.

import "fmt"
import "github.com/mndrix/golog/term"

// !/0
func BuiltinCut(m Machine, args []term.Term) (bool, Machine) {
    frame := m.Stack()
    frame = frame.CutChoicePoints()

    // push true/0 to handle conjunctions following the cut
    m1, err := m.SetStack(frame).PushGoal(term.NewTerm("true"), nil)
    maybePanic(err)
    return true, m1
}

// call/*
func BuiltinCall(m Machine, args []term.Term) (bool, Machine) {
    // which goal is being called?
    bodyTerm := args[0]
    if term.IsVariable(bodyTerm) {
        bindings := m.Bindings()
        bodyTerm = bindings.Resolve_(bodyTerm.(*term.Variable))
    }

    // build a new goal with extra arguments attached
    functor := bodyTerm.Functor()
    newArgs := make([]term.Term, 0)
    newArgs = append(newArgs, bodyTerm.Arguments()...)
    newArgs = append(newArgs, args[1:]...)
    goal := term.NewTerm(functor, newArgs...)

    // construct a machine that will prove this goal
    m1, err := m.PushGoal(goal, nil)
    maybePanic(err)
    return true, m1
}

// listing/0
// This should be implemented in pure Prolog, but for debugging purposes,
// I'm doing it for now as a foreign predicate.
func BuiltinListing0(m Machine, args []term.Term) (bool, Machine) {
    fmt.Println(m.String())
    return true, nil
}
