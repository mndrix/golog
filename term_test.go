package golog

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
