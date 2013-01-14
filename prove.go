package golog

import . "github.com/mndrix/golog/term"

// IsTrue returns true if goal can be proven from facts and clauses
// in the database
func IsTrue(db Database, goal Term) bool {
    env := NewEnvironment()
    return isTrue(env, db, goal)
}

func isTrue(env Environment, db Database, goal Term) bool {
    candidates := db.Candidates(goal)
    for _, candidate := range candidates {
        if candidate.IsClause() {
            newEnv, err := Unify(env, goal, candidate.Head())
            if err != nil {  // this clause applies
                if isTrue(newEnv, db, candidate.Body()) {
                    return true
                }
            }
        } else {
            _, err := Unify(env, goal, candidate)
            if err == nil {
                return true  // we proved the goal
            }
        }
    }

    // apparently, we were unable to prove the goal
    return false
}
