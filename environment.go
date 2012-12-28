package golog

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
