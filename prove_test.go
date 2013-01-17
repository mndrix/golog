package golog

import "testing"

func TestFacts (t *testing.T) {
    m := NewMachine().Consult(`
        father(michael).
        father(marc).

        mother(gail).

        parent(X) :-
            father(X).
        parent(X) :-
            mother(X).
    `)
    t.Logf("%s\n", m.String())

    // these should be provably true
    if !m.CanProve(`father(michael).`) {
        t.Errorf("Couldn't prove father(michael)")
    }
    if !m.CanProve(`father(marc).`) {
        t.Errorf("Couldn't prove father(marc)")
    }
    if !m.CanProve(`parent(michael).`) {
        t.Errorf("Couldn't prove parent(michael)")
    }
    if !m.CanProve(`parent(marc).`) {
        t.Errorf("Couldn't prove parent(marc)")
    }

    // these should not be provable
    if m.CanProve(`father(sue).`) {
        t.Errorf("Proved father(sue)")
    }
    if m.CanProve(`father(michael,marc).`) {
        t.Errorf("Proved father(michael, marc)")
    }
    if m.CanProve(`mother(michael).`) {
        t.Errorf("Proved mother(michael)")
    }
    if m.CanProve(`parent(sue).`) {
        t.Errorf("Proved parent(sue)")
    }

    // trivial predicate with multiple solutions
    solutions := m.ProveAll(`father(X).`)
    if len(solutions) != 2 {
        t.Errorf("Wrong number of solutions: %d vs 2", len(solutions))
    }
    if x := solutions[0].ByName_("X").String(); x != "michael" {
        t.Errorf("Wrong first solution: %s", x)
    }
    if x := solutions[1].ByName_("X").String(); x != "marc" {
        t.Errorf("Wrong second solution: %s", x)
    }

    // simple predicate with multiple solutions
    solutions = m.ProveAll(`parent(Name).`)
    if len(solutions) != 3 {
        t.Errorf("Wrong number of solutions: %d vs 2", len(solutions))
    }
    if x := solutions[0].ByName_("Name").String(); x != "michael" {
        t.Errorf("Wrong first solution: %s", x)
    }
    if x := solutions[1].ByName_("Name").String(); x != "marc" {
        t.Errorf("Wrong second solution: %s", x)
    }
    if x := solutions[2].ByName_("Name").String(); x != "gail" {
        t.Errorf("Wrong third solution: %s", x)
    }
}

/* not quite ready yet
func TestConjunction(t *testing.T) {
    m := NewMachine().Consult(`
        floor_wax(briwax).
        floor_wax(shimmer).
        floor_wax(minwax).

        dessert(shimmer).
        dessert(cake).
        dessert(pie).

        snl(Item) :-
            floor_wax(Item),
            dessert(Item).
    `)

    skits := m.ProveAll(`snl(X).`)
    if len(skits) != 1 {
        t.Errorf("Wrong number of solutions: %d vs 1", len(skits))
    }
    if x := skits[0].ByName_("X").String(); x != "shimmer" {
        t.Errorf("Wrong solution: %s vs shimmer", x)
    }
}
*/
