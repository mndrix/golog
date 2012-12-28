package golog

import "fmt"

var CantUnify error = fmt.Errorf("Can't unify the given terms")

// Unify unifies two terms in the presence of an environment.
// On succes, returns a new environment with additional variable
// bindings.  On failure, returns CantUnify error along with the
// original environment
func Unify(e Environment, a, b Term) (Environment, error) {
    // functor and arity must match for unification to work
    arity := a.Arity()
    if arity != b.Arity() {
        return e, CantUnify
    }
    if a.Functor() != b.Functor() {
        return e, CantUnify
    }

    // try unifying each subterm
    var err error
    env := e
    aArgs := a.Arguments()
    bArgs := b.Arguments()
    for i:=0; i<arity; i++ {
        env, err = Unify(env, aArgs[i], bArgs[i])
        if err != nil {
            return e, err // return original environment along with error
        }
    }

    // unification succeeded
    return env, nil
}
