package term

import . "fmt"

import "bytes"

// NewTerm creates a new term with the given functor and optional arguments
func NewTerm(functor string, arguments ...Term) Term {
    return &Compound{
        Func:   functor,
        Args:   arguments,
    }
}

// ISO calls this a "compound term" see §6.1.2(e)
// We currently use this type to cover atoms defined in §6.1.2(b)
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
