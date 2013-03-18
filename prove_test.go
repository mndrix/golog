package golog

import "testing"
import "github.com/mndrix/golog/read"

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

    // cut in the top level query
    solutions = m.ProveAll(`parent(Name), !.`)
    if len(solutions) != 1 {
        t.Errorf("Wrong number of solutions: %d vs 1", len(solutions))
    }
    if x := solutions[0].ByName_("Name").String(); x != "michael" {
        t.Errorf("Wrong first solution: %s", x)
    }
}

func TestConjunction(t *testing.T) {
    m := NewMachine().Consult(`
        floor_wax(briwax).
        floor_wax(shimmer).
        floor_wax(minwax).

        dessert(shimmer).
        dessert(cake).
        dessert(pie).

        verb(glimmer).
        verb(shimmer).

        snl(Item) :-
            floor_wax(Item),
            dessert(Item).

        three(Item) :-
            verb(Item),
            dessert(Item),
            floor_wax(Item).
    `)

    skits := m.ProveAll(`snl(X).`)
    if len(skits) != 1 {
        t.Errorf("Wrong number of solutions: %d vs 1", len(skits))
    }
    if x := skits[0].ByName_("X").String(); x != "shimmer" {
        t.Errorf("Wrong solution: %s vs shimmer", x)
    }

    skits = m.ProveAll(`three(W).`)
    if len(skits) != 1 {
        t.Errorf("Wrong number of solutions: %d vs 1", len(skits))
    }
    if x := skits[0].ByName_("W").String(); x != "shimmer" {
        t.Errorf("Wrong solution: %s vs shimmer", x)
    }
}

func TestCut(t *testing.T) {
    m := NewMachine().Consult(`
        single(foo) :-
            !.
        single(bar).

        twice(X) :-
            single(X).  % cut inside here doesn't cut twice/1
        twice(bar).
    `)

    proofs := m.ProveAll(`single(X).`)
    if len(proofs) != 1 {
        t.Errorf("Wrong number of solutions: %d vs 1", len(proofs))
    }
    if x := proofs[0].ByName_("X").String(); x != "foo" {
        t.Errorf("Wrong solution: %s vs foo", x)
    }

    proofs = m.ProveAll(`twice(X).`)
    if len(proofs) != 2 {
        t.Errorf("Wrong number of solutions: %d vs 2", len(proofs))
    }
    if x := proofs[0].ByName_("X").String(); x != "foo" {
        t.Errorf("Wrong solution: %s vs foo", x)
    }
    if x := proofs[1].ByName_("X").String(); x != "bar" {
        t.Errorf("Wrong solution: %s vs bar", x)
    }
}

func TestAppend(t *testing.T) {
    m := NewMachine().Consult(`
        append([], A, A).   % test same variable name as other clauses
        append([A|B], C, [A|D]) :-
            append(B, C, D).
    `)

    proofs := m.ProveAll(`append([a], [b], List).`)
    if len(proofs) != 1 {
        t.Errorf("Wrong number of answers: %d vs 1", len(proofs))
    }
    if x := proofs[0].ByName_("List").String(); x != "'.'(a, '.'(b, []))" {
        t.Errorf("Wrong solution: %s vs '.'(a, '.'(b, []))", x)
    }

    proofs = m.ProveAll(`append([a,b,c], [d,e], List).`)
    if len(proofs) != 1 {
        t.Errorf("Wrong number of answers: %d vs 1", len(proofs))
    }
    if x := proofs[0].ByName_("List").String(); x != "'.'(a, '.'(b, '.'(c, '.'(d, '.'(e, [])))))" {
        t.Errorf("Wrong solution: %s", x)
    }
}

func TestCall (t *testing.T) {
    m := NewMachine().Consult(`
        bug(spider).
        bug(fly).

        squash(Animal, Class) :-
            call(Class, Animal).
    `)

    proofs := m.ProveAll(`squash(It, bug).`)
    if len(proofs) != 2 {
        t.Errorf("Wrong number of answers: %d vs 2", len(proofs))
    }
    if x := proofs[0].ByName_("It").String(); x != "spider" {
        t.Errorf("Wrong solution: %s vs spider", x)
    }
    if x := proofs[1].ByName_("It").String(); x != "fly" {
        t.Errorf("Wrong solution: %s vs fly", x)
    }
}

