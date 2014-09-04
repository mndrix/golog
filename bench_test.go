package golog

import "fmt"
import "strconv"
import "testing"
import "github.com/mndrix/golog/read"
import "github.com/mndrix/golog/term"

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

// unify two compounds terms with deep structure. unification succeeds
func BenchmarkUnifyDeep(b *testing.B) {
	x := read.Term_(`a(b(c(d(e(f(g(h(i(j))))))))).`)
	y := read.Term_(`a(b(c(d(e(f(g(h(i(X))))))))).`)

	env := term.NewBindings()
	for i := 0; i < b.N; i++ {
		_, _ = x.Unify(env, y)
	}
}

// unify two compounds terms with deep structure. unification fails
func BenchmarkUnifyDeepFail(b *testing.B) {
	x := read.Term_(`a(b(c(d(e(f(g(h(i(j))))))))).`)
	y := read.Term_(`a(b(c(d(e(f(g(h(i(x))))))))).`)

	env := term.NewBindings()
	for i := 0; i < b.N; i++ {
		_, _ = x.Unify(env, y)
	}
}

func BenchmarkUnificationHash(b *testing.B) {
	x := read.Term_(`a(b(c(d(e(f(g(h(i(j))))))))).`)
	for i := 0; i < b.N; i++ {
		_ = term.UnificationHash([]term.Term{x}, 64, true)
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

func BenchmarkDCGish(b *testing.B) {
	m := NewMachine().Consult(`
        name([alice   |X], X).
        name([bob     |X], X).
        name([charles |X], X).
        name([david   |X], X).
        name([eric    |X], X).
        name([francis |X], X).
        name([george  |X], X).
        name([harry   |X], X).
        name([ignatius|X], X).
        name([john    |X], X).
        name([katie   |X], X).
        name([larry   |X], X).
        name([michael |X], X).
        name([nancy   |X], X).
        name([oliver  |X], X).
    `)
	g := read.Term_(`name([george,the,third], Rest).`)

	for i := 0; i < b.N; i++ {
		_ = m.ProveAll(g)
	}
}

func BenchmarkRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = read.Term_(`reverse([1,2,3,4,5,6,7], Xs).`)
	}
}

// Low level benchmarks to test Go's implementation
func init() { // avoid import errors when low level benchmarks comment out
	_ = fmt.Sprintf("")
	_ = strconv.Itoa(1)
}

/*
func BenchmarkLowLevelCompareUint64(b *testing.B) {
	var nintendo uint64 = 282429536481
	var other uint64 = 387429489
	for i := 0; i < b.N; i++ {
		if nintendo == other {
			// do nothing
		}
	}
}

func BenchmarkLowLevelCompareString(b *testing.B) {
	nintendo := "nintendo"
	other := "other"
	for i := 0; i < b.N; i++ {
		if nintendo == other {
			// do nothing
		}
	}
}
func BenchmarkLowLevelBitwise(b *testing.B) {
	var nintendo uint64 = 282429536481
	var other uint64 = 387429489
	for i := 0; i < b.N; i++ {
		if nintendo&other == nintendo {
			// do nothing
		}
	}
}
func BenchmarkLowLevelFloatBinaryExponent(b *testing.B) {
	f := 3.1415
	for i := 0; i < b.N; i++ {
		_ = strconv.FormatFloat(f, 'b', 0, 64)
	}
}
func BenchmarkLowLevelFloatDecimalExponent(b *testing.B) {
	f := 3.1415
	for i := 0; i < b.N; i++ {
		_ = strconv.FormatFloat(f, 'e', 64, 64)
	}
}
func BenchmarkLowLevelIntDecimal(b *testing.B) {
	var x uint64 = 1967
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("%d", x)
	}
}
func BenchmarkLowLevelIntHex(b *testing.B) {
	var x uint64 = 1967
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("%x", x)
	}
}

// benchmarks to compare performance on interface-related code
type AnInterface interface {
	AMethod() int
}
type ImplementationOne int

func (*ImplementationOne) AMethod() int { return 1 }

type ImplementationTwo int

func (*ImplementationTwo) AMethod() int { return 2 }

func NotAMethod(x AnInterface) int {
	switch x.(type) {
	case *ImplementationOne:
		return 1
	case *ImplementationTwo:
		return 2
	}
	panic("impossible")
}

func NotAMethodManual(x AnInterface) int {
	kind := x.AMethod()
	switch kind {
	case 1:
		return 1
	case 2:
		return 2
	}
	panic("impossible")
}

// how expensive is it to call a method?
func BenchmarkInterfaceMethod(b *testing.B) {
	var x AnInterface
	num := 100
	x = (*ImplementationOne)(&num)

	for i := 0; i < b.N; i++ {
		_ = x.AMethod()
	}
}

// how expensive is it to call a function that acts like a method?
func BenchmarkInterfaceFunctionTypeSwitch(b *testing.B) {
	var x AnInterface
	num := 100
	x = (*ImplementationOne)(&num)

	for i := 0; i < b.N; i++ {
		_ = NotAMethod(x)
	}
}

// how expensive is it to call a function that acts like a method?
func BenchmarkInterfaceFunctionManualTypeSwitch(b *testing.B) {
	var x AnInterface
	num := 100
	x = (*ImplementationOne)(&num)

	for i := 0; i < b.N; i++ {
		_ = NotAMethodManual(x)
	}
}

// how expensive is it to inline a type switch that acts like a method?
func BenchmarkInterfaceInlineTypeSwitch(b *testing.B) {
	var x AnInterface
	num := 100
	x = (*ImplementationOne)(&num)

	for i := 0; i < b.N; i++ {
		var y int
		switch x.(type) {
		case *ImplementationOne:
			y = 1
		case *ImplementationTwo:
			y = 2
		}
		_ = y
	}
}

// how expensive is a manually-implemented type switch?
func BenchmarkInterfaceManualTypeSwitch(b *testing.B) {
	var x AnInterface
	num := 100
	x = (*ImplementationOne)(&num)

	for i := 0; i < b.N; i++ {
		var y int
		kind := x.AMethod()
		switch kind {
		case 1:
			y = 1
		case 2:
			y = 2
		}
		_ = y
	}
}

*/
