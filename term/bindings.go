package term

import "fmt"
import "github.com/mndrix/ps"

// Returned by Bind() if the variable in question has already
// been bound to a value.
var AlreadyBound = fmt.Errorf("Variable was already bound")

// Returned by Value() and ByName() if the variable in question has
// no bindings yet.
var NotBound = fmt.Errorf("Variable is not bound")

type Bindings interface {
	// Bind returns a new Environment, like the old one, but with the variable
	// bound to its new value; error is AlreadyBound if the variable had a
	// value previously.
	Bind(*Variable, Term) (Bindings, error)

	// ByName is similar to Value() but it searches for a variable
	// binding by using that variable's name.  Names can be ambiguous so
	// use with caution.
	ByName(string) (Term, error)

	// ByName_ is like ByName() but panics on error.
	ByName_(string) Term

	// Resolve follows bindings recursively until a term is found for
	// which no binding exists.  If you want to know the value of a
	// variable, this is your best bet.
	Resolve(*Variable) (Term, error)

	// Resolve_ is like Resolve() but panics on error.
	Resolve_(*Variable) Term

	// Size returns the number of variable bindings in this environment.
	Size() int

	// Value returns the value of a bound variable; error is NotBound if
	// the variable is free.
	Value(*Variable) (Term, error)

	// WithNames returns a new bindings with human-readable names attached
	// for convenient lookup.  Panics if names have already been attached.
	WithNames(ps.Map) Bindings
}

// NewBindings returns a new, empty bindings value.
func NewBindings() Bindings {
	var newEnv envMap
	newEnv.bindings = ps.NewMap()
	newEnv.names = ps.NewMap()
	return &newEnv
}

type envMap struct {
	bindings ps.Map // v.Indicator() => Term
	names    ps.Map // v.Name => *Variable
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
	newEnv := *self
	return &newEnv
}

func (self *envMap) ByName(name string) (Term, error) {
	v, ok := self.names.Lookup(name)
	if !ok {
		return nil, NotBound
	}
	return self.Resolve(v.(*Variable))
}

func (self *envMap) ByName_(name string) Term {
	x, err := self.ByName(name)
	maybePanic(err)
	return x
}

func (self *envMap) String() string {
	return self.bindings.String()
}

func (self *envMap) WithNames(names ps.Map) Bindings {
	if !self.names.IsNil() {
		panic("Can't set names when names have already been set")
	}

	b := self.clone()
	b.names = names
	return b
}
