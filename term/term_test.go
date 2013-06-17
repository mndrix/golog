package term

import . "regexp"
import "testing"

func TestAtom(t *testing.T) {
	atom := NewAtom("prolog")
	if atom.Arity() != 0 {
		t.Errorf("atom's arity wasn't 0 it was %d", atom.Arity())
	}

	if atom.Name() != "prolog" {
		t.Errorf("atom's has the wrong functor: %s", atom.Name())
	}

	if atom.Indicator() != "prolog/0" {
		t.Errorf("atom's indicator is wrong: %s", atom.Indicator())
	}

	if atom.String() != "prolog" {
		t.Errorf("wrong string representationt: %s", atom.String())
	}
}

func TestVariable(t *testing.T) {
	v := NewVar("X")
	if n := v.String(); n != "X" {
		t.Errorf("variable X has the wrong name %s", n)
	}

	a0 := NewVar("_")
	ok, err := MatchString(`^_`, a0.Indicator())
	maybePanic(err)
	if !ok {
		t.Errorf("a0 has the wrong indicator")
	}
	t.Logf("a0: %s", a0)

	a1 := NewVar("_")
	ok, err = MatchString(`^_`, a1.Indicator())
	maybePanic(err)
	if !ok {
		t.Errorf("a1 has the wrong indicator")
	}
	t.Logf("a1: %s", a1)

	if a0.Indicator() == a1.Indicator() {
		t.Errorf("anonymous variables are accidentally sharing names")
	}
}

func TestShallowTerm(t *testing.T) {
	shallow := NewCallable("prolog", NewAtom("in"), NewAtom("go"))

	if shallow.Arity() != 2 {
		t.Errorf("wrong arity: %d", shallow.Arity())
	}

	if shallow.Name() != "prolog" {
		t.Errorf("wrong functor: %s", shallow.Name())
	}

	if shallow.Indicator() != "prolog/2" {
		t.Errorf("indicator is wrong: %s", shallow.Indicator())
	}

	if shallow.String() != "prolog(in, go)" {
		t.Errorf("wrong string representation: %s", shallow.String())
	}
}

func TestQuoting(t *testing.T) {
	// functors entirely out of punctuation don't need quotes
	x := NewCallable(":-", NewCallable("foo"), NewCallable("bar"))
	if x.String() != ":-(foo, bar)" {
		t.Errorf("Clause has wrong quoting: %s", x.String())
	}

	// functors with punctuation and letters need quoting
	x = NewCallable("/a", NewCallable("foo"), NewCallable("bar"))
	if x.String() != "'/a'(foo, bar)" {
		t.Errorf("Clause has wrong quoting: %s", x.String())
	}

	// initial capital letters must be quoted
	x = NewCallable("Caps")
	if x.String() != "'Caps'" {
		t.Errorf("Capitalized atom has wrong quoting: %s", x.String())
	}

	// all lowercase atoms don't need quotes
	x = NewCallable("lower")
	if x.String() != "lower" {
		t.Errorf("Atom shouldn't be quoted: %s", x.String())
	}

	// initial lowercase atoms don't need quotes
	x = NewCallable("lower_Then_Caps")
	if x.String() != "lower_Then_Caps" {
		t.Errorf("Mixed case atom shouldn't be quoted: %s", x.String())
	}

	// empty list atom doesn't need quoting, but cons does
	x = NewCallable("[]")
	if x.String() != "[]" {
		t.Errorf("empty list atom shouldn't be quoted: %s", x.String())
	}
	x = NewCallable(".")
	if x.String() != "'.'" {
		t.Errorf("cons must be quoted: %s", x.String())
	}

	// cut doesn't need quoting
	x = NewCallable("!")
	if x.String() != "!" {
		t.Errorf("cut shouldn't be quoted: %s", x.String())
	}
}

func TestInteger(t *testing.T) {
	tests := make(map[string]Number)
	tests[`123`] = NewInt64(123)
	tests[`0xf`] = NewInt64(15)
	tests[`0o10`] = NewInt64(8)
	tests[`0b10`] = NewInt64(2)
	tests[`0' `] = NewInt64(32)
	tests[`0'\s`] = NewInt64(32) // SWI-Prolog extension
	tests[`0',`] = NewInt64(44)
	tests["0'\\x2218\\"] = NewInt64(0x2218)
	tests["0'\\21030\\"] = NewInt64(0x2218)

	for text, expected := range tests {
		x := NewInt(text)
		if NumberCmp(x, expected) != 0 {
			t.Errorf("Integer `%s` parsed as `%s` wanted `%s`", text, x, expected)
		}
	}

	large := NewInt(`989050597012214992552592926549`)
	if large.String() != `989050597012214992552592926549` {
		t.Errorf("Can't handle large integers")
	}
}

func TestFloat(t *testing.T) {
	tests := make(map[string]float64)
	tests[`3.14159`] = 3.14159
	tests[`2.0e2`] = 200.0
	tests[`2.5e-2`] = 0.025
	tests[`0.9E4`] = 9000.0

	for text, expected := range tests {
		x := NewFloat(text)
		if x.Float64() != expected {
			t.Errorf("Float `%s` parsed as `%f` wanted `%f`", text, x.Float64(), expected)
		}
	}
}
