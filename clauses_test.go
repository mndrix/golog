package golog

import "testing"

import "github.com/mndrix/golog/read"

func TestClauses(t *testing.T) {
	rt := read.Term_ // convenience

	cs0 := newClauses()
	if cs0.count() != 0 {
		t.Errorf("Initial clauses are not empty")
	}

	// add some terms
	cs1 := cs0.
		cons(rt(`hi(two).`)).
		cons(rt(`hi(one).`)).
		snoc(rt(`hi(three).`)).
		snoc(rt(`hi(four).`))
	if n := cs1.count(); n != 4 {
		t.Errorf("Incorrect term count: %d vs 4", n)
	}
	expected := []string{
		`hi(one)`,
		`hi(two)`,
		`hi(three)`,
		`hi(four)`,
	}
	for i, got := range cs1.all() {
		if got.String() != expected[i] {
			t.Errorf("Clause %d wrong: %s vs %s", i, got, expected[i])
		}
	}
}
