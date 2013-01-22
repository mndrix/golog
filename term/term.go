// Represent and unify Prolog terms.
package term

import . "fmt"
import . "regexp"
import "strings"
import "github.com/mndrix/golog/lex"

// Term represents a single Prolog term which might be an atom, a structure,
// a number, etc.
type Term interface {
    // Functor returns the term's name
    Functor() string

    // Arity returns the number of arguments a term has. An atom has 0 arity.
    Arity() int

    // Arguments returns a slice of this term's arguments, if any
    Arguments() []Term

    // Body returns a term's second argument; otherwise, panics
    Body() Term

    // Head returns a term's first argument; otherwise, panics
    Head() Term

    // Error returns an error value if this is an error term
    Error() error

    // IsClause returns true if the term is like 'Head :- Body'
    IsClause() bool

    // String provides a string representation of a term
    String() string

    // Indicator() provides a "predicate indicator" representation of a term
    Indicator() string
}

func IsVariable(t Term) bool {
    switch t.(type) {
        case *Compound:
            return false
        case *Variable:
            return true
        case *Error:
            return false
    }
    msg := Sprintf("Unexpected term type: %#v", t)
    panic(msg)
}

func IsError(t Term) bool {
    switch t.(type) {
        case *Compound:
            return false
        case *Variable:
            return false
        case *Error:
            return true
    }
    msg := Sprintf("Unexpected term type: %#v", t)
    panic(msg)
}

// QuoteFunctor returns a canonical representation of a term's name
// by quoting characters that require quoting
func QuoteFunctor(name string) string {
    // cons must be quoted (to avoid confusion with full stop)
    if name == "." {
        return Sprintf("'%s'", name)
    }

    // names composed entirely of graphic characters need no quoting
    allGraphic := true
    for _, c := range name {
        if !lex.IsGraphic(c) {
            allGraphic = false
            break
        }
    }
    if allGraphic || name == "[]" {
        return name
    }

    nonAlpha, err := MatchString(`\W`, name)
    maybePanic(err)
    nonLower, err := MatchString(`^[^a-z]`, name)
    if nonAlpha || nonLower {
        escapedName := strings.Replace(name, `'`, `\'`, -1)
        return Sprintf("'%s'", escapedName)
    }

    return name
}

func maybePanic(err error) {
    if err != nil {
        panic(err)
    }
}
