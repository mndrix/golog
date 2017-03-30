package term

import (
	"bytes"
	"strings"
)

func IsEmptyList(t Term) bool {
	return IsAtom(t) && t.(*Atom).Name() == "[]"
}

func IsString(t Term) bool {
	if IsEmptyList(t) {
		return true
	}
	if !IsCompound(t) {
		return false
	}
	// TODO(olegs): Handle circular structures
	c := t.(*Compound)
	for {
		if c.Arity() != 2 {
			return false
		}
		if c.Func != "." {
			return false
		}
		args := c.Arguments()
		if !IsInteger(args[0]) {
			return false
		}
		if IsCompound(args[1]) {
			c = args[1].(*Compound)
		} else if IsEmptyList(args[1]) {
			break
		} else {
			return false
		}
	}
	return true
}

func IsList(t Term) bool {
	if IsEmptyList(t) {
		return true
	}
	if !IsCompound(t) {
		return false
	}
	// TODO(olegs): Handle circular structures
	c := t.(*Compound)
	for {
		if c.Arity() != 2 {
			return false
		}
		if c.Func != "." {
			return false
		}
		args := c.Arguments()
		if IsCompound(args[1]) {
			c = args[1].(*Compound)
		} else if IsEmptyList(args[1]) {
			break
		} else {
			return false
		}
	}
	return true
}

func PrettyList(t Term) string {
	b := bytes.NewBuffer([]byte{})
	_ = b.WriteByte('[')
	for !IsEmptyList(t) {
		c := t.(*Compound)
		_, _ = b.WriteString(c.Arguments()[0].String())
		_ = b.WriteByte(',')
		t = c.Arguments()[1]
	}
	return string(b.Bytes()[:b.Len()-1]) + "]"
}

func PrettyString(t Term) string {
	return "\"" + strings.Replace(RawString(t), "\"", "\\\"", -1) + "\""
}

func RawString(t Term) string {
	var chars []rune
	for !IsEmptyList(t) {
		c := t.(*Compound)
		code := c.Arguments()[0].(*Integer).Code()
		chars = append(chars, code)
		t = c.Arguments()[1]
	}
	return string(chars)
}

func ListToSlice(t Term) (result []Term) {
	for !IsEmptyList(t) {
		c := t.(Callable)
		result = append(result, c.Arguments()[0])
		t = c.Arguments()[1]
	}
	return result
}

func SliceToList(ts []Term) (result Term) {
	result = NewAtom("[]")
	if len(ts) == 0 {
		return result
	}
	result = NewCallable(".", ts[len(ts)-1], result)
	for i := len(ts) - 2; i >= 0; i-- {
		result = NewCallable(".", ts[i], result)
	}
	return result
}
