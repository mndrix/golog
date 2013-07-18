package golog

// All of Golog's builtin, foreign-implemented predicates
// are defined here.

import "fmt"
import "math/big"
import "sort"
import "strings"
import "github.com/mndrix/golog/term"
import . "github.com/mndrix/golog/util"

// !/0
func BuiltinCut(m Machine, args []term.Term) ForeignReturn {
	// if were anything to cut, !/0 would have already been
	// replaced with '$cut_to/1'  Since this goal wasn't there
	// must be nothing cut, so treat it as an alias for "true/0"
	return ForeignTrue()
}

// $cut_to/1
//
// An internal system predicate which might be removed at any time
// in the future.  It cuts all disjunctions on top of a specific cut
// barrier.
func BuiltinCutTo(m Machine, args []term.Term) ForeignReturn {
	barrierId := args[0].(*term.Integer).Value().Int64()
	return m.CutTo(barrierId)
}

// ,/2
func BuiltinComma(m Machine, args []term.Term) ForeignReturn {
	return m.PushConj(args[1].(term.Callable)).PushConj(args[0].(term.Callable))
}

// ground/1
func BuiltinGround(m Machine, args []term.Term) ForeignReturn {
	switch args[0].Type() {
	case term.VariableType:
		return ForeignFail()
	case term.AtomType,
		term.IntegerType,
		term.FloatType,
		term.ErrorType:
		return ForeignTrue()
	case term.CompoundType:
		// recursively evaluate compound term's arguments
		x := args[0].(*term.Compound)
		for _, arg := range x.Arguments() {
			f := BuiltinGround(m, []term.Term{arg})
			switch f.(type) {
			case *foreignFail:
				return ForeignFail()
			}
		}
		return ForeignTrue()
	}
	msg := fmt.Sprintf("Unexpected term type: %#v", args[0])
	panic(msg)
}

// is/2
func BuiltinIs(m Machine, args []term.Term) ForeignReturn {
	value := args[0]
	expression := args[1]
	num, err := term.ArithmeticEval(expression)
	MaybePanic(err)
	return ForeignUnify(value, num)
}

// ->/2
func BuiltinIfThen(m Machine, args []term.Term) ForeignReturn {
	cond := args[0]
	then := args[1]

	// CUT_BARRIER, (cond, !, then)
	cut := term.NewCallable("!")
	goal := term.NewCallable(",", cond, term.NewCallable(",", cut, then))
	return m.DemandCutBarrier().PushConj(goal)
}

// ;/2
//
// Implements disjunction and if-then-else.
func BuiltinSemicolon(m Machine, args []term.Term) ForeignReturn {
	if term.IsCompound(args[0]) {
		ct := args[0].(*term.Compound)
		if ct.Arity() == 2 && ct.Name() == "->" { // §7.8.8
			return ifThenElse(m, args)
		}
	}

	cp := NewSimpleChoicePoint(m, args[1])
	return m.PushDisj(cp).PushConj(args[0].(term.Callable))
}
func ifThenElse(m Machine, args []term.Term) ForeignReturn {
	semicolon := args[0].(*term.Compound)
	cond := semicolon.Arguments()[0]
	then := semicolon.Arguments()[1]
	els := args[1]

	// CUT_BARRIER, (call(cond), !, then; else)
	cut := term.NewCallable("!")
	cond = term.NewCallable("call", cond)
	goal := term.NewCallable(",", cond, term.NewCallable(",", cut, then))
	goal = term.NewCallable(";", goal, els)
	return m.DemandCutBarrier().PushConj(goal)
}

// =/2
func BuiltinUnify(m Machine, args []term.Term) ForeignReturn {
	return ForeignUnify(args[0], args[1])
}

// =:=/2
func BuiltinNumericEquals(m Machine, args []term.Term) ForeignReturn {
	// evaluate each arithmetic argument
	a, err := term.ArithmeticEval(args[0])
	MaybePanic(err)
	b, err := term.ArithmeticEval(args[1])
	MaybePanic(err)

	// perform the actual comparison
	if term.NumberCmp(a, b) == 0 {
		return ForeignTrue()
	}
	return ForeignFail()
}

