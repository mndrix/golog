package golog

import "github.com/mndrix/golog/read"

import "testing"

func TestFacts (t *testing.T) {
    rt := read.Term_
    facts := read.TermAll_(`
        father(michael).
        father(marc).
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
}
