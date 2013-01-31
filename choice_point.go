package golog

import "github.com/mndrix/golog/term"

// ChoicePoint represents a Prolog choice point which is an
// alternative computation.  The simplest choice point is one
// representing a goal to prove.  It might also be a Golog machine
// running concurrently to investigate a choice point.  In that case,
// following the choice point, just waits for the concurrent machine to
// send itself down a channel, where we continue the computation.
type ChoicePoint interface {
    // Follow produces a new machine, based on an existing one, which
    // differs only in having begun to prove this choice point
    Follow(Machine) (Machine, error)
}

type simpleCP struct {
    clause  term.Term
}
func NewSimpleChoicePoint(t term.Term) ChoicePoint {
    if t.IsClause() {
        return &simpleCP{clause: t}
    }
    return &simpleCP{clause: term.NewTerm(":-", t, term.NewTerm("true"))}
}
func (cp *simpleCP) Follow(m Machine) (Machine, error) {
    // rename variables so recursive clauses work
    clause := term.RenameVariables(cp.clause)

    // the machine's current goal unify with our head?
    env := m.Bindings()
    env1, err := term.Unify(env, m.Goal(), clause.Head())
    if err == term.CantUnify {
        return nil, PleaseBackTrack
    }
    maybePanic(err)

    // yup, update the environment and top goal
    return m.PushGoal(clause.Body(), env1)
}
