package native

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/mndrix/golog"
	"github.com/mndrix/golog/term"
)

type AllBools struct {
	FirstField   bool
	SECond_Field bool
	ThirdFi31d   bool
}

type RecStruct struct {
	FirstField bool
	Sub1       AllBools
	Sub2       AllBools
}

type IntsStrings struct {
	Int    int
	String string
}

type List struct {
	List []AllBools
}

type MixedSimple struct {
	F1 float64
	F2 float32
	I  int
	I1 int8
	I2 int16
	I3 int32
	I4 int64
	U  uint
	U1 uint8
	U2 uint16
	U3 uint32
	U4 uint64
}

type BehindInterface interface {
	Foo()
}

func (ab *AllBools) Foo() {}

type FakeInt int

func (fi FakeInt) Foo() {}

type InterfaceWrapper struct {
	Wrapped BehindInterface
}

type MarshalledBool bool

func (b *MarshalledBool) UnmarshalProlog(cache map[uintptr]term.Term) term.Term {
	if b == nil || !(*b) {
		return term.NewAtom("false")
	}
	return term.NewAtom("true")
}

func (b *MarshalledBool) MarshalProlog(m golog.Machine, t term.Term) error {
	if !term.IsAtom(t) {
		return fmt.Errorf("expected %s to be an atom", t)
	}
	n := t.(*term.Atom).Name()
	if n != "true" || n != "false" {
		return fmt.Errorf("%T can only be true or false", b)
	}
	*b = (n == "true")
	return nil
}

type MarshalWrapper struct {
	MB MarshalledBool
}

func normalizeOutput(s string) string {
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(strings.TrimSpace(s), "")
}

func TestAllBools(t *testing.T) {
	ab := AllBools{
		FirstField:   true,
		SECond_Field: false,
		ThirdFi31d:   true,
	}
	expected := normalizeOutput(`
all_bools([
  first_field(yes), 
  second_field(no), 
  third_fi31d(yes)])
`)
	encoder := NewEncoder()
	comp := normalizeOutput(encoder.Encode(ab).String())
	if comp != expected {
		t.Fatalf("Expected: %s, actual: %s", expected, comp)
	}
}

func TestRecBools(t *testing.T) {
	ab := RecStruct{
		FirstField: true,
		Sub1: AllBools{
			FirstField:   true,
			SECond_Field: false,
			ThirdFi31d:   true,
		},
		Sub2: AllBools{
			FirstField:   false,
			SECond_Field: true,
			ThirdFi31d:   false,
		},
	}
	expected := normalizeOutput(`
rec_struct([
  first_field(yes),
  sub1(
    all_bools([
      first_field(yes), 
      second_field(no), 
      third_fi31d(yes)])),
  sub2(
    all_bools([
      first_field(no), 
      second_field(yes), 
      third_fi31d(no)]))])
`)
	encoder := NewEncoder()
	comp := normalizeOutput(encoder.Encode(ab).String())
	if comp != expected {
		t.Fatalf("Expected: %s, actual: %s", expected, comp)
	}
}

func TestIntsStrings(t *testing.T) {
	ab := IntsStrings{
		Int:    123456,
		String: "foo",
	}
	expected := normalizeOutput(`
ints_strings([
  int(123456), 
  string("foo")])
`)
	encoder := NewEncoder()
	comp := normalizeOutput(encoder.Encode(ab).String())
	if comp != expected {
		t.Fatalf("Expected: %s, actual: %s", expected, comp)
	}
}

func TestList(t *testing.T) {
	ab := List{
		List: []AllBools{
			AllBools{
				FirstField:   true,
				SECond_Field: false,
				ThirdFi31d:   true,
			},
			AllBools{
				FirstField:   false,
				SECond_Field: true,
				ThirdFi31d:   false,
			},
		},
	}
	expected := normalizeOutput(`
list([
  list([
    all_bools([
      first_field(yes),
      second_field(no),
      third_fi31d(yes)]),
    all_bools([
      first_field(no),
      second_field(yes),
      third_fi31d(no)])])])
`)
	encoder := NewEncoder()
	comp := normalizeOutput(encoder.Encode(ab).String())
	if comp != expected {
		t.Fatalf("Expected: %s, actual: %s", expected, comp)
	}
}

func TestMixed(t *testing.T) {
	ab := MixedSimple{
		F1: 0.12345,
		F2: 1,
		I:  -12345,
		I1: -123,
		I2: -1234,
		I3: -12345,
		I4: -123456,
		U:  12345,
		U1: 123,
		U2: 1234,
		U3: 12345,
		U4: 123456,
	}
	expected := normalizeOutput(`
mixed_simple([
  f1(0.12345),
  f2(1),
  i(-12345),
  i1(-123),
  i2(-1234),
  i3(-12345),
  i4(-123456),
  u(12345),
  u1(123),
  u2(1234),
  u3(12345),
  u4(123456)])
`)
	encoder := NewEncoder()
	comp := normalizeOutput(encoder.Encode(ab).String())
	if comp != expected {
		t.Fatalf("Expected: %s, actual: %s", expected, comp)
	}
}

func TestInterface(t *testing.T) {
	ab := InterfaceWrapper{
		Wrapped: FakeInt(42),
	}
	expected := normalizeOutput(`interface_wrapper([wrapped(42)])`)
	encoder := NewEncoder()
	comp := normalizeOutput(encoder.Encode(ab).String())
	if comp != expected {
		t.Fatalf("Expected: %s, actual: %s", expected, comp)
	}
	ab = InterfaceWrapper{
		Wrapped: &AllBools{},
	}
	expected = normalizeOutput(`
interface_wrapper([
  wrapped(
    all_bools([
      first_field(no),
      second_field(no),
      third_fi31d(no)]))])`)
	comp = normalizeOutput(encoder.Encode(ab).String())
	if comp != expected {
		t.Fatalf("Expected: %s, actual: %s", expected, comp)
	}
}

func TestMarshalledSimple(t *testing.T) {
	ab := MarshalWrapper{
		MB: true,
	}

	expected := normalizeOutput(`marshal_wrapper([mb(true)])`)
	encoder := NewEncoder()
	comp := normalizeOutput(encoder.Encode(ab).String())
	if comp != expected {
		t.Fatalf("Expected: %s, actual: %s", expected, comp)
	}
}
