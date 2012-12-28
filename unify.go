package golog

import "fmt"

var CantUnify error = fmt.Errorf("Can't unify the given terms")
type Environment interface {
    // Exists returns true if a variable of the given name exists
    Exists(string) bool

    // Bind returns a new Environment, like the old one, but with the variable
    // bound to its new value; panics if the variable had a value previously
    Bind(string, Term) Environment

    // Value returns the value of a bound variable; panics if not bound
    Value(string) Term
}
func NewEnvironment() Environment {
    env := envMap(make(map[string]Term))
    return &env
}

// TODO make a real implementation using ps.Map
type envMap map[string]Term
func (self *envMap) Exists(name string) bool {
    return false
}
func (self *envMap) Bind(name string, val Term) Environment {
    return self
}
func (self *envMap) Value(name string) Term {
    return NewTerm("TODO")
}


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
