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
