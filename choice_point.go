package golog

import "fmt"
import "github.com/mndrix/golog/term"
import . "github.com/mndrix/golog/util"

// ChoicePoint represents a place in the execution where we had a
// choice between multiple alternatives.  Following a choice point
// is like making one of those choices.
//
// The simplest choice point simply tries to prove some goal; usually
// the body of a clause.  We can imagine more complex choice points.  Perhaps
// when the choice point was created, we spawned a goroutine to investigate
// that portion of the execution tree.  If that "clone" finds something
// interesting, it can send itself down a channel.  In this case, the
// ChoicePoint's Follow() method would just return that speculative, cloned
// machine.  In effect it's saying, "If you want to pursue this execution
// path, don't bother with the computation.  I've already done it"
//
// One can imagine a ChoicePoint implementation which clones the Golog
// machine onto a separate server in a cluster.  Following that choice point
// waits for the other server to finish evaluating its execution branch.
type ChoicePoint interface {
	// Follow produces a new machine which sets out to explore a
	// choice point.
	Follow() (Machine, error)
}

// a choice point which follows Head :- Body clauses
type headbodyCP struct {
	machine Machine
	goal    term.Term
	clause  term.Term
}

// A head-body choice point is one which, when followed, unifies a
// goal g with the head of a term t.  If unification fails, the choice point
// fails.  If unification succeeds, the machine tries to prove t's body.
func NewHeadBodyChoicePoint(m Machine, g, t term.Term) ChoicePoint {
	return &headbodyCP{machine: m, goal: g, clause: t}
}
func (cp *headbodyCP) Follow() (Machine, error) {
	// rename variables so recursive clauses work
	clause := term.RenameVariables(cp.clause)

	// does the machine's current goal unify with our head?
	head := clause
	if clause.Arity() == 2 && clause.Functor() == ":-" {
		head = clause.Head()
	}
	env, err := cp.goal.Unify(cp.machine.Bindings(), head)
	if err == term.CantUnify {
		return nil, err
	}
	MaybePanic(err)

	// yup, update the environment and top goal
	if clause.Arity() == 2 && clause.Functor() == ":-" {
		return cp.machine.SetBindings(env).PushConj(clause.Body()), nil
	}
	return cp.machine.SetBindings(env), nil // don't need to push "true"
}
func (cp *headbodyCP) String() string {
	return fmt.Sprintf("prove goal `%s` against rule `%s`", cp.goal, cp.clause)
}

// a choice point that just pushes a term onto conjunctions
type simpleCP struct {
	machine Machine
	goal    term.Term
}

// Following a simple choice point makes the machine start proving
// goal g.  If g is true/0, this choice point can be used to revert
// back to a previous version of a machine.  It can be useful for
// building certain control constructs.
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
var barrierID int64 = 0 // thread unsafe counter variable. fix when needed
type barrierCP struct {
	machine Machine
	id      int64
}

// NewCutBarrier creates a special choice point which acts as a sentinel
// value in the Golog machine's disjunction stack.  Attempting to follow
// a cut barrier choice point panics.
func NewCutBarrier(m Machine) ChoicePoint {
	barrierID++
	return &barrierCP{machine: m, id: barrierID}
}

var CutBarrierFails error = fmt.Errorf("Cut barriers never succeed")

func (cp *barrierCP) Follow() (Machine, error) {
	return nil, CutBarrierFails
}
func (cp *barrierCP) String() string {
	return fmt.Sprintf("cut barrier %d", cp.id)
}

// If cp is a cut barrier choice point, BarrierId returns an identifier
// unique to this cut barrier and true.  If cp is not a cut barrier,
// the second return value is false.  BarrierId is mostly useful for
// those hacking on the interpreter or doing strange control constructs
// with foreign predicates.  You probably don't need this.  Incidentally,
// !/0 is implemented in terms of this.
func BarrierId(cp ChoicePoint) (int64, bool) {
	switch b := cp.(type) {
	case *barrierCP:
		return b.id, true
	}
	return -1, false
}
