package native

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/mndrix/golog"
	"github.com/mndrix/golog/term"
)

type Decoder struct {
	m golog.Machine
}

func NewDecoder(m golog.Machine) *Decoder {
	return &Decoder{
		m: m,
	}
}

func (d *Decoder) Decode(t term.Term, val interface{}) error {
	return d.pgValue(t, reflect.ValueOf(val))
}

func (d *Decoder) pgValue(t term.Term, val reflect.Value) error {
	if !val.IsValid() {
		return nil
	}
	if val.CanInterface() {
		if _, ok := val.Interface().(Marshaler); ok {
			if val.CanAddr() {
				val.Set(reflect.New(val.Type().Elem()))
			} else {
				val.Elem().Set(reflect.New(val.Elem().Type()).Elem())
			}
			return val.Interface().(Marshaler).MarshalProlog(d.m, t)
		}
		maybeVal := reflect.New(val.Type())
		if _, ok := maybeVal.Interface().(Marshaler); ok {
			err := maybeVal.Interface().(Marshaler).MarshalProlog(d.m, t)
			if err != nil {
				return err
			}
			val.Set(maybeVal.Elem())
			return nil
		}
	}
	switch val.Type().Kind() {
	case reflect.Bool:
		return d.pgBool(t, val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return d.pgInt(t, val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return d.pgUint(t, val)
	case reflect.Float32, reflect.Float64:
		return d.pgFloat(t, val)
	case reflect.Complex64, reflect.Complex128:
		return d.pgComplex(t, val)
	case reflect.Array:
		return d.pgArray(t, val)
	case reflect.Slice:
		return d.pgSlice(t, val)
	case reflect.String:
		return d.pgString(t, val)
	case reflect.Struct:
		return d.pgStruct(t, val)
	case reflect.Chan, reflect.Func, reflect.Map,
		reflect.Uintptr, reflect.UnsafePointer:
		val.Set(reflect.ValueOf(t.(*Native).val))
		return nil
	case reflect.Ptr:
		if IsNative(t) {
			nv := reflect.ValueOf(t.(*Native).val)
			if nv.IsValid() && !nv.IsNil() {
				val.Elem().Set(nv)
			} else {
				val.Elem().Set(reflect.New(val.Type()))
			}
			return nil
		}
		return d.pgValue(t, val.Elem())
	case reflect.Interface:
		if nt, ok := t.(*Native); ok {
			if nt.val != nil {
				val.Set(reflect.ValueOf(nt.val))
			} else {
				val.Set(reflect.Zero(val.Type()))
			}
			return nil
		}
		return d.pgValue(t, val.Elem())
	}
	return fmt.Errorf("Go added new type: %s", val.Type().Kind())
}

func (d *Decoder) pgBool(t term.Term, val reflect.Value) error {
	if !term.IsAtom(t) {
		return fmt.Errorf("Expected bool, but got: %+v", t)
	}
	b := t.(*term.Atom)
	val.Set(reflect.ValueOf(b.Name() == "yes"))
	return nil
}

func (d *Decoder) pgInt(t term.Term, val reflect.Value) error {
	if !term.IsInteger(t) {
		return fmt.Errorf("Expected integer, but got: %+v", t)
	}
	i := t.(*term.Integer)
	bi := i.Value()
	// TODO(olegs): Handle overflow
	switch val.Type().Kind() {
	case reflect.Int:
		val.Set(reflect.ValueOf(int(bi.Int64())))
	case reflect.Int8:
		val.Set(reflect.ValueOf(int8(bi.Int64())))
	case reflect.Int16:
		val.Set(reflect.ValueOf(int16(bi.Int64())))
	case reflect.Int32:
		val.Set(reflect.ValueOf(int32(bi.Int64())))
	case reflect.Int64:
		val.Set(reflect.ValueOf(bi.Int64()))
	}
	return nil
}

func (d *Decoder) pgUint(t term.Term, val reflect.Value) error {
	if !term.IsInteger(t) {
		return fmt.Errorf("Expected integer, but got: %+v", t)
	}
	i := t.(*term.Integer)
	bi := i.Value()
	// TODO(olegs): Handle overflow
	switch val.Type().Kind() {
	case reflect.Uint:
		val.Set(reflect.ValueOf(uint(bi.Int64())))
	case reflect.Uint8:
		val.Set(reflect.ValueOf(uint8(bi.Int64())))
	case reflect.Uint16:
		val.Set(reflect.ValueOf(uint16(bi.Int64())))
	case reflect.Uint32:
		val.Set(reflect.ValueOf(uint32(bi.Int64())))
	case reflect.Uint64:
		val.Set(reflect.ValueOf(uint64(bi.Int64())))
	}
	return nil
}

func (d *Decoder) pgFloat(t term.Term, val reflect.Value) error {
	if !term.IsFloat(t) {
		return fmt.Errorf("Expected integer, but got: %+v", t)
	}
	f := t.(*term.Float)
	// TODO(olegs): Handle overflow
	val.Set(reflect.ValueOf(f.Value()))
	return nil
}

func (d *Decoder) pgComplex(t term.Term, val reflect.Value) error {
	if !term.IsCallable(t) || t.(term.Callable).Name() != "complex" {
		return fmt.Errorf("Expected complex, but got: %+v", t)
	}
	c := t.(term.Callable)
	if len(c.Arguments()) != 2 {
		return fmt.Errorf("Malformed complex: %+v", c)
	}
	pr := c.Arguments()[0]
	pi := c.Arguments()[1]
	if !term.IsFloat(pr) {
		return fmt.Errorf("Real part must be a float: %+v", pr)
	}
	if !term.IsFloat(pi) {
		return fmt.Errorf("Imaginary part must be a float: %+v", pi)
	}
	gr := pr.(*term.Float).Value()
	gi := pi.(*term.Float).Value()
	val.Set(reflect.ValueOf(complex(gr, gi)))
	return nil
}

func (d *Decoder) pgArray(t term.Term, val reflect.Value) error {
	return nil
}

func (d *Decoder) pgSlice(t term.Term, val reflect.Value) error {
	if !term.IsList(t) {
		return fmt.Errorf("%v is not a slice", t)
	}
	len := d.listLen(t)
	slice := reflect.MakeSlice(val.Type(), len, len)
	for i := 0; i < len; i++ {
		c := t.(term.Callable)
		e := c.Arguments()[0]
		t = c.Arguments()[1]
		if err := d.pgValue(e, slice.Index(i)); err != nil {
			return err
		}
	}
	val.Set(slice)
	return nil
}

func (d *Decoder) listLen(t term.Term) (len int) {
	for !term.IsEmptyList(t) {
		t = t.(term.Callable).Arguments()[1]
		len++
	}
	return len
}

func (d *Decoder) pgString(t term.Term, val reflect.Value) error {
	if !term.IsString(t) {
		return fmt.Errorf("%v is not a string", t)
	}
	b := bytes.NewBuffer([]byte{})
	if term.IsCompound(t) {
		c := t.(*term.Compound)
		for {
			args := c.Arguments()
			i := args[0].(*term.Integer)
			b.WriteRune(i.Code())
			if term.IsEmptyList(args[1]) {
				break
			} else {
				c = args[1].(*term.Compound)
			}
		}
	}
	val.Set(reflect.ValueOf(string(b.Bytes())))
	return nil
}

func (d *Decoder) pgStruct(t term.Term, val reflect.Value) error {
	names := map[string]term.Term{}
	if !term.IsCallable(t) {
		return fmt.Errorf("%s must be *term.Callable", t)
	}
	c := t.(term.Callable)
	if len(c.Arguments()) == 0 {
		return nil
	}
	if !term.IsList(c.Arguments()[0]) {
		return fmt.Errorf("fields of %s must be inside a list", t)
	}
	for _, a := range term.ListToSlice(c.Arguments()[0]) {
		if !term.IsCallable(a) {
			return fmt.Errorf("%s must be *term.Callable", a)
		}
		f := a.(term.Callable)
		args := f.Arguments()
		if len(args) != 1 {
			return fmt.Errorf("invalid field specification: %s", f)
		}
		names[f.Name()] = args[0]
	}
	for i := 0; i < val.NumField(); i++ {
		tf := val.Type().Field(i)
		if tf.PkgPath == "" {
			tag := tf.Tag.Get("prolog")
			if tag == "" {
				tag = gpName(tf.Name)
			}
			f := val.Field(i)
			fv := names[tag]
			if fv != nil {
				if err := d.pgValue(fv, f); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
