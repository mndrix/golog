package golog

import "fmt"
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
    Follow() (Machine, error)
}

// a choice point which follows Head :- Body clauses
type headbodyCP struct {
    machine     Machine
    goal        term.Term
    clause      term.Term
}
func NewHeadBodyChoicePoint(m Machine, g, t term.Term) ChoicePoint {
    return &headbodyCP{machine: m, goal: g, clause: t}
}
func (cp *headbodyCP) Follow() (Machine, error) {
    // rename variables so recursive clauses work
    clause := term.RenameVariables(cp.clause)

    // does the machine's current goal unify with our head?
    head := clause
    if clause.Indicator() == ":-/2" {
        head = clause.Head()
    }
    env, err := term.Unify(cp.machine.Bindings(), cp.goal, head)
    if err == term.CantUnify {
        return nil, err
    }
    maybePanic(err)

    // yup, update the environment and top goal
    if clause.Indicator() == ":-/2" {
        return cp.machine.SetBindings(env).PushConj(clause.Body()), nil
    }
    return cp.machine.SetBindings(env), nil  // don't need to push "true"
}
func (cp *headbodyCP) String() string {
    return fmt.Sprintf("prove goal `%s` against rule `%s`", cp.goal, cp.clause)
}


// a choice point that just pushes a term onto conjunctions
type simpleCP struct {
    machine     Machine
    goal        term.Term
}
func NewSimpleChoicePoint(m Machine, g term.Term) ChoicePoint {
    return &simpleCP{machine: m, goal: g}
}
func (cp *simpleCP) Follow() (Machine, error) {
    return cp.machine.PushConj(cp.goal), nil
}
func (cp *simpleCP) String() string {
    return fmt.Sprintf("push conj %s", cp.goal)
}


// a noop choice point that represents a cut barrier
var barrierID int64 = 0  // thread unsafe counter variable. fix when needed
type barrierCP struct {
    machine     Machine
    id          int64
}
func NewCutBarrier(m Machine) ChoicePoint {
    barrierID++
    return &barrierCP{machine: m, id: barrierID}
}
func (cp *barrierCP) Follow() (Machine, error) {
    return nil, fmt.Errorf("Cut barriers never suceed")
}
func (cp *barrierCP) String() string {
    return fmt.Sprintf("cut barrier %d", cp.id)
}
func BarrierId(cp ChoicePoint) (int64, bool) {
    switch b := cp.(type) {
        case *barrierCP:
            return b.id, true
    }
    return -1, false
}
