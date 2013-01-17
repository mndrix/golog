package golog

import . "github.com/mndrix/golog/term"

// IsTrue returns true if goal can be proven from facts and clauses
// in the database
func IsTrue(db Database, goal Term) bool {
    solutions := ProveAll(db, goal)
    return len(solutions) > 0
}

func ProveAll(db Database, goal Term) []Bindings {
    env := NewBindings()
    return proveAll(env, db, goal)
}

func proveAll(env Bindings, db Database, goal Term) []Bindings {
    solutions := make([]Bindings, 0)
    candidates := db.Candidates(goal)
    for _, candidate := range candidates {
        if candidate.IsClause() {
            newEnv, err := Unify(env, goal, candidate.Head())
            if err == nil {  // this clause applies
                subSolutions := proveAll(newEnv, db, candidate.Body())
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
