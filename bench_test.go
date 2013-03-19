package golog

import "testing"
import "github.com/mndrix/golog/read"

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
