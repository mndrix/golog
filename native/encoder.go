package native

import (
	"math"
	"math/big"
	"reflect"

	"github.com/mndrix/golog/term"
)

type Encoder struct {
	p map[uintptr]term.Term
}

func NewEncoder() *Encoder {
	return &Encoder{
		p: map[uintptr]term.Term{},
	}
}

func (e *Encoder) Encode(val interface{}) term.Term {
	return e.gpValue(reflect.ValueOf(val))
}

func (e *Encoder) gpStruct(val reflect.Value) term.Callable {
	name := gpName(val.Type().Name())
	return term.NewCallable(name, term.SliceToList(e.gpFields(val)))
}

func (e *Encoder) gpFields(val reflect.Value) (fields []term.Term) {
	for i := 0; i < val.NumField(); i++ {
		if val.Type().Field(i).PkgPath == "" {
			t := e.gpValue(val.Field(i))
			c := term.NewCallable(gpName(val.Type().Field(i).Name), t)
			fields = append(fields, c)
		}
	}
	return fields
}

func (e *Encoder) gpBool(val reflect.Value) term.Term {
	b := false
	rb := reflect.ValueOf(&b)
	rb.Elem().Set(reflect.ValueOf(val.Bool()))
	var v string
	if b {
		v = "yes"
	} else {
		v = "no"
	}
	return term.NewAtom(v)
}

func (e *Encoder) gpString(val reflect.Value) term.Term {
	var s string
	rs := reflect.ValueOf(&s)
	rs.Elem().Set(val)
	return term.NewCodeList(s)
}

func (e *Encoder) gpInt(val reflect.Value) term.Term {
	return term.NewBigInt(big.NewInt(val.Int()))
}

func (e *Encoder) gpUint(val reflect.Value) term.Term {
	ui := val.Uint()
	if ui > math.MaxInt64 {
		res := big.NewInt(math.MaxInt64)
		res2 := big.NewInt(int64(ui - math.MaxInt64))
		return term.NewBigInt(res.Add(res, res2))
	}
	return term.NewBigInt(big.NewInt(int64(ui)))
}

func (e *Encoder) gpList(val reflect.Value) (t term.Term) {
	t = term.NewAtom("[]")
	for i := val.Len() - 1; i >= 0; i-- {
		e := e.gpValue(val.Index(i))
		t = term.NewCallable(".", e, t)
	}
	return t
}

func (e *Encoder) gpFloat(val reflect.Value) term.Term {
	return term.NewFloat64(val.Float())
}

func (e *Encoder) gpComplex(val reflect.Value) term.Term {
	c := val.Complex()
	return term.NewCallable(
		"complex",
		term.NewFloat64(real(c)),
		term.NewFloat64(imag(c)),
	)
}

func (e *Encoder) gpValue(val reflect.Value) (t term.Term) {
	if !val.IsValid() {
		return NewNative(nil)
	}
	if val.CanInterface() {
		if v, ok := val.Interface().(Unmarshaler); ok {
			t = v.UnmarshalProlog(e.p)
			e.p[val.Pointer()] = t
			return t
		}
		maybeVal := reflect.New(val.Type())
		if _, ok := maybeVal.Interface().(Unmarshaler); ok {
			maybeVal.Elem().Set(val)
			t = maybeVal.Interface().(Unmarshaler).UnmarshalProlog(e.p)
			return t
		}
	}
	switch val.Type().Kind() {
	case reflect.Bool:
		t = e.gpBool(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		t = e.gpInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		t = e.gpUint(val)
	case reflect.Float32, reflect.Float64:
		t = e.gpFloat(val)
	case reflect.Complex64, reflect.Complex128:
		t = e.gpComplex(val)
	case reflect.Array, reflect.Slice:
		t = e.gpList(val)
	case reflect.String:
		t = e.gpString(val)
	case reflect.Struct:
		t = e.gpStruct(reflect.ValueOf(val.Interface()))
	case reflect.Chan, reflect.Func, reflect.Map, reflect.UnsafePointer:
		if mt := e.p[val.Pointer()]; mt != nil {
			t = mt
		} else {
			t = &Native{
				val: val.Interface(),
			}
			e.p[val.Pointer()] = t
		}
	case reflect.Uintptr:
		ptr := uintptr(val.Uint())
		if mt := e.p[ptr]; mt != nil {
			t = mt
		} else {
			t = &Native{
				val: val.Interface(),
			}
			e.p[ptr] = t
		}
	case reflect.Ptr:
		if mt := e.p[val.Pointer()]; mt != nil {
			t = mt
		} else {
			t = e.gpValue(reflect.Indirect(val.Elem()))
			e.p[val.Pointer()] = t
		}
	case reflect.Interface:
		p := val.InterfaceData()[1]
		if mt := e.p[p]; mt != nil {
			t = mt
		} else {
			t = e.gpValue(reflect.Indirect(val.Elem()))
			e.p[p] = t
		}
	}
	return t
}
