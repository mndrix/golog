// Represent and unify Prolog terms.  Along with golog.Machine, term.Term
// is one of the most important data types in Golog.  It provides a Go
// representation of Prolog terms.  Terms represent Prolog code,
// Prolog queries and Prolog results.
//
// The current term API is messy and will definitely change in the future.
package term

import . "fmt"
import . "regexp"
import "strings"
import "github.com/mndrix/golog/lex"
import "github.com/mndrix/ps"

// Term represents a single Prolog term which might be an atom, a
// compound structure, an integer, etc.  Many methods on Term will
// be replaced with functions in the future.  The Term interface is
// also likely to be split into several smaller interfaces like Atomic,
// Number, etc.
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

    // ReplaceVariables replaces any internal variables with the values
    // to which they're bound.  Unbound variables are left as they are
    ReplaceVariables(Bindings) Term

    // String provides a string representation of a term
    String() string

    // Indicator() provides a "predicate indicator" representation of a term
    Indicator() string

    // Unifies the invocant and another term in the presence of an
    // environment.
    // On succes, returns a new environment with additional variable
    // bindings.  On failure, returns CantUnify error along with the
    // original environment
    Unify(Bindings, Term) (Bindings, error)
}

// Returns true if term t is a compound term.
func IsCompound(t Term) bool {
    switch t.(type) {
        case *Compound:
            return true
        case *Variable:
            return false
        case *Integer:
            return false
        case *Float:
            return false
        case *Error:
            return false
    }
    msg := Sprintf("Unexpected term type: %#v", t)
    panic(msg)
}

// Returns true if term t is a variable.
func IsVariable(t Term) bool {
    switch t.(type) {
        case *Compound:
            return false
        case *Variable:
            return true
        case *Integer:
            return false
        case *Float:
            return false
        case *Error:
            return false
    }
    msg := Sprintf("Unexpected term type: %#v", t)
    panic(msg)
}

// Returns true if term t is an error term.
func IsError(t Term) bool {
    switch t.(type) {
        case *Compound:
            return false
        case *Variable:
            return false
        case *Integer:
            return false
        case *Float:
            return false
        case *Error:
            return true
    }
    msg := Sprintf("Unexpected term type: %#v", t)
    panic(msg)
}

// Returns true if term t is a directive like `:- foo.`
func IsDirective(t Term) bool {
    return t.Indicator() == ":-/1"
}

// Returns true if term t is an integer
func IsInteger(t Term) bool {
    switch t.(type) {
        case *Compound:
            return false
        case *Variable:
            return false
        case *Integer:
            return true
        case *Float:
            return false
        case *Error:
            return false
    }
    msg := Sprintf("Unexpected term type: %#v", t)
    panic(msg)
}

// Returns true if term t is an floating point number
func IsFloat(t Term) bool {
    switch t.(type) {
        case *Compound:
            return false
        case *Variable:
            return false
        case *Integer:
            return false
        case *Float:
            return true
        case *Error:
            return false
    }
    msg := Sprintf("Unexpected term type: %#v", t)
    panic(msg)
}

// RenameVariables returns a new term like t with all variables replaced
// by fresh ones.
func RenameVariables(t Term) Term {
    renamed := make(map[string]*Variable)
    return renameVariables(t, renamed)
}

func renameVariables(t Term, renamed map[string]*Variable) Term {
    switch x := t.(type) {
        case *Float:    return x
        case *Integer:  return x
        case *Error:    return x
        case *Compound:
            if x.Arity() == 0 { return t }  // no variables in atoms
            newArgs := make([]Term, x.Arity())
            for i, arg := range x.Arguments() {
                newArgs[i] = renameVariables(arg, renamed)
            }
            return NewTerm(x.Functor(), newArgs...)
        case *Variable:
            name := x.Name
            v, ok := renamed[name]
            if ok {
                return v
            } else {
                v = x.WithNewId()
                renamed[name] = v
                return v
            }
    }
    panic("Unexpected term implementation")
}

