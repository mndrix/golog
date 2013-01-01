package golog

import . "fmt"
import . "regexp"
import . "strings"
import "bytes"

var anonCounter <-chan int64
func init() {
    // goroutine providing a counter for anonymous variables
    c := make(chan int64)
    var i int64 = 1
    go func() {
        for {
            c <- i
            i++
        }
    }()
    anonCounter = c
}

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

    // IsClause returns true if the term is like 'Head :- Body'
    IsClause() bool

    // String provides a string representation of a term
    String() string

    // Indicator() provides a "predicate indicator" representation of a term
    Indicator() string
}

type Structure struct {
    Func    string
    Args    []Term
}
func (self *Structure) Functor() string {
    return self.Func
}
func (self *Structure) Arity() int {
    return len(self.Args)
}
func (self *Structure) Arguments() []Term {
    return self.Args
}
func (self *Structure) Body() Term {
    return self.Args[1]
}
func (self *Structure) Head() Term {
    return self.Args[0]
}
func (self *Structure) IsClause() bool {
    return self.Arity() == 2 && self.Functor() == ":-"
}
func (self *Structure) String() string {
    // an atom
    quotedFunctor := QuoteFunctor(self.Functor())
    if self.Arity() == 0 {
        return quotedFunctor
    }

    var buf bytes.Buffer
    Fprintf(&buf, "%s(", quotedFunctor)
    arity := self.Arity()
    for i := 0; i<arity; i++ {
        if i>0 {
            Fprintf(&buf, ", ")
        }
        Fprintf(&buf, "%s", self.Arguments()[i])
    }
    Fprintf(&buf, ")")
    return buf.String()
}
func (self *Structure) Indicator() string {
    return Sprintf("%s/%d", self.Functor(), self.Arity())
}


type Variable struct {
    Name    string
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
func (self *Variable) Body() Term {
    panic("Variables have no Body()")
}
func (self *Variable) Head() Term {
    panic("Variables have no Head()")
}
func (self *Variable) IsClause() bool {
    return false
}
func (self *Variable) String() string {
    return self.Name
}
func (self *Variable) Indicator() string {
    return Sprintf("%s", self.Name)
}

// NewTerm creates a new term with the given functor and optional arguments
func NewTerm(functor string, arguments ...Term) Term {
    return &Structure{
        Func:   functor,
        Args:   arguments,
    }
}

func NewVar(name string) Term {
    // sanity check the variable name's syntax
    isCapitalized, err := MatchString(`^[A-Z_]`, name)
    maybePanic(err)
    if !isCapitalized {
        panic("Variable names must start with a capital letter or underscore")
    }

    // make sure anonymous variables are unique
    if name == "_" {
        i := <-anonCounter
        name = Sprintf("_G%d", i)
    }
    return &Variable{
        Name:   name,
    }
}


func IsVariable(t Term) bool {
    switch t.(type) {
        case *Structure:
            return false
        case *Variable:
            return true
    }
    msg := Sprintf("Unexpected term type: %#v", t)
    panic(msg)
}

// QuoteFunctor returns a canonical representation of a term's name
// by quoting characters that require quoting
func QuoteFunctor(name string) string {
    nonAlpha, err := MatchString(`\W`, name)
    maybePanic(err)
    nonLower, err := MatchString(`^[^a-z]`, name)
    if nonAlpha || nonLower {
        escapedName := Replace(name, `'`, `\'`, -1)
        return Sprintf("'%s'", escapedName)
    }

    return name
}

func maybePanic(err error) {
    if err != nil {
        panic(err)
    }
}