// ==/2
func BuiltinTermEquals(m Machine, args []term.Term) ForeignReturn {
	a := args[0]
	b := args[1]
	if !term.Precedes(a, b) && !term.Precedes(b, a) {
		return ForeignTrue()
	}
	return ForeignFail()
}

// \==/2
func BuiltinTermNotEquals(m Machine, args []term.Term) ForeignReturn {
	a := args[0]
	b := args[1]
	if !term.Precedes(a, b) && !term.Precedes(b, a) {
		return ForeignFail()
	}
	return ForeignTrue()
}

// @</2
func BuiltinTermLess(m Machine, args []term.Term) ForeignReturn {
	a := args[0]
	b := args[1]
	if term.Precedes(a, b) {
		return ForeignTrue()
	}
	return ForeignFail()
}

// @=</2
func BuiltinTermLessEquals(m Machine, args []term.Term) ForeignReturn {
	a := args[0]
	b := args[1]
	if term.Precedes(a, b) {
		return ForeignTrue()
	}
	return BuiltinTermEquals(m, args)
}

// @>/2
func BuiltinTermGreater(m Machine, args []term.Term) ForeignReturn {
	a := args[0]
	b := args[1]
	if term.Precedes(b, a) {
		return ForeignTrue()
	}
	return ForeignFail()
}

// @>=/2
func BuiltinTermGreaterEquals(m Machine, args []term.Term) ForeignReturn {
	a := args[0]
	b := args[1]
	if term.Precedes(b, a) {
		return ForeignTrue()
	}
	return BuiltinTermEquals(m, args)
}

// (\+)/1
func BuiltinNot(m Machine, args []term.Term) ForeignReturn {
	var answer term.Bindings
	var err error
	m = m.ClearConjs().PushConj(args[0].(term.Callable))

	for {
		m, answer, err = m.Step()
		if err == MachineDone {
			return ForeignTrue()
		}
		MaybePanic(err)
		if answer != nil {
			return ForeignFail()
		}
	}
	panic("impossible")
}

// atom_codes/2 see ISO §8.16.5
func BuiltinAtomCodes2(m Machine, args []term.Term) ForeignReturn {

	if !term.IsVariable(args[0]) {
		atom := args[0].(*term.Atom)
		list := term.NewCodeList(atom.Name())
		return ForeignUnify(args[1], list)
	} else if !term.IsVariable(args[1]) {
		runes := make([]rune, 0)
		list := args[1].(term.Callable)
		for {
			switch list.Arity() {
			case 2:
				if list.Name() == "." {
					args := list.Arguments()
					code := args[0].(*term.Integer)
					runes = append(runes, code.Code())
					list = args[1].(term.Callable)
				}
			case 0:
				if list.Name() == "[]" {
					atom := term.NewAtom(string(runes))
					return ForeignUnify(args[0], atom)
				}
			default:
				msg := fmt.Sprintf("unexpected code list %s", args[1])
				panic(msg)
			}
		}
	}

	msg := fmt.Sprintf("atom_codes/2: error with the arguments %s and %s", args[0], args[1])
	panic(msg)
}

// atom_number/2 as defined in SWI-Prolog
func BuiltinAtomNumber2(m Machine, args []term.Term) (ret ForeignReturn) {
	number := args[1]

	if !term.IsVariable(args[0]) {
		atom := args[0].(term.Callable)
		defer func() { // convert parsing panics into fail
			if x := recover(); x != nil {
				ret = ForeignFail()
			}
		}()
		if strings.Contains(atom.Name(), ".") {
			number = term.NewFloat(atom.Name())
		} else {
			number = term.NewInt(atom.Name())
		}
		return ForeignUnify(args[1], number)
	} else if !term.IsVariable(number) {
		atom := term.NewAtom(number.String())
		return ForeignUnify(args[0], atom)
	}

	msg := fmt.Sprintf("atom_number/2: Arguments are not sufficiently instantiated: %s and %s", args[0], args[1])
	panic(msg)
}

