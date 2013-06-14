package term

import "fmt"
import "math/big"

// Rational is a specialized, internal representation of floats.
// The goal is to facilitate more accurate numeric computations than
// floats allow.  This isn't always possible, but it's a helpful optimization
// in many practical circumstances.
type Rational big.Rat

// NewInt parses an integer's string representation to create a new
// integer value. Panics if the string's is not a valid integer
func NewRational(text string) (*Rational, bool) {
	r, ok := new(big.Rat).SetString(text)
	return (*Rational)(r), ok
}

func (self *Rational) Value() *big.Rat {
	return (*big.Rat)(self)
}

func (self *Rational) String() string {
	val := self.Value()
	if val.IsInt() {
		return val.RatString()
	}
	f, _ := val.Float64()
	return fmt.Sprintf("%g", f)
}

func (self *Rational) Type() int {
	val := self.Value()
	if val.IsInt() {
		return IntegerType
	}
	return FloatType
}

func (self *Rational) Indicator() string {
	return self.String()
}

func (a *Rational) Unify(e Bindings, b Term) (Bindings, error) {
	if IsVariable(b) {
		return b.Unify(e, a)
	}
	if IsRational(b) {
		if a.Value().Cmp(b.(*Rational).Value()) == 0 {
			return e, nil
		}
		return e, CantUnify
	}

	x := a.Value()
	switch b.Type() {
	case IntegerType:
		y := b.(*Integer)
		if !x.IsInt() {
			return e, CantUnify
		}
		if x.Num().Cmp(y.Value()) == 0 {
			return e, nil
		}
	case FloatType:
		f, _ := x.Float64()
		if f == b.(*Float).Value() {
			return e, nil
		}
	}

	return e, CantUnify
}

func (self *Rational) ReplaceVariables(env Bindings) Term {
	return self
}

// implement Number interface
func (self *Rational) Float64() float64 {
	f, _ := self.Value().Float64()
	return f
}
