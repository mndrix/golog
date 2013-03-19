package golog

import . "github.com/mndrix/golog/term"

import "github.com/mndrix/golog/read"

import "testing"

func TestAsserta(t *testing.T) {
    db0 := NewDatabase()
    db1 := db0.Asserta(NewAtom("alpha"))
    db2 := db1.Asserta(NewAtom("beta"))

    db3 := db2.Asserta(read.Term_(`foo(one,two) :- alpha.`))
    t.Logf(db3.String()) // helpful for debugging

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

    // is alpha/0 present where it should be?
    if cs := db1.Candidates_(NewAtom("alpha")); len(cs) != 1 {
        t.Errorf("db1: can't find alpha/0")
    }
    if cs := db2.Candidates_(NewAtom("alpha")); len(cs) != 1 {
        t.Errorf("db2: can't find alpha/0")
    }
    if cs := db3.Candidates_(NewAtom("alpha")); len(cs) != 1 {
        t.Errorf("db3: can't find alpha/0")
    }

    // is beta/0 present where it should be?
    if _, err := db1.Candidates(NewAtom("beta")); err == nil {
        t.Errorf("db1: shouldn't have found beta/0")
    }
    if cs := db2.Candidates_(NewAtom("beta")); len(cs) != 1 {
        t.Errorf("db2: can't find beta/0")
    }
    if cs := db3.Candidates_(NewAtom("beta")); len(cs) != 1 {
        t.Errorf("db3: can't find beta/0")
    }

    // is foo/2 present where it should be?
    term := read.Term_(`foo(one,two).`)
    if _, err := db1.Candidates(term); err == nil {
        t.Errorf("db1: shouldn't have found foo/2")
    }
    if _, err := db2.Candidates(term); err == nil {
        t.Errorf("db2: shouldn't have found foo/2")
    }
    if cs := db3.Candidates_(term); len(cs) != 1 {
        t.Errorf("db3: can't find foo/2")
    }
}
