package golog

import "testing"

func TestAsserta(t *testing.T) {
    db0 := NewDatabase()
    db1 := db0.Asserta(NewTerm("alpha"))
    db2 := db1.Asserta(NewTerm("beta"))
    db3 := db2.Asserta(NewTerm("gamma", NewTerm("greek to me")))

    // do we have the right number of clauses?
    if db0.ClauseCount() != 0 {
        t.Errorf("db0: wrong number of clauses: %d", db0.ClauseCount())
    }
    if db1.ClauseCount() != 1 {
        t.Errorf("db0: wrong number of clauses: %d", db0.ClauseCount())
    }
    if db2.ClauseCount() != 2 {
        t.Errorf("db0: wrong number of clauses: %d", db0.ClauseCount())
    }
    if db3.ClauseCount() != 3 {
        t.Errorf("db0: wrong number of clauses: %d", db0.ClauseCount())
    }
}
