package golog

// Tests for foreign predicates

import "testing"
import . "github.com/mndrix/golog/term"

func TestDeterministic(t *testing.T) {
	x := 0
	m := NewMachine().RegisterForeign(map[string]ForeignPredicate{
		"furrener/0": func(m Machine, args []Term) ForeignReturn {
			x++
			return ForeignTrue()
		},
	})

	// run the foreign predicate once
	if !m.CanProve(`furrener.`) {
		t.Errorf("Can't prove furrener/0 the first time")
	}
	if x != 1 {
		t.Errorf("x has the wrong value: %d vs 1", x)
	}

	// make sure it works when called again
	if !m.CanProve(`furrener.`) {
		t.Errorf("Can't prove furrener/0 the second time")
	}
	if x != 2 {
		t.Errorf("x has the wrong value: %d vs 2", x)
	}
}
