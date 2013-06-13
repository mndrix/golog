package term

import . "fmt"
import . "regexp"
import . "github.com/mndrix/golog/util"

var anonCounter <-chan int64

func init() {
	// goroutine providing a counter for anonymous variables
	c := make(chan int64)
	var i int64 = 1000
	go func() {
		for {
			c <- i
			i++
		}
	}()
	anonCounter = c
}

// A Prolog logic variable.  See ISO §6.1.2(a)
type Variable struct {
	Name string
	id   int64 // uniquely identifiers this variable
}

// Creates a new logic variable with the given name.
func NewVar(name string) *Variable {
	// sanity check the variable name's syntax
	isCapitalized, err := MatchString(`^[A-Z_]`, name)
	maybePanic(err)
	if !isCapitalized {
		panic("Variable names must start with a capital letter or underscore")
	}

	// make sure anonymous variables are unique
	var i int64
	if name == "_" {
		i = <-anonCounter
	}
	return &Variable{
		Name: name,
		id:   i,
	}
}

// Id returns a unique identifier for this variable
func (self *Variable) Id() int64 {
	return self.id
}

func (self *Variable) Functor() string {
	panic("Variables have no Functor()")
}

func (self *Variable) Arity() int {
	panic("Variables have no Arity()")
}

func (self *Variable) Arguments() []Term {
	panic("Variables have no Arguments()")
}

func (self *Variable) String() string {
	if Debugging() && self.Name == "_" {
		return self.Indicator()
	}
	return self.Name
}

func (self *Variable) Type() int {
	return VariableType
}

func (self *Variable) Indicator() string {
	return Sprintf("_V%d", self.id)
}

func (a *Variable) Unify(e Bindings, b Term) (Bindings, error) {
	var aTerm, bTerm Term

	// a variable always unifies with itself
	if IsVariable(b) {
		if a.Indicator() == b.Indicator() {
			return e, nil
		}
		bTerm = e.Resolve_(b.(*Variable))
	} else {
		bTerm = b
	}

	// resolve any previous bindings
	aTerm = e.Resolve_(a)

	// bind unbound variables
	if IsVariable(aTerm) {
		return e.Bind(aTerm.(*Variable), b)
	}
	if IsVariable(bTerm) {
		return e.Bind(bTerm.(*Variable), a)
	}

	// otherwise, punt
	return aTerm.Unify(e, bTerm)
}

func (self *Variable) ReplaceVariables(env Bindings) Term {
	return env.Resolve_(self)
}

func (self *Variable) WithNewId() *Variable {
	return &Variable{
		Name: self.Name,
		id:   <-anonCounter,
	}
}
