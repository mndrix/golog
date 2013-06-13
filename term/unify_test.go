package term

import "testing"

// convenience for below
func unify(env Bindings, a, b Term) (Bindings, error) {
	return a.Unify(env, b)
}

func TestUnifyConstants(t *testing.T) {
	env := NewBindings()

	// atoms
	_, err := NewAtom("hi").Unify(env, NewAtom("hi"))
	if err != nil {
		t.Errorf("hi/0 and hi/0 don't unify")
	}
	_, err = NewAtom("n").Unify(env, NewCode('n'))
	if err == nil {
		t.Errorf("n and 0'n unified")
	}

	// shallow terms
	_, err = unify(env,
		NewCallable("hi", NewCallable("you")),
		NewCallable("hi", NewCallable("you")),
	)
	if err != nil {
		t.Errorf("hi(you) and hi(you) don't unify")
	}

	// atom and deeper term don't unify
	_, err = unify(env,
		NewCallable("foo"),
		NewCallable("bar", NewCallable("baz")),
	)
	if err == nil {
		t.Errorf("foo and bar(baz) should not unify")
	}

	// integers and floats
	_, err = unify(env, NewInt("1234"), NewInt("1234"))
	if err != nil {
		t.Errorf("1234 and 1234 don't unify")
	}
	_, err = unify(env, NewInt("1234"), NewInt("1235"))
	if err == nil {
		t.Errorf("1234 and 1235 unify")
	}
	_, err = unify(env, NewFloat("99.2"), NewFloat("99.2"))
	if err != nil {
		t.Errorf("99.2 and 99.2 don't unify")
	}
	_, err = unify(env, NewFloat("8.2"), NewFloat("8.1"))
	if err == nil {
		t.Errorf("8.2 and 8.1 unify")
	}
	_, err = unify(env, NewInt("6"), NewFloat("6.0"))
	if err == nil {
		t.Errorf("6 and 6.0 unify")
	}
	_, err = unify(env, NewFloat("5.0"), NewInt("5"))
	if err == nil {
		t.Errorf("5.0 and 5 unify")
	}
}

func nv(name string) *Variable {
	return NewVar(name)
}

func TestUnifyAtomWithUnboundVariable(t *testing.T) {
	env0 := NewBindings()

	env1, err := unify(env0,
		NewCallable("x"),
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
	env1, err := unify(NewBindings(),
		NewVar("X"),
		NewCallable("alpha", NewCallable("beta")),
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
	env0 := NewBindings()
	env1, err := unify(env0,
		NewCallable("etc", NewCallable("stuff")),
		NewCallable("etc", nv("A")),
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
	env0 := NewBindings()
	env1, err := unify(env0, NewVar("X"), NewVar("X"))
	maybePanic(err)

	if env0.Size() != 0 {
		t.Errorf("env0 has bindings")
	}
	if env1.Size() != 0 {
		t.Errorf("env1 has bindings")
	}
}

func TestUnifyVariableAliases(t *testing.T) {
	env0 := NewBindings()

	// make two variables aliases for each other
	env1, err := unify(env0, NewVar("X0"), NewVar("X1"))
	maybePanic(err)

	// unify one of the aliased variables with a term
	env2, err := unify(env1, NewCallable("hello"), NewVar("X0"))
	maybePanic(err)

	// does X0 have the right value?
	x0 := env2.Resolve_(nv("X0"))
	if x0.String() != "hello" {
		t.Errorf("X0 has the wrong value: %s", x0)
	}

	// does X1 have the right value?
	x1 := env2.Resolve_(nv("X1"))
	maybePanic(err)
	if x1.String() != "hello" {
		t.Errorf("X1 has the wrong value: %s", x1)
	}
}

// same as TestUnifyVariableAliases but with first unification order switched
func TestUnifyVariableAliases2(t *testing.T) {
	env0 := NewBindings()

	// make two variables aliases for each other
	env1, err := unify(env0, nv("X1"), nv("X0"))
	maybePanic(err)

	// unify one of the aliased variables with a term
	env2, err := unify(env1, NewCallable("hello"), nv("X0"))
	maybePanic(err)

	// does X0 have the right value?
	x0 := env2.Resolve_(nv("X0"))
	if x0.String() != "hello" {
		t.Errorf("X0 has the wrong value: %s", x0)
	}

	// does X1 have the right value?
	x1 := env2.Resolve_(nv("X1"))
	if x1.String() != "hello" {
		t.Errorf("X1 has the wrong value: %s", x1)
	}
}
