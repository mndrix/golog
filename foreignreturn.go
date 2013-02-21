package golog

// simulate an algebraic datatype representing the return value
// of foreign predicates

import "github.com/mndrix/golog/term"

// ForeignReturn represents the return type of ForeignPredicate functions.
// Values of ForeignReturn indicate certain success or failure conditions
// to the Golog machine.  If you're writing a foreign predicate, see
// functions named Foreign* for possible return values.
type ForeignReturn interface {
    IsaForeignReturn()      // useless method to restrict implementations
}


// ForeignTrue indicates a foreign predicate that succeeds deterministically
func ForeignTrue() ForeignReturn {
    t := true
    return (*foreignTrue)(&t)
}
type foreignTrue bool
func (*foreignTrue) IsaForeignReturn() {}


// ForeignFail indicates a foreign predicate that failed
func ForeignFail() ForeignReturn {
    f := false
    return (*foreignFail)(&f)
}
type foreignFail bool
func (*foreignFail) IsaForeignReturn() {}


// ForeignUnify indicates a predicate that succeeds or fails depending
// on the results of a unification.  The list of arguments must have
// an even number of elements.  Each pair is unified in turn.  Variables
// bound during unification become part of the Golog machines's bindings.
func ForeignUnify(ts... term.Term) ForeignReturn {
    if len(ts) % 2 != 0 {
        panic("Uneven number of arguments to ForeignUnify")
    }
    return (*foreignUnify)(&ts)
}
type foreignUnify []term.Term
func (*foreignUnify) IsaForeignReturn() {}
