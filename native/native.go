package native

import (
	"fmt"
	"reflect"

	"github.com/mndrix/golog"
	"github.com/mndrix/golog/term"
)

type PrologStringer interface {
	PrologString() string
}

type Native struct {
	val interface{}
}

func IsNative(t term.Term) bool {
	_, ok := t.(*Native)
	return ok
}

func NewNative(v interface{}) *Native {
	return &Native{
		val: v,
	}
}

func (n *Native) Val() interface{} {
	return n.val
}

func (n *Native) ReplaceVariables(b term.Bindings) term.Term {
	return n
}

func (n *Native) String() string {
	rv := reflect.ValueOf(n.val)
	if rv.IsValid() && rv.CanInterface() {
		if v, ok := rv.Interface().(PrologStringer); ok {
			return v.PrologString()
		}
	}
	return fmt.Sprintf("native('%s')", rv.Type())
}

func (n *Native) Type() int {
	return term.AtomType
}

func (n *Native) Indicator() string {
	return "native/1"
}

func (n *Native) Unify(e term.Bindings, b term.Term) (term.Bindings, error) {
	switch t := b.(type) {
	case *term.Variable:
		return b.Unify(e, n)
	case *Native:
		if *n == *t {
			return e, nil
		}
		if n.val == nil {
			n.val = t.val
			return e, nil
		}
		if t.val == nil {
			t.val = n.val
			return e, nil
		}
		return e, term.CantUnify
	default:
		return e, term.CantUnify
	}
}

func NativeNil(m golog.Machine, args []term.Term) golog.ForeignReturn {
	if term.IsVariable(args[0]) {
		return golog.ForeignUnify(&Native{}, args[0])
	}
	if n, ok := args[0].(*Native); ok {
		if n.val == nil {
			return golog.ForeignTrue()
		}
	}
	return golog.ForeignFail()
}

func NewNativeMachine(interactive bool) golog.Machine {
	var m golog.Machine
	if interactive {
		m = golog.NewInteractiveMachine()
	} else {
		m = golog.NewMachine()
	}
	return m.RegisterForeign(
		map[string]golog.ForeignPredicate{
			"go_nil/1": NativeNil,
		},
	)
}