// call/*
func BuiltinCall(m Machine, args []term.Term) ForeignReturn {

	// build a new goal with extra arguments attached
	bodyTerm := args[0].(term.Callable)
	functor := bodyTerm.Name()
	newArgs := make([]term.Term, 0)
	newArgs = append(newArgs, bodyTerm.Arguments()...)
	newArgs = append(newArgs, args[1:]...)
	goal := term.NewCallable(functor, newArgs...)

	// construct a machine that will prove this goal next
	return m.DemandCutBarrier().PushConj(goal)
}

// downcase_atom(+AnyCase, -LowerCase)
//
// Converts the characters of AnyCase into lowercase and unifies the
// lowercase atom with LowerCase.
func BuiltinDowncaseAtom2(m Machine, args []term.Term) ForeignReturn {
	if term.IsVariable(args[0]) {
		panic("downcase_atom/2: instantiation_error")
	}
	anycase := args[0].(term.Callable)
	if anycase.Arity() != 0 {
		msg := fmt.Sprintf("downcase_atom/2: type_error(atom, %s)", anycase)
		panic(msg)
	}

	lowercase := term.NewAtom(strings.ToLower(anycase.Name()))
	return ForeignUnify(args[1], lowercase)
}

// fail/0
func BuiltinFail(m Machine, args []term.Term) ForeignReturn {
	return ForeignFail()
}

// findall/3
func BuiltinFindall3(m Machine, args []term.Term) ForeignReturn {
	template := args[0]
	goal := args[1]

	// call(Goal), X=Template
	x := term.NewVar("_")
	call := term.NewCallable("call", goal)
	unify := term.NewCallable("=", x, template)
	prove := term.NewCallable(",", call, unify)
	proofs := m.ClearConjs().ClearDisjs().ProveAll(prove)

	// build a list from the results
	instances := make([]term.Term, 0)
	for _, proof := range proofs {
		t, err := proof.Resolve(x)
		MaybePanic(err)
		instances = append(instances, t)
	}

	return ForeignUnify(args[2], term.NewTermList(instances))
}

// listing/0
// This should be implemented in pure Prolog, but for debugging purposes,
// I'm doing it for now as a foreign predicate.  This will go away.
func BuiltinListing0(m Machine, args []term.Term) ForeignReturn {
	fmt.Println(m.String())
	return ForeignTrue()
}

// msort(+Unsorted:list, -Sorted:list) is det.
//
// True if Sorted is a sorted version of Unsorted.  Duplicates are
// not removed.
//
// This is currently implemented using Go's sort.Sort.
// The exact implementation is subject to change.  I make no
// guarantees about sort stability.
func BuiltinMsort2(m Machine, args []term.Term) ForeignReturn {
	terms := term.ProperListToTermSlice(args[0])
	sort.Sort((*term.TermSlice)(&terms))
	list := term.NewTermList(terms)
	return ForeignUnify(args[1], list)
}

// A temporary hack for debugging.  This will disappear once Golog has
// proper support for format/2
func BuiltinPrintf(m Machine, args []term.Term) ForeignReturn {
	template := args[0].(*term.Atom).Name()
	template = strings.Replace(template, "~n", "\n", -1)
	if len(args) == 1 {
		fmt.Printf(template)
	} else if len(args) == 2 {
		fmt.Printf(template, args[1])
	}
	return ForeignTrue()
}

// succ(?A:integer, ?B:integer) is det.
//
// True if B is one greater than A and A >= 0.
func BuiltinSucc2(m Machine, args []term.Term) ForeignReturn {
	x := args[0]
	y := args[1]
	zero := big.NewInt(0)

	if term.IsInteger(x) {
		a := x.(*term.Integer)
		if a.Value().Cmp(zero) < 0 {
			panic("succ/2: first argument must be 0 or greater")
		}
		result := new(big.Int).Add(a.Value(), big.NewInt(1))
		return ForeignUnify(y, term.NewBigInt(result))
	} else if term.IsInteger(y) {
		b := y.(*term.Integer)
		result := new(big.Int).Add(b.Value(), big.NewInt(-1))
		if result.Cmp(zero) < 0 {
			panic("succ/2: first argument must be 0 or greater")
		}
		return ForeignUnify(x, term.NewBigInt(result))
	}

	panic("succ/2: one argument must be an integer")
}
