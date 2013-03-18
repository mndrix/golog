package term

import . "fmt"

import "bytes"
import "strings"

// NewTerm creates a new term with the given functor and optional arguments
func NewTerm(functor string, arguments ...Term) Term {
    return &Compound{
        Func:   functor,
        Args:   arguments,
    }
}

// NewAtom creates a new atom with the given name.
// This is just a 0-arity compound term, for now.  Eventually, it will
// have an optimized implementation.
func NewAtom(name string) Term {
    return NewTerm(name)
}

// Unlikely to be useful outside of the parser
func NewAtomFromLexeme(possiblyQuotedName string) Term {
    if len(possiblyQuotedName) == 0 {
        panic("Atoms must have some content")
    }
    name := possiblyQuotedName

    // remove quote characters, if they exist
    runes := []rune(possiblyQuotedName)
    if runes[0] == '\'' {
        if runes[len(runes)-1] == '\'' {
            name = string(runes[1:len(runes)-1])
            name = strings.Replace(name, `''`, `'`, -1)
        } else {
            msg := Sprintf("Atom needs closing quote: %s", possiblyQuotedName)
            panic(msg)
        }
    }

    return NewTerm(name)
}

// NewCodeList returns a compound term consisting of the character codes
// of the given string.  The internal representation may eventually optimize
// for storing character codes.
func NewCodeList(s string) Term {
    runes := []rune(s)
    list := NewTerm("[]")
    for i:=len(runes)-1; i>=0; i-- {
        list = NewTerm(".", NewCode(runes[i]), list)
    }
    return list
}

// NewTermList returns a list term consisting of each term from the slice.
// A future implementation may optimize the data structure that's returned.
func NewTermList(terms []Term) Term {
    list := NewAtom("[]")
    for i:=len(terms)-1; i>=0; i-- {
        list = NewTerm(".", terms[i], list)
    }
    return list
}

// ISO calls this a "compound term" see §6.1.2(e)
// We currently use this type to cover atoms defined in §6.1.2(b)
// by treating atoms as compound terms with 0 arity.
type Compound struct {
    Func    string
    Args    []Term
}
func (self *Compound) Functor() string {
    return self.Func
}
func (self *Compound) Arity() int {
    return len(self.Args)
}
func (self *Compound) Arguments() []Term {
    return self.Args
}
func (self *Compound) Body() Term {
    return self.Args[1]
}
func (self *Compound) Head() Term {
    return self.Args[0]
}
func (self *Compound) IsClause() bool {
    return self.Arity() == 2 && self.Functor() == ":-"
}
func (self *Compound) String() string {
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
func (self *Compound) Indicator() string {
    return Sprintf("%s/%d", self.Functor(), self.Arity())
}
func (self *Compound) Error() error {
    panic("Can't call Error() on a Structure")
}

func (self *Compound) ReplaceVariables(env Bindings) Term {
    arity := self.Arity()
    if arity == 0 {
        return self     // atoms have no variables to replace
    }

    newArgs := make([]Term, arity)
    for i, arg := range self.Arguments() {
        newArgs[i] = arg.ReplaceVariables(env)
    }
    return NewTerm(self.Functor(), newArgs...)
}

func (a *Compound) Unify(e Bindings, b Term) (Bindings, error) {
    if IsVariable(b) {
        return b.Unify(e, a)
    }
    if !IsCompound(b) {
        return e, CantUnify
    }

    // functor and arity must match for unification to work
    arity := a.Arity()
    if arity != b.Arity() {
        return e, CantUnify
    }
    if a.Functor() != b.Functor() {
        return e, CantUnify
    }

    // try unifying each subterm
    var err error
    env := e
    aArgs := a.Arguments()
    bArgs := b.Arguments()
    for i:=0; i<arity; i++ {
        env, err = aArgs[i].Unify(env, bArgs[i])
        if err != nil {
            return e, err // return original environment along with error
        }
    }

    // unification succeeded
    return env, nil
}
