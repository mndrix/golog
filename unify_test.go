package golog

import . "github.com/mndrix/golog/term"

import "testing"

func TestUnifyConstants(t *testing.T) {
    env := NewEnvironment()

    // atoms
    _, err := Unify(env, NewTerm("hi"), NewTerm("hi"))
    if err != nil {
        t.Errorf("hi/0 and hi/0 don't unify")
    }

    // shallow terms
    _, err = Unify( env,
        NewTerm("hi", NewTerm("you")),
        NewTerm("hi", NewTerm("you")),
    )
    if err != nil {
        t.Errorf("hi(you) and hi(you) don't unify")
    }

    // atom and deeper term don't unify
    _, err = Unify( env,
        NewTerm("foo"),
        NewTerm("bar", NewTerm("baz")),
    )
    if err == nil {
        t.Errorf("foo and bar(baz) should not unify")
    }
}

func nv(name string) *Variable {
    return NewVar(name).(*Variable)
}

func TestUnifyAtomWithUnboundVariable(t *testing.T) {
    env0 := NewEnvironment()

    env1, err := Unify( env0,
        NewTerm("x"),
        nv("X"),
    )
    if err != nil {
        t.Errorf("Couldn't unify `x=X`")
    }
    x0, err := env1.Value(nv("X"))
    if err != nil {
        t.Errorf("No binding produced for X")
    }
    if x0.String() != "x" {
        t.Errorf("X has the wrong value: %s", x0)
    }
}

func TestUnifyUnboundVariableWithStructure(t *testing.T) {
    env1, err := Unify( NewEnvironment(),
        NewVar("X"),
        NewTerm("alpha", NewTerm("beta")),
    )
    if err != nil {
        t.Errorf("Couldn't unify `X=alpha(beta)`")
    }
    x0, err := env1.Value(nv("X"))
    if err != nil {
        t.Errorf("No binding produced for X")
    }
    if x0.String() != "alpha(beta)" {
        t.Errorf("X has the wrong value: %s", x0)
    }
}

func TestUnifyNestedVariable(t *testing.T) {
    env0 := NewEnvironment()
    env1, err := Unify( env0,
        NewTerm("etc", NewTerm("stuff")),
        NewTerm("etc", nv("A")),
    )
    if err != nil {
        t.Errorf("Couldn't unify `etc(stuff)=etc(A)`")
    }
    x0, err := env1.Value(nv("A"))
    if err != nil {
        t.Errorf("No binding produced for A")
    }
    if x0.String() != "stuff" {
        t.Errorf("A has the wrong value: %s", x0)
    }


    // A shouldn't be bound in the original, empty environment
    _, err = env0.Value(nv("A"))
    if err != NotBound {
        t.Errorf("Unification changed the original environment")
    }
}

func TestUnifySameVariable(t *testing.T) {
    env0 := NewEnvironment()
    env1, err := Unify(env0, NewVar("X"), NewVar("X"))
    maybePanic(err)

    if env0.Size() != 0 {
        t.Errorf("env0 has bindings")
    }
    if env1.Size() != 0 {
        t.Errorf("env1 has bindings")
    }
}

func TestUnifyVariableAliases(t *testing.T) {
    env0 := NewEnvironment()

    // make two variables aliases for each other
    env1, err := Unify( env0, NewVar("X0"), NewVar("X1"))
    maybePanic(err)

    // unify one of the aliased variables with a term
    env2, err := Unify( env1, NewTerm("hello"), NewVar("X0"))
    maybePanic(err)

    // does X0 have the right value?
    x0 := env2.Resolve(nv("X0"))
    if x0.String() != "hello" {
        t.Errorf("X0 has the wrong value: %s", x0)
    }

    // does X1 have the right value?
    x1 := env2.Resolve(nv("X1"))
    maybePanic(err)
    if x1.String() != "hello" {
        t.Errorf("X1 has the wrong value: %s", x1)
    }
}


// same as TestUnifyVariableAliases but with first unification order switched
func TestUnifyVariableAliases2(t *testing.T) {
    env0 := NewEnvironment()

    // make two variables aliases for each other
    env1, err := Unify( env0, nv("X1"), nv("X0"))
    maybePanic(err)

    // unify one of the aliased variables with a term
    env2, err := Unify( env1, NewTerm("hello"), nv("X0"))
    maybePanic(err)

    // does X0 have the right value?
    x0 := env2.Resolve(nv("X0"))
    if x0.String() != "hello" {
        t.Errorf("X0 has the wrong value: %s", x0)
    }

    // does X1 have the right value?
    x1 := env2.Resolve(nv("X1"))
    if x1.String() != "hello" {
        t.Errorf("X1 has the wrong value: %s", x1)
    }
}
