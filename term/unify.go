package term

import "fmt"

var CantUnify error = fmt.Errorf("Can't unify the given terms")

// Unify unifies two terms in the presence of an environment.
// On succes, returns a new environment with additional variable
// bindings.  On failure, returns CantUnify error along with the
// original environment
func Unify(e Bindings, a, b Term) (Bindings, error) {
    // variables always unify with themselves
    if IsVariable(a) && IsVariable(b) {
        if a.Indicator() == b.Indicator() {
            return e, nil
        }
    }

    // resolve any previous bindings
    if IsVariable(a) {
        a = e.Resolve(a.(*Variable))
    }
    if IsVariable(b) {
        b = e.Resolve(b.(*Variable))
    }

    // bind unbound variables
    if IsVariable(a) {
        return e.Bind(a.(*Variable), b)
    }
    if IsVariable(b) {
        return e.Bind(b.(*Variable), a)
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
