package golog

// All of Golog's builtin, foreign-implemented predicates
// are defined here.

import "fmt"
import "github.com/mndrix/golog/term"

// !/0
func BuiltinCut(m Machine, args []term.Term) (Machine, bool) {
    // if were anything to cut, !/0 would have already been
    // replaced with '$cut_to/1'  Since this goal wasn't there
    // must be nothing cut, so treat it as an alias for "true/0"
    return m, true
}

// $cut_to/1
func BuiltinCutTo(m Machine, args []term.Term) (Machine, bool) {
    barrierId := args[0].(*term.Integer).Value().Int64()
    return m.CutTo(barrierId), true
}

// ,/2
func BuiltinComma(m Machine, args []term.Term) (Machine, bool) {
    return m.PushConj(args[1]).PushConj(args[0]), true
}

// ->/2
func BuiltinIfThen(m Machine, args []term.Term) (Machine, bool) {
    cond := args[0]
    then := args[1]

    cut := term.NewTerm("!")
    goal := term.NewTerm(",", cond, term.NewTerm(",", cut, then))
    m1 := m.PushCutBarrier().PushConj(goal)
    return m1, true
}

// ;/2
func BuiltinSemicolon(m Machine, args []term.Term) (Machine, bool) {
    cp := NewSimpleChoicePoint(m, args[1])
    m = m.PushDisj(cp)

    return m.PushConj(args[0]), true
}

// =/2
func BuiltinUnify(m Machine, args []term.Term) (Machine, bool) {
    bindings, err := term.Unify( m.Bindings(), args[0], args[1])
    if err == term.CantUnify {
        return nil, false
    }

    return m.SetBindings(bindings), true
}

// call/*
func BuiltinCall(m Machine, args []term.Term) (Machine, bool) {
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

    // construct a machine that will prove this goal next
    return m.PushCutBarrier().PushConj(goal), true
}

// fail/0
func BuiltinFail(m Machine, args []term.Term) (Machine, bool) {
    return nil, false
}

// listing/0
// This should be implemented in pure Prolog, but for debugging purposes,
// I'm doing it for now as a foreign predicate.
func BuiltinListing0(m Machine, args []term.Term) (Machine, bool) {
    fmt.Println(m.String())
    return m, true
}
