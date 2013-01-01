package golog

import "fmt"

var CantUnify error = fmt.Errorf("Can't unify the given terms")

// Unify unifies two terms in the presence of an environment.
// On succes, returns a new environment with additional variable
// bindings.  On failure, returns CantUnify error along with the
// original environment
func Unify(e Environment, a, b Term) (Environment, error) {
    // can variable 'a' unify trivially?
    env, err := UnifyFreeVariable(e, a, b)
    if err == nil {
        return env, nil
    }

    // can variable 'b' unify trivially?
    env, err = UnifyFreeVariable(e, b, a)
    if err == nil {
        return env, nil
    }

    // at this point, neither term is a variable so try harder

    // functor and arity must match for unification to work
    arity := a.Arity()
    if arity != b.Arity() {
        return e, CantUnify
    }
    if a.Functor() != b.Functor() {
        return e, CantUnify
    }

    // try unifying each subterm
    env = e
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

// UnifyFreeVariable tries to bind free variable 'v' to term 't', returning
// a new environment on success; otherwise, giving an error explaining why
func UnifyFreeVariable(e Environment, v, t Term) (Environment, error) {
    if !IsVariable(v) {
        return e, NotVariable
    }
    env, err := e.Bind(v.(*Variable), t)
    if err != nil {
        return e, err  // binding failed, return original environment
    }
    return env, nil
}
