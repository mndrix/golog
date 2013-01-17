package golog

import . "github.com/mndrix/golog/term"

import "github.com/mndrix/golog/read"

import "fmt"

type Machine interface {
    CanProve(interface{}) bool
    Consult(interface{}) Machine
    ProveAll(interface{}) []Bindings
    String() string
}

type machine struct {
    db      Database
}

func NewMachine() Machine {
    var m machine
    m.db = NewDatabase()
    return &m
}

func (m *machine) clone() *machine {
    var m1 machine
    m1.db = m.db
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

func (m *machine) ProveAll(goal interface{}) []Bindings {
    g := m.toGoal(goal)
    env := NewBindings()
    return m.proveAll(env, g)
}

func (m *machine) proveAll(env Bindings, goal Term) []Bindings {
    solutions := make([]Bindings, 0)
    candidates := m.db.Candidates(goal)
    for _, candidate := range candidates {
        if candidate.IsClause() {
            newEnv, err := Unify(env, goal, candidate.Head())
            if err == nil {  // this clause applies
                subSolutions := m.proveAll(newEnv, candidate.Body())
                solutions = append(solutions, subSolutions...)
            }
        } else {
            newEnv, err := Unify(env, goal, candidate)
            if err == nil {
                solutions = append(solutions, newEnv)  // we proved the goal
            }
        }
    }

    return solutions
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

func (m *machine) readTerm(src interface{}) Term {
    return read.Term_(src)
}