// Variables returns a ps.Map whose keys are human-readable variable names
// and those values are *Variable used inside term t.
func Variables(t Term) ps.Map {
    names := ps.NewMap()
    switch x := t.(type) {
        case *Float:    return names
        case *Integer:  return names
        case *Error:    return names
        case *Compound:
            if x.Arity() == 0 { return names }  // no variables in an atom
            for _, arg := range x.Arguments() {
                innerNames := Variables(arg)
                innerNames.ForEach(func (key string, val ps.Any) {
                    names = names.Set(key, val)
                })
            }
            return names
        case *Variable:
            return names.Set(x.Name, x)
    }
    panic("Unexpected term implementation")
}

// QuoteFunctor returns a canonical representation of a term's name
// by quoting characters that require quoting
func QuoteFunctor(name string) string {
    // cons must be quoted (to avoid confusion with full stop)
    if name == "." || name == "" {
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
    if allGraphic || name == "[]" || name == "!" || name == ";" {
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

// NewCodeList constructs a list of character codes from a string.
// The string should include opening and closing " characters.
// Nominally, the resulting term is just a chain of cons cells ('.'/2),
// but it might actually be a more efficient implementation under the hood.
func NewCodeListFromDoubleQuotedString(s string) Term {
    // make sure the content is long enough
    runes := []rune(s)
    end := len(runes) - 2
    if end < 0 {
        msg := Sprintf("Code list string must have bracketing double quotes: %s", s)
        panic(msg)
    }

    // build a cons cell chain, starting at the end ([])
    codes := NewTerm(`[]`)
    for i := end; i > 0; i-- {
        c := NewCode(runes[i])
        codes = NewTerm(`.`, c, codes)
    }

    return codes
}

// Precedes returns true if the first argument 'term-precedes'
// the second argument according to ISO §7.2
func Precedes(a, b Term) bool {
    aP := precedence(a)
    bP := precedence(b)
    if aP < bP { return true }
    if aP > bP { return false }

    // both terms have the same precedence by type, so delve deeper
    switch x := a.(type) {
        case *Variable:
            y := b.(*Variable)
            return x.Id() < y.Id()
        case *Float:
            y := b.(*Float)
            return x.Value() < y.Value()
        case *Integer:
            y := b.(*Integer)
            return x.Value().Cmp(y.Value()) < 0
        case *Compound:
            y := b.(*Compound)
            if x.Arity() == 0 { // atoms
                return x.Functor() < y.Functor()
            } else {    // compound terms
                if x.Arity() < y.Arity() { return true }
                if x.Arity() > y.Arity() { return false }
                if x.Functor() < y.Functor() { return true }
                if x.Functor() > y.Functor() { return false }
                for i:=0; i<x.Arity(); i++ {
                    if Precedes(x.Arguments()[i], y.Arguments()[i]) {
                        return true
                    } else if Precedes(y.Arguments()[i], x.Arguments()[i]) {
                        return false
                    }
                }
                return false    // identical terms
            }
    }

    msg := Sprintf("Unexpected term type %s\n", a)
    panic(msg)
}
func precedence(t Term) int {
    switch t.(type) {
        case *Variable:
            return 0
        case *Float:
            return 1
        case *Integer:
            return 2
        case *Compound:
            if t.Arity() == 0 { return 3 }
            return 4
    }
    msg := Sprintf("Unexpected term type %s\n", t)
    panic(msg)
}

// Converts a '.'/2 list terminated in []/0 into a slice of the associated
// terms.  Panics if the argument is not a proper list.
func ProperListToTermSlice(t Term) []Term {
    l := make([]Term, 0)
    if !IsCompound(t) { panic("Not a list") }
    for {
        switch t.Indicator() {
            case "[]/0":
                return l
            case "./2":
                l = append(l, t.Arguments()[0])
                t = t.Arguments()[1]
            default:
                panic("Improper list")
        }
    }
    return l
}

// Implement sort.Interface for []Term
type TermSlice []Term
func (self *TermSlice) Len() int {
    ts := []Term(*self)
    return len(ts)
}
func (self *TermSlice) Less(i, j int) bool {
    ts := []Term(*self)
    return Precedes(ts[i], ts[j]);
}
func (self *TermSlice) Swap(i, j int) {
    ts := []Term(*self)
    tmp := ts[i]
    ts[i] = ts[j]
    ts[j] = tmp
}

func maybePanic(err error) {
    if err != nil {
        panic(err)
    }
}
