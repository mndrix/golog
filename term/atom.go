package term

import . "fmt"

type Atom string

// NewAtom creates a new atom with the given name.
// This is just a 0-arity compound term, for now.  Eventually, it will
// have an optimized implementation.
func NewAtom(name string) Term {
    return (*Atom)(&name)
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
            raw := runes[1:len(runes)-1]
            unescaped := make([]rune, len(raw))
            var i, j int
            for i < len(raw) {
                if raw[i] == '\\' && i < len(raw)-1 && raw[i+1] == '\'' {
                    unescaped[j] = '\''
                    i += 2
                } else {
                    unescaped[j] = raw[i]
                    i++
                }
                j++
            }
            name = string(unescaped[0:j])
        } else {
            msg := Sprintf("Atom needs closing quote: %s", possiblyQuotedName)
            panic(msg)
        }
    }

    return NewAtom(name)
}

func (self *Atom) Functor() string {
    return string(*self)
}
func (self *Atom) Arity() int {
    return 0
}
func (self *Atom) Arguments() []Term {
    return make([]Term, 0)
}
func (self *Atom) Body() Term {
    panic("Atoms have no Body()")
}
func (self *Atom) Head() Term {
    panic("Atoms have no Head()")
}
func (self *Atom) IsClause() bool {
    return false
}
func (self *Atom) String() string {
    return QuoteFunctor(self.Functor())
}
func (self *Atom) Indicator() string {
    return Sprintf("%s/0", self.Functor())
}
func (self *Atom) Error() error {
    panic("Can't call Error() on an Atom")
}

func (self *Atom) ReplaceVariables(env Bindings) Term {
    return self
}

func (a *Atom) Unify(e Bindings, b Term) (Bindings, error) {
    switch t := b.(type) {
        case *Variable:
            return b.Unify(e, a)
        case *Atom:
            if *a == *t {
                return e, nil
            }
            return e, CantUnify
        default:
            return e, CantUnify
    }
    panic("Impossible")
}
