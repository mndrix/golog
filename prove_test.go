package golog

import "testing"

func TestFacts (t *testing.T) {
    db := NewDatabase().
            Asserta(NewTerm("father", NewTerm("michael"))).
            Asserta(NewTerm("father", NewTerm("marc")))
    t.Logf("%s\n", db.String())

    // these should be provably true
    if !IsTrue(db, NewTerm("father", NewTerm("michael"))) {
        t.Errorf("Couldn't prove father(michael)")
    }
    if !IsTrue(db, NewTerm("father", NewTerm("marc"))) {
        t.Errorf("Couldn't prove father(marc)")
    }

    // these should not be provable
    if IsTrue(db, NewTerm("father", NewTerm("sue"))) {
        t.Errorf("Proved father(sue)")
    }
    if IsTrue(db, NewTerm("mother", NewTerm("michael"))) {
        t.Errorf("Proved mother(michael)")
    }
}
