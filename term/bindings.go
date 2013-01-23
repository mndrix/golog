package term

import "fmt"
import "github.com/mndrix/ps"

var NotBound = fmt.Errorf("Variable is not bound")
var AlreadyBound = fmt.Errorf("Variable was already bound")
type Bindings interface {
    // Bind returns a new Environment, like the old one, but with the variable
    // bound to its new value; error is AlreadyBound if the variable had a
    // value previously.
    Bind(*Variable, Term) (Bindings, error)

    // ByName is similar to Value() but it searches for a variable
    // binding by using that variable's name
    ByName(string) (Term, error)

    // ByName_ is like ByName() but panics on error
    ByName_(string) Term

    // Resolve follows bindings recursively until a term is found for
    // which no binding exists
    Resolve(*Variable) (Term, error)

    // Resolve_ is like Resolve() but panics on error
    Resolve_(*Variable) Term

    // Size returns the number of variable bindings in this environment
    Size() int

    // Value returns the value of a bound variable; error is NotBound if
    // the variable is free
    Value(*Variable) (Term, error)
}
func NewBindings() Bindings {
    var newEnv envMap
    newEnv.bindings = ps.NewMap()
    return &newEnv
}

type envMap struct {
    bindings    ps.Map     // v.Indicator() => Term
}
func (self *envMap) Bind(v *Variable, val Term) (Bindings, error) {
    _, ok := self.bindings.Lookup(v.Indicator())
    if ok {
        // binding already exists for this variable
        return self, AlreadyBound
    }

    // at this point, we know that v is a free variable

    // create a new environment with the binding in place
    newEnv := self.clone()
    newEnv.bindings = self.bindings.Set(v.Indicator(), val)

    // error slot in return is for attributed variables someday
    return newEnv, nil
}
func (self *envMap) Resolve_(v *Variable) Term {
    r, err := self.Resolve(v)
    maybePanic(err)
    return r
}

func (self *envMap) Resolve(v *Variable) (Term, error) {
    for {
        t, err := self.Value(v)
        if err == NotBound {
            return v, nil
        }
        if err != nil {
            return nil, err
        }
        if IsVariable(t) {
            v = t.(*Variable)
        } else {
            return t.ReplaceVariables(self), nil
        }
    }
    panic("Shouldn't reach here")
}
func (self *envMap) Size() int {
    return self.bindings.Size()
}
func (self *envMap) Value(v *Variable) (Term, error) {
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
    return &newEnv
}

func (self *envMap) ByName(name string) (Term, error) {
    v := NewVar(name).(*Variable)
    return self.Resolve(v)
}

func (self *envMap) ByName_(name string) Term {
    x, err := self.ByName(name)
    maybePanic(err)
    return x
}
