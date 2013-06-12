package golog

// All of Golog's builtin, foreign-implemented predicates
// are defined here.

import "fmt"
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
	return m.PushConj(args[1]).PushConj(args[0])
}

// ground/1
func BuiltinGround(m Machine, args []term.Term) ForeignReturn {
	switch x := args[0].(type) {
	case *term.Variable:
		return ForeignFail()
	case *term.Atom,
		*term.Integer,
		*term.Float,
		*term.Error:
		return ForeignTrue()
	case *term.Compound:
		// recursively evaluate compound term's arguments
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

// ->/2
func BuiltinIfThen(m Machine, args []term.Term) ForeignReturn {
	cond := args[0]
	then := args[1]

	// CUT_BARRIER, (cond, !, then)
	cut := term.NewTerm("!")
	goal := term.NewTerm(",", cond, term.NewTerm(",", cut, then))
	return m.DemandCutBarrier().PushConj(goal)
}

// ;/2
//
// Implements disjunction and if-then-else.
func BuiltinSemicolon(m Machine, args []term.Term) ForeignReturn {
	arity := args[0].Arity()
	functor := args[0].Functor()
	if arity == 2 && functor == "->" { // §7.8.8
		return ifThenElse(m, args)
	}

	cp := NewSimpleChoicePoint(m, args[1])
	return m.PushDisj(cp).PushConj(args[0])
}
func ifThenElse(m Machine, args []term.Term) ForeignReturn {
	cond := args[0].Arguments()[0]
	then := args[0].Arguments()[1]
	els := args[1]

	// CUT_BARRIER, (call(cond), !, then; else)
	cut := term.NewTerm("!")
	cond = term.NewTerm("call", cond)
	goal := term.NewTerm(",", cond, term.NewTerm(",", cut, then))
	goal = term.NewTerm(";", goal, els)
	return m.DemandCutBarrier().PushConj(goal)
}

// =/2
func BuiltinUnify(m Machine, args []term.Term) ForeignReturn {
	return ForeignUnify(args[0], args[1])
}

// (\+)/1
func BuiltinNot(m Machine, args []term.Term) ForeignReturn {
	var answer term.Bindings
	var err error
	m = m.ClearConjs().PushConj(args[0])

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
	atom := args[0]
	list := args[1]

	if !term.IsVariable(atom) {
		list = term.NewCodeList(atom.Functor())
		return ForeignUnify(args[1], list)
	} else if !term.IsVariable(list) {
		runes := make([]rune, 0)
		for {
			switch list.Arity() {
			case 2:
				if list.Functor() == "." {
					args := list.Arguments()
					code := args[0].(*term.Integer)
					runes = append(runes, code.Code())
					list = args[1]
				}
			case 0:
				if list.Functor() == "[]" {
					atom = term.NewAtom(string(runes))
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
	atom := args[0]
	number := args[1]

	if !term.IsVariable(atom) {
		defer func() { // convert parsing panics into fail
			if x := recover(); x != nil {
				ret = ForeignFail()
			}
		}()
		if strings.Contains(atom.Functor(), ".") {
			number = term.NewFloat(atom.Functor())
		} else {
			number = term.NewInt(atom.Functor())
		}
		return ForeignUnify(args[1], number)
	} else if !term.IsVariable(number) {
		atom = term.NewAtom(number.String())
		return ForeignUnify(args[0], atom)
	}

	msg := fmt.Sprintf("atom_number/2: Arguments are not sufficiently instantiated: %s and %s", args[0], args[1])
	panic(msg)
}

// call/*
func BuiltinCall(m Machine, args []term.Term) ForeignReturn {
	// which goal is being called?
	bodyTerm := args[0]
	if term.IsVariable(bodyTerm) {
		bindings := m.Bindings()
		bodyTerm = bindings.Resolve_(bodyTerm.(*term.Variable))
	}

	// build a new goal with extra arguments attached
	functor := bodyTerm.Functor()
	newArgs := make([]term.Term, 0)
	newArgs = append(newArgs, bodyTerm.Arguments()...)
	newArgs = append(newArgs, args[1:]...)
	goal := term.NewTerm(functor, newArgs...)

	// construct a machine that will prove this goal next
	return m.DemandCutBarrier().PushConj(goal)
}

// downcase_atom(+AnyCase, -LowerCase)
//
// Converts the characters of AnyCase into lowercase and unifies the
// lowercase atom with LowerCase.
func BuiltinDowncaseAtom2(m Machine, args []term.Term) ForeignReturn {
	anycase := args[0]
	if term.IsVariable(anycase) {
		panic("downcase_atom/2: instantiation_error")
	}
	if anycase.Arity() != 0 {
		msg := fmt.Sprintf("downcase_atom/2: type_error(atom, %s)", anycase)
		panic(msg)
	}

	lowercase := term.NewAtom(strings.ToLower(anycase.Functor()))
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
	call := term.NewTerm("call", goal)
	unify := term.NewTerm("=", x, template)
	prove := term.NewTerm(",", call, unify)
	proofs := m.ClearConjs().ProveAll(prove)

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
	template := args[0].Functor()
	template = strings.Replace(template, "~n", "\n", -1)
	if len(args) == 1 {
		fmt.Printf(template)
	} else if len(args) == 2 {
		fmt.Printf(template, args[1])
	}
	return ForeignTrue()
}
