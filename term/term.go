// Represent and unify Prolog terms.
package term

import . "fmt"
import . "regexp"
import "strings"
import "github.com/mndrix/golog/lex"
import "github.com/mndrix/ps"

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

    // ReplaceVariables replaces any internal variables with the values
    // to which they're bound.  Unbound variables are left as they are
    ReplaceVariables(Bindings) Term

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

// Returns true if this term is a directive like :- foo.
func IsDirective(t Term) bool {
    return t.Indicator() == ":-/1"
}

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
// and those values are *Variable
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

// NewCodeList constructs a list of character codes from a string.
// The string should include opening and closing " characters.
// Nominally, the resulting term is just a chain of cons cells ('.'/2),
// but it might actually be a more efficient implementation under the hood.
func NewCodeList(s string) Term {
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

func maybePanic(err error) {
    if err != nil {
        panic(err)
    }
}
