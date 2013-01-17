package golog

import "github.com/mndrix/golog/read"

import "testing"

func TestFacts (t *testing.T) {
    rt := read.Term_
    facts := read.TermAll_(`
        father(michael).
        father(marc).
        mother(gail).
        parent(X) :-
            father(X).
        parent(X) :-
            mother(X).
    `)
    db := NewDatabase()
    for _, fact := range facts {
        db = db.Asserta(fact)
    }
    t.Logf("%s\n", db.String())

    // these should be provably true
    if !IsTrue(db, rt(`father(michael).`)) {
        t.Errorf("Couldn't prove father(michael)")
    }
    if !IsTrue(db, rt(`father(marc).`)) {
        t.Errorf("Couldn't prove father(marc)")
    }
    if !IsTrue(db, rt(`parent(michael).`)) {
        t.Errorf("Couldn't prove parent(michael)")
    }
    if !IsTrue(db, rt(`parent(marc).`)) {
        t.Errorf("Couldn't prove parent(marc)")
    }

    // these should not be provable
    if IsTrue(db, rt(`father(sue).`)) {
        t.Errorf("Proved father(sue)")
    }
    if IsTrue(db, rt(`father(michael,marc).`)) {
        t.Errorf("Proved father(michael, marc)")
    }
    if IsTrue(db, rt(`mother(michael).`)) {
        t.Errorf("Proved mother(michael)")
    }
    if IsTrue(db, rt(`parent(sue).`)) {
        t.Errorf("Proved parent(sue)")
    }

    // trivial predicate with multiple solutions
    solutions := ProveAll(db, rt(`father(X).`))
    if len(solutions) != 2 {
        t.Errorf("Wrong number of solutions: %d vs 2", len(solutions))
    }
    if x := solutions[0].ByName_("X").String(); x != "marc" {  // 1st by Asserta
        t.Errorf("Wrong first solution: %s", x)
    }
    if x := solutions[1].ByName_("X").String(); x != "michael" {  // 2nd by Asserta
        t.Errorf("Wrong second solution: %s", x)
    }

    // simple predicate with multiple solutions
    solutions = ProveAll(db, rt(`parent(X).`))
    if len(solutions) != 3 {
        t.Errorf("Wrong number of solutions: %d vs 2", len(solutions))
    }
    if x := solutions[0].ByName_("X").String(); x != "gail" {  // 1st by Asserta
        t.Errorf("Wrong first solution: %s", x)
    }
    if x := solutions[1].ByName_("X").String(); x != "marc" {  // 2nd by Asserta
        t.Errorf("Wrong second solution: %s", x)
    }
    if x := solutions[2].ByName_("X").String(); x != "michael" {  // 3rd by Asserta
        t.Errorf("Wrong third solution: %s", x)
    }
}
