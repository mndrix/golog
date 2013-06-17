package term

import "fmt"
import "math/big"
import . "github.com/mndrix/golog/util"

// Number represents either an integer or a floating point number.  This
// interface is convenient when working with arithmetic
type Number interface {
	Term

	// Float64 gives a floating point representation of this number.  The
	// representation inherits all weaknesses of the float64 data type.
	Float64() float64

	// LosslessInt returns a big.Int representation of this number, if
	// possible.  If int64 can't represent the number perfectly, returns
	// false.
	LosslessInt() (*big.Int, bool)

	// LosslessRat returns a big.Rational representation of this number,
	// if possible.  If a big.Rational can't represent the number perfectly,
	// returns false.  Floats never convert to rationals.
	LosslessRat() (*big.Rat, bool)
}

// Returns true if term t is an integer
func IsInteger(t Term) bool {
	return t.Type() == IntegerType
}

// Returns true if term t is an floating point number
func IsFloat(t Term) bool {
	return t.Type() == FloatType
}

// Returns true if term t is a rational number
func IsRational(t Term) bool {
	switch t.(type) {
	case *Rational:
		return true
	default:
		return false
	}
}

// Returns true if term is a number (integer, float, etc)
func IsNumber(t Term) bool {
	return IsInteger(t) || IsFloat(t) || IsRational(t)
}

// Evaluate an arithmetic expression to produce a number.  This is
// conceptually similar to Prolog: X is Expression.  Returns false if
// the expression cannot be evaluated (unbound variables, in)
func ArithmeticEval(t0 Term) (Number, error) {
	Debugf("arith eval: %s\n", t0)

	// number terms require no additional evaluation
	if IsNumber(t0) {
		return t0.(Number), nil
	}
	var t Callable
	if IsCallable(t0) {
		t = t0.(Callable)
	} else {
		msg := fmt.Sprintf("Unexpected arithmetic term: %s", t0)
		panic(msg)
	}

	// evaluate arithmetic expressions
	if t.Arity() == 2 {
		// evaluate arguments recursively
		args := t.Arguments()
		a, b, err := ArithmeticEval2(args[0], args[1])
		if err != nil {
			return nil, err
		}

		switch t.Name() {
		case "*": // multiplication
			return ArithmeticMultiply(a, b)
		case "+": // addition
			return ArithmeticAdd(a, b)
		case "-": // subtraction
			return ArithmeticMinus(a, b)
		case "/": // division
			return ArithmeticDivide(a, b)
		}

	}

	// this term doesn't look like an expression
	msg := fmt.Errorf("Not an expression: %s", t)
	return nil, msg
}

func ArithmeticEval2(first, second Term) (Number, Number, error) {
	// recursively evaluate each argument
	a, err := ArithmeticEval(first)
	if err != nil {
		return nil, nil, err
	}
	b, err := ArithmeticEval(second)
	if err != nil {
		return nil, nil, err
	}
	return a, b, nil
}

// Add two Golog numbers returning the result as a new Golog number
func ArithmeticAdd(a, b Number) (Number, error) {

	// as integers?
	if xi, ok := a.LosslessInt(); ok {
		if yi, ok := b.LosslessInt(); ok {
			r := new(big.Int).Add(xi, yi)
			return NewBigInt(r), nil
		}
	}

	// as rationals?
	if xr, ok := a.LosslessRat(); ok {
		if yr, ok := b.LosslessRat(); ok {
			r := new(big.Rat).Add(xr, yr)
			return NewBigRat(r), nil
		}
	}

	// as floats?
	r := a.Float64() + b.Float64()
	return NewFloat64(r), nil
}

// Divide two Golog numbers returning the result as a new Golog number.
// The return value uses the most precise internal type possible.
func ArithmeticDivide(a, b Number) (Number, error) {

	// as integers?
	if xi, ok := a.LosslessInt(); ok {
		if yi, ok := b.LosslessInt(); ok {
			r := new(big.Rat).SetFrac(xi, yi)
			return NewBigRat(r), nil
		}
	}

	// as rationals?
	if xr, ok := a.LosslessRat(); ok {
		if yr, ok := b.LosslessRat(); ok {
			r := new(big.Rat).Quo(xr, yr)
			return NewBigRat(r), nil
		}
	}

	// as floats?
	r := a.Float64() / b.Float64()
	return NewFloat64(r), nil
}

// Subtract two Golog numbers returning the result as a new Golog number
func ArithmeticMinus(a, b Number) (Number, error) {

	// as integers?
	if xi, ok := a.LosslessInt(); ok {
		if yi, ok := b.LosslessInt(); ok {
			r := new(big.Int).Sub(xi, yi)
			return NewBigInt(r), nil
		}
	}

	// as rationals?
	if xr, ok := a.LosslessRat(); ok {
		if yr, ok := b.LosslessRat(); ok {
			r := new(big.Rat).Sub(xr, yr)
			return NewBigRat(r), nil
		}
	}

	// as floats?
	r := a.Float64() - b.Float64()
	return NewFloat64(r), nil
}

// Multiply two Golog numbers returning the result as a new Golog number
func ArithmeticMultiply(a, b Number) (Number, error) {

	// as integers?
	if xi, ok := a.LosslessInt(); ok {
		if yi, ok := b.LosslessInt(); ok {
			r := new(big.Int).Mul(xi, yi)
			return NewBigInt(r), nil
		}
	}

	// as rationals?
	if xr, ok := a.LosslessRat(); ok {
		if yr, ok := b.LosslessRat(); ok {
			r := new(big.Rat).Mul(xr, yr)
			return NewBigRat(r), nil
		}
	}

	// as floats?
	r := a.Float64() * b.Float64()
	return NewFloat64(r), nil
}

// Compare two Golog numbers.  Returns
//
//    -1 if a <  b
//     0 if a == b
//    +1 if a > b
func NumberCmp(a, b Number) int {
	// compare as integers?
	if xi, ok := a.LosslessInt(); ok {
		if yi, ok := b.LosslessInt(); ok {
			return xi.Cmp(yi)
		}
	}

	// compare as rationals?
	if xr, ok := a.LosslessRat(); ok {
		if yr, ok := b.LosslessRat(); ok {
			return xr.Cmp(yr)
		}
	}

	// compare as floats?
	diff := a.Float64() - b.Float64()
	if diff < 0 {
		return -1
	}
	if diff > 0 {
		return 1
	}
	return 0
}