func TestUnify (t *testing.T) {
    m := NewMachine().Consult(`
        thing(Z) :-
            Z = whatever.
        two(X, Y) :-
            X = a,
            Y = b.
    `)

    proofs := m.ProveAll(`thing(It).`)
    if len(proofs) != 1 {
        t.Errorf("Wrong number of answers: %d vs 1", len(proofs))
    }
    if x := proofs[0].ByName_("It").String(); x != "whatever" {
        t.Errorf("Wrong solution: %s vs whatever", x)
    }

    proofs = m.ProveAll(`two(First, Second).`)
    if len(proofs) != 1 {
        t.Errorf("Wrong number of answers: %d vs 1", len(proofs))
    }
    if x := proofs[0].ByName_("First").String(); x != "a" {
        t.Errorf("Wrong solution: %s vs a", x)
    }
    if x := proofs[0].ByName_("Second").String(); x != "b" {
        t.Errorf("Wrong solution: %s vs b", x)
    }

    proofs = m.ProveAll(`two(j, k).`)
    if len(proofs) != 0 {
        t.Errorf("Proved the impossible")
    }
}

func TestDisjunction (t *testing.T) {
    m := NewMachine().Consult(`
        insect(fly).
        arachnid(spider).
        squash(Critter) :-
            arachnid(Critter) ; insect(Critter).
    `)

    proofs := m.ProveAll(`squash(It).`)
    if len(proofs) != 2 {
        t.Errorf("Wrong number of answers: %d vs 2", len(proofs))
    }
    if x := proofs[0].ByName_("It").String(); x != "spider" {
        t.Errorf("Wrong solution: %s vs spider", x)
    }
    if x := proofs[1].ByName_("It").String(); x != "fly" {
        t.Errorf("Wrong solution: %s vs fly", x)
    }
}

func TestIfThenElse (t *testing.T) {
    m := NewMachine().Consult(`
        succeeds(yes).
        succeeds(yup).
        alpha(X) :- succeeds(yes) -> X = ok.
        beta(X) :- succeeds(no) -> X = ok.
    `)

    proofs := m.ProveAll(`alpha(Y).`)
    if len(proofs) != 1 {
        t.Errorf("Wrong number of answers: %d vs 1", len(proofs))
    }
    if x := proofs[0].ByName_("Y").String(); x != "ok" {
        t.Errorf("Wrong solution: %s vs ok", x)
    }

    proofs = m.ProveAll(`beta(Y).`)
    if len(proofs) != 0 {
        t.Errorf("Wrong number of answers: %d vs 0", len(proofs))
    }
}

func BenchmarkTrue(b *testing.B) {
    m := NewMachine()
    g := read.Term_(`true.`)
    for i := 0; i < b.N; i++ {
        _ = m.ProveAll(g)
    }
}

func BenchmarkAppend(b *testing.B) {
    m := NewMachine().Consult(`
        append([], A, A).   % test same variable name as other clauses
        append([A|B], C, [A|D]) :-
            append(B, C, D).
    `)
    g := read.Term_(`append([a,b,c], [d,e], List).`)

    for i := 0; i < b.N; i++ {
        _ = m.ProveAll(g)
    }
}

// test performance of a standard maplist implementation
func BenchmarkMaplist(b *testing.B) {
    m := NewMachine().Consult(`
        always_a(_, a).

        maplist(C, A, B) :-
            maplist_(A, B, C).

        maplist_([], [], _).
        maplist_([B|D], [C|E], A) :-
            call(A, B, C),
            maplist_(D, E, A).
    `)
    g := read.Term_(`maplist(always_a, [1,2,3,4,5], As).`)

    for i := 0; i < b.N; i++ {
        _ = m.ProveAll(g)
    }
}


// traditional, naive reverse benchmark
// The Art of Prolog by Sterling, etal says that reversing a 30 element
// list using this technique does 496 reductions.  From this we can
// calculate a rough measure of Golog's LIPS.
func BenchmarkNaiveReverse(b *testing.B) {
    m := NewMachine().Consult(`
        append([], A, A).
        append([A|B], C, [A|D]) :-
            append(B, C, D).

        reverse([],[]).
        reverse([X|Xs], Zs) :-
            reverse(Xs, Ys),
            append(Ys, [X], Zs).
    `)
    g := read.Term_(`reverse([1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30], As).`)

    for i := 0; i < b.N; i++ {
        _ = m.ProveAll(g)
    }
}

func BenchmarkRead(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _ = read.Term_(`reverse([1,2,3,4,5,6,7], Xs).`)
    }
}


// two small benchmarks for integer and string comparison
func BenchmarkCompareUint64(b *testing.B) {
    var nintendo uint64 = 282429536481
    var other uint64 = 387429489
    for i := 0; i < b.N; i++ {
        if nintendo == other {
            // do nothing
        }
    }
}
func BenchmarkCompareString(b *testing.B) {
    nintendo := "nintendo"
    other := "other"
    for i := 0; i < b.N; i++ {
        if nintendo == other {
            // do nothing
        }
    }
}
