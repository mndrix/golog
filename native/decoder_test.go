package native

import (
	"testing"

	"github.com/mndrix/golog"
	"github.com/mndrix/golog/read"
	"github.com/mndrix/golog/term"
)

type MarshallerImpl struct {
	Value int
}

type WrapMarshaller struct {
	Marshaller *MarshallerImpl
}

func (mi *MarshallerImpl) MarshalProlog(m golog.Machine, t term.Term) error {
	*mi = MarshallerImpl{}
	d := NewDecoder(m)
	var val int
	if err := d.Decode(t, &val); err != nil {
		return err
	}
	mi.Value = val
	return nil
}

func TestDecodeAllBools(t *testing.T) {
	var ab AllBools
	decoder := NewDecoder(golog.NewMachine())
	expected := normalizeOutput(`
all_bools(
  [first_field(yes), 
   second_field(no), 
   third_fi31d(yes)]).
`)
	input := read.TermAll_(expected)[0]
	if err := decoder.Decode(input, &ab); err != nil {
		t.Fatalf("Decoding failed: %s", err)
	}
	if !ab.FirstField {
		t.Fatalf("Expected ab.FirstField: to be true")
	}
}

func TestDecodeSlice(t *testing.T) {
	var ab List
	decoder := NewDecoder(golog.NewMachine())
	expected := normalizeOutput(`
list(
  [list([
      all_bools([
        first_field(yes),
        second_field(no),
        third_fi31d(yes)]),
      all_bools([
        first_field(no),
        second_field(yes),
        third_fi31d(no)])])]).
`)
	input := read.TermAll_(expected)[0]
	if err := decoder.Decode(input, &ab); err != nil {
		t.Fatalf("Decoding failed: %s", err)
	}
	if !ab.List[1].SECond_Field {
		t.Fatalf("Expected ab.FirstField: to be true")
	}
}

func TestDecodeMarshaller(t *testing.T) {
	var ab WrapMarshaller
	decoder := NewDecoder(golog.NewMachine())
	expected := normalizeOutput(`wrap_marshaller([marshaller(42)]).`)
	input := read.TermAll_(expected)[0]
	if err := decoder.Decode(input, &ab); err != nil {
		t.Fatalf("Decoding failed: %s", err)
	}
	if ab.Marshaller == nil {
		t.Fatalf("Expected ab.Marshaller to exist")
	}
	if ab.Marshaller.Value != 42 {
		t.Fatalf("Expected ab.Marshaller.Value = 42")
	}
}
