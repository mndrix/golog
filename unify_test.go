package golog

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
