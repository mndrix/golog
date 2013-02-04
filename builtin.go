package golog

// All of Golog's builtin, foreign-implemented predicates
// are defined here.

import "fmt"
import "github.com/mndrix/golog/term"

// !/0
func BuiltinCut(m Machine, args []term.Term) (bool, Machine) {
    frame := m.Stack()
    frame = frame.CutChoicePoints()
    return true, m.SetStack(frame)
}

// listing/0
// This should be implemented in pure Prolog, but for debugging purposes,
// I'm doing it for now as a foreign predicate.
func BuiltinListing0(m Machine, args []term.Term) (bool, Machine) {
    fmt.Println(m.String())
    return true, nil
}
