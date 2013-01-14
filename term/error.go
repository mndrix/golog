package term

import . "fmt"

type Error string

func (self *Error) Functor() string {
    panic("Errors have no Functor()")
}

func (self *Error) Arity() int {
    panic("Errors have no Arity()")
}

func (self *Error) Arguments() []Term {
    panic("Errors have no Arguments()")
}

func (self *Error) Body() Term {
    panic("Errors have no Body()")
}

func (self *Error) Head() Term {
    panic("Errors have no Head()")
}

func (self *Error) IsClause() bool {
    return false
}

func (self *Error) String() string {
    return string(*self)
}

func (self *Error) Indicator() string {
    panic("Errors have no Indicator()")
}

func (self *Error) Error() error {
    return Errorf("%s", *self)
}
