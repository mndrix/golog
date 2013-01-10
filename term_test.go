package golog

import . "regexp"
import "testing"

func TestAtom(t *testing.T) {
    atom := NewTerm("prolog")
    if atom.Arity() != 0 {
        t.Errorf("atom's arity wasn't 0 it was %d", atom.Arity())
    }

    if atom.Functor() != "prolog" {
        t.Errorf("atom's has the wrong functor: %s", atom.Functor())
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
    if v.Indicator() != "X" {
        t.Errorf("variable X has the wrong indicator")
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
    shallow := NewTerm("prolog", NewTerm("in"), NewTerm("go"))

    if shallow.Arity() != 2 {
        t.Errorf("wrong arity: %d", shallow.Arity())
    }

    if shallow.Functor() != "prolog" {
        t.Errorf("wrong functor: %s", shallow.Functor())
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
    x := NewTerm(":-", NewTerm("foo"), NewTerm("bar"))
    if x.String() != ":-(foo, bar)" {
        t.Errorf("Clause has wrong quoting: %s", x.String())
    }

    // functors with punctuation and letters need quoting
    x = NewTerm("/a", NewTerm("foo"), NewTerm("bar"))
    if x.String() != "'/a'(foo, bar)" {
        t.Errorf("Clause has wrong quoting: %s", x.String())
    }


    // initial capital letters must be quoted
    x = NewTerm("Caps")
    if x.String() != "'Caps'" {
        t.Errorf("Capitalized atom has wrong quoting: %s", x.String())
    }

    // all lowercase atoms don't need quotes
    x = NewTerm("lower")
    if x.String() != "lower" {
        t.Errorf("Atom shouldn't be quoted: %s", x.String())
    }

    // initial lowercase atoms don't need quotes
    x = NewTerm("lower_Then_Caps")
    if x.String() != "lower_Then_Caps" {
        t.Errorf("Mixed case atom shouldn't be quoted: %s", x.String())
    }
}
