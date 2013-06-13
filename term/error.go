package term

import . "fmt"
import "github.com/mndrix/golog/lex"

type Error struct {
	message string
	eme     *lex.Eme
}

// Returns an error term.  These are used internally for error handling and
// may disappear in the future.
func NewError(message string, eme *lex.Eme) Term {
	return &Error{
		message: message,
		eme:     eme,
	}
}

func (self *Error) Functor() string {
	panic("Errors have no Functor()")
}

func (self *Error) Arity() int {
	panic("Errors have no Arity()")
}

func (self *Error) Arguments() []Term {
	panic("Errors have no Arguments()")
}

func (self *Error) String() string {
	return Sprintf("%s at %s", self.message, self.eme.Pos)
}

func (self *Error) Type() int {
	return ErrorType
}

func (self *Error) Indicator() string {
	panic("Errors have no Indicator()")
}

func (self *Error) ReplaceVariables(env Bindings) Term {
	return self
}

func (a *Error) Unify(e Bindings, b Term) (Bindings, error) {
	panic("Errors can't Unify()")
}
