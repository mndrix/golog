package golog

import "fmt"
import "github.com/mndrix/ps"

var NotBound = fmt.Errorf("Variable is not bound")
var NotVariable = fmt.Errorf("Term is not a variable")
var AlreadyBound = fmt.Errorf("Variable was already bound")
type Environment interface {
    // Bind returns a new Environment, like the old one, but with the variable
    // bound to its new value; error is AlreadyBound if the variable had a
    // value previously.
    Bind(*Variable, Term) (Environment, error)

    // Value returns the value of a bound variable; error is NotBound if
    // the variable is free or NotVariable if not a variable term
    Value(*Variable) (Term, error)
}
func NewEnvironment() Environment {
    var newEnv envMap
    newEnv.bindings = ps.NewMap()
    newEnv.aliases = ps.NewMap()
    return &newEnv
}

type envMap struct {
    bindings    *ps.Map     // v.Indicator() => Term
    aliases     *ps.Map     // v.Indicator() => *Variable
}
func (self *envMap) Bind(v *Variable, val Term) (Environment, error) {
    v = self.resolveAliases(v)
    _, ok := self.bindings.Lookup(v.Indicator())
    if ok {
        // binding already exists for this variable
        return self, AlreadyBound
    }

    // at this point, we know that v is a free variable

    // are we aliasing 'v' to another variable?
    newEnv := self.clone()
    if IsVariable(val) {
        val = self.resolveAliases(val.(*Variable))
        // lexicographically smaller name is canonical
        switch {
            case v.Indicator() == val.Indicator():  // already aliased
                return self, AlreadyBound
            case v.Indicator() < val.Indicator():
                newEnv.aliases = self.aliases.Set(val.Indicator(), v)
            default:
                newEnv.aliases = self.aliases.Set(v.Indicator(), val)
        }
        return newEnv, nil
    }

    // create a new environment with the binding in place
    newEnv.bindings = self.bindings.Set(v.Indicator(), val)

    // error slot in return is for attributed variables someday
    return newEnv, nil
}
func (self *envMap) Value(v *Variable) (Term, error) {
    v = self.resolveAliases(v)
    name := v.Indicator()
    value, ok := self.bindings.Lookup(name)
    if !ok {
        return nil, NotBound
    }
    return value.(Term), nil
}
func (self *envMap) clone() *envMap {
    var newEnv envMap
    newEnv.bindings = self.bindings
    newEnv.aliases = self.aliases
    return &newEnv
}
func (self *envMap) resolveAliases(original *Variable) *Variable {
    name := original.Indicator()
    alias, ok := self.aliases.Lookup(name)
    if !ok {
        // this variable is not aliased
        return original
    }

    return self.resolveAliases(alias.(*Variable))
}
