// Represent and unify Prolog terms.  Along with golog.Machine, term.Term
// is one of the most important data types in Golog.  It provides a Go
// representation of Prolog terms.  Terms represent Prolog code,
// Prolog queries and Prolog results.
//
// The current term API is messy and will definitely change in the future.
package term

import . "fmt"
import . "regexp"
import "math/big"
import "math"
import "strconv"
import "strings"
import "github.com/mndrix/golog/lex"
import "github.com/mndrix/ps"

// Returned by Unify() if the unification fails
var CantUnify error = Errorf("Can't unify the given terms")

// Possible term types, in order according to ISO §7.2
const (
	VariableType = iota
	FloatType
	IntegerType
	AtomType
	CompoundType

	// odd man out
	ErrorType
)

// Term represents a single Prolog term which might be an atom, a
// compound structure, an integer, etc.  Many methods on Term will
// be replaced with functions in the future.  The Term interface is
// also likely to be split into several smaller interfaces like Atomic,
// Number, etc.
type Term interface {
	// ReplaceVariables replaces any internal variables with the values
	// to which they're bound.  Unbound variables are left as they are
	ReplaceVariables(Bindings) Term

	// String provides a string representation of a term
	String() string

	// Type indicates whether this term is an atom, number, compound, etc.
	// ISO §7.2 uses the word "type" to descsribe this idea.  Constants are
	// defined for each type.
	Type() int

	// Indicator() provides a "predicate indicator" representation of a term
	Indicator() string

	// Unifies the invocant and another term in the presence of an
	// environment.
	// On succes, returns a new environment with additional variable
	// bindings.  On failure, returns CantUnify error along with the
	// original environment
	Unify(Bindings, Term) (Bindings, error)
}

// Callable represents either an atom or a compound term.  This is the
// terminology used by callable/1 in many Prologs.
type Callable interface {
	Term

	// Name returns the term's name.  Some people might call this the term's
	// functor, but there's ambiguity surrounding that word in the Prolog
	// community (some use it for Name/Arity pairs).
	Name() string

	// Arity returns the number of arguments a term has. An atom has 0 arity.
	Arity() int

	// Arguments returns a slice of this term's arguments, if any
	Arguments() []Term
}

// Returns true if term t is an atom
func IsAtom(t Term) bool {
	return t.Type() == AtomType
}

// IsClause returns true if the term is like 'Head :- Body', otherwise false
func IsClause(t Term) bool {
	switch t.Type() {
	case CompoundType:
		x := t.(*Compound)
		return x.Arity() == 2 && x.Name() == ":-"
	case AtomType,
		VariableType,
		IntegerType,
		FloatType,
		ErrorType:
		return false
	}
	msg := Sprintf("Unexpected term type: %#v", t)
	panic(msg)
}

// Returns true if term t is a compound term.
func IsCompound(t Term) bool {
	return t.Type() == CompoundType
}

// Returns true if term t is an atom or compound term.
func IsCallable(t Term) bool {
	tp := t.Type()
	return tp == AtomType || tp == CompoundType
}

// Returns true if term t is a variable.
func IsVariable(t Term) bool {
	return t.Type() == VariableType
}

// Returns true if term t is an error term.
func IsError(t Term) bool {
	return t.Type() == ErrorType
}

// Returns true if term t is a directive like `:- foo.`
func IsDirective(t Term) bool {
	return t.Indicator() == ":-/1"
}

// Head returns a term's first argument. Panics if there isn't one
func Head(t Term) Callable {
	return t.(*Compound).Arguments()[0].(Callable)
}

// Body returns a term's second argument. Panics if there isn't one
func Body(t Term) Callable {
	return t.(*Compound).Arguments()[1].(Callable)
}

// RenameVariables returns a new term like t with all variables replaced
// by fresh ones.
func RenameVariables(t Term) Term {
	renamed := make(map[string]*Variable)
	return renameVariables(t, renamed)
}

func renameVariables(t Term, renamed map[string]*Variable) Term {
	switch t.Type() {
	case FloatType,
		IntegerType,
		AtomType,
		ErrorType:
		return t
	case CompoundType:
		x := t.(*Compound)
		newArgs := make([]Term, x.Arity())
		for i, arg := range x.Arguments() {
			newArgs[i] = renameVariables(arg, renamed)
		}
		newTerm := NewCallable(x.Name(), newArgs...)
		newTerm.(*Compound).ucache = x.ucache
		return newTerm
	case VariableType:
		x := t.(*Variable)
		name := x.Name
		if name == "_" {
			name = x.Indicator()
		}
		v, ok := renamed[name]
		if ok {
			return v
		} else {
			v = x.WithNewId()
			renamed[name] = v
			return v
		}
	}
	panic("Unexpected term type")
}

// Variables returns a ps.Map whose keys are human-readable variable names
// and those values are *Variable used inside term t.
func Variables(t Term) ps.Map {
	names := ps.NewMap()
	switch t.Type() {
	case AtomType,
		FloatType,
		IntegerType,
		ErrorType:
		return names
	case CompoundType:
		x := t.(*Compound)
		if x.Arity() == 0 {
			return names
		} // no variables in an atom
		for _, arg := range x.Arguments() {
			innerNames := Variables(arg)
			innerNames.ForEach(func(key string, val interface{}) {
				names = names.Set(key, val)
			})
		}
		return names
	case VariableType:
		x := t.(*Variable)
		return names.Set(x.Name, x)
	}
	panic("Unexpected term implementation")
}

// QuoteFunctor returns a canonical representation of a term's name
// by quoting characters that require quoting
func QuoteFunctor(name string) string {
	// cons must be quoted (to avoid confusion with full stop)
	if name == "." || name == "" {
		return Sprintf("'%s'", name)
	}

	// names composed entirely of graphic characters need no quoting
	allGraphic := true
	for _, c := range name {
		if !lex.IsGraphic(c) {
			allGraphic = false
			break
		}
	}
	if allGraphic || name == "[]" || name == "!" || name == ";" {
		return name
	}

	nonAlpha, err := MatchString(`\W`, name)
	maybePanic(err)
	nonLower, err := MatchString(`^[^a-z]`, name)
	if nonAlpha || nonLower {
		escapedName := strings.Replace(name, `'`, `\'`, -1)
		return Sprintf("'%s'", escapedName)
	}

	return name
}

// NewCodeList constructs a list of character codes from a string.
// The string should include opening and closing " characters.
// Nominally, the resulting term is just a chain of cons cells ('.'/2),
// but it might actually be a more efficient implementation under the hood.
func NewCodeListFromDoubleQuotedString(s string) Term {
	// make sure the content is long enough
	runes := []rune(s)
	end := len(runes) - 2
	if end < 0 {
		msg := Sprintf("Code list string must have bracketing double quotes: %s", s)
		panic(msg)
	}

	// build a cons cell chain, starting at the end ([])
	codes := NewAtom(`[]`)
	for i := end; i > 0; i-- {
		c := NewCode(runes[i])
		codes = NewCallable(".", c, codes)
	}

	return codes
}

// Precedes returns true if the first argument 'term-precedes'
// the second argument according to ISO §7.2
func Precedes(a, b Term) bool {
	aP := precedence(a)
	bP := precedence(b)
	if aP < bP {
		return true
	}
	if aP > bP {
		return false
	}

	// both terms have the same precedence by type, so delve deeper
	switch a.Type() {
	case VariableType:
		x := a.(*Variable)
		y := b.(*Variable)
		return x.Id() < y.Id()
	case FloatType,
		IntegerType: // See Note_1

		// comparing via float64 breaks in many, many ways.
		// improve as necessary.
		x := a.(Number).Float64()
		y := b.(Number).Float64()
		if x == y && IsFloat(a) && !IsFloat(b) {
			return true
		}
		return x < y
	case AtomType:
		x := a.(*Atom)
		y := b.(*Atom)
		return x.Name() < y.Name()
	case CompoundType:
		x := a.(*Compound)
		y := b.(*Compound)
		if x.Arity() < y.Arity() {
			return true
		}
		if x.Arity() > y.Arity() {
			return false
		}
		if x.Name() < y.Name() {
			return true
		}
		if x.Name() > y.Name() {
			return false
		}
		for i := 0; i < x.Arity(); i++ {
			if Precedes(x.Arguments()[i], y.Arguments()[i]) {
				return true
			} else if Precedes(y.Arguments()[i], x.Arguments()[i]) {
				return false
			}
		}
		return false // identical terms
	}

	msg := Sprintf("Unexpected term type %s\n", a)
	panic(msg)
}
func precedence(t Term) int {
	value := t.Type()       // Type() promises values in precedence order
	if value == FloatType { // See Note_1
		return IntegerType
	}
	return value
}

// Note_1:
//
// I've chosen to willfully violate the ISO standard in §7.2 because it
// mandates that floats precede all integers.  That means
// `42.3 @< 9` which isn't helpful.  I don't deviate lightly, but strongly
// believe it's the right way.
// Incidentally, SWI-Prolog behaves this way by default.

// UnificationHash generates a special hash value representing the
// terms in a slice.  Golog uses these hashes to optimize
// unification.  You probably don't need to call this function directly.
//
// In more detail, UnificationHash generates a 64-bit hash which
// represents the shape and content of a term.  If two terms share the same
// hash, those terms are likely to unify, although not guaranteed.  If
// two terms have different hashes, the two terms are guaranteed not
// to unify.  A compound term splits its 64-bit hash into multiple, smaller
// n-bit hashes for its functor and arguments.  Other terms occupy the entire
// hash space themselves.
//
// Variables require special handling.  During "preparation" we can think of
// 1-bits as representing what content a term "provides".  During "query" we
// can think of 1-bits as representing what content a term "requires".
// In the first phase, a variable hashes to all 1s since it can provide
// whatever is needed.  In the second phase, a variable hashes to all 0s since
// it demands nothing of the opposing term.
var bigMaxInt64 *big.Int

func init() {
	bigMaxInt64 = big.NewInt(math.MaxInt64)
}
func UnificationHash(terms []Term, n uint, preparation bool) uint64 {
	var hash uint64 = 0
	var blockSize uint = n / uint(len(terms))

	// mask to select blockSize least significant bits
	var mask uint64
	if blockSize == 64 {
		mask = math.MaxUint64
	} else if blockSize == 0 {
		// pretend that terms was a single variable
		if preparation {
			return (1 << n) - 1
		} else {
			return 0
		}
	} else {
		mask = (1 << blockSize) - 1
	}

	for _, term := range terms {
		hash = hash << blockSize
		switch t := term.(type) {
		case *Atom:
			hash = hash | (hashString(t.Name()) & mask)
		case *Integer:
			if t.Value().Sign() < 0 || t.Value().Cmp(bigMaxInt64) > 0 {
				str := Sprintf("%x", t.Value())
				hash = hash | (hashString(str) & mask)
			} else {
				hash = hash | (uint64(t.Value().Int64()) & mask)
			}
		case *Rational:
			str := t.String()
			hash = hash | (hashString(str) & mask)
		case *Float:
			str := strconv.FormatFloat(t.Value(), 'b', 0, 64)
			hash = hash | (hashString(str) & mask)
		case *Error:
			panic("No UnificationHash for Error terms")
		case *Compound:
			var termHash uint64
			arity := uint(t.Arity())
			if arity == 2 && t.Name() == "." { // don't hash pair's functor
				rightSize := blockSize / 2
				leftSize := blockSize - rightSize
				termHash = UnificationHash(t.Args[0:1], leftSize, preparation)
				termHash = termHash << rightSize
				termHash |= UnificationHash(t.Args[1:2], rightSize, preparation)
			} else {
				// how many bits allocated to functor vs arguments?
				functorBits := blockSize / (arity + 1)
				if functorBits > 12 {
					functorBits = 12
				}
				argumentBits := (blockSize - functorBits) / arity

				// give extra bits (from rounding) back to the functor
				functorBits = blockSize - argumentBits*arity

				// generate the hash
				var functorMask uint64 = (1 << functorBits) - 1
				termHash = hashString(t.Name()) & functorMask
				for _, arg := range t.Arguments() {
					termHash = termHash << argumentBits
					termHash = termHash | UnificationHash([]Term{arg}, argumentBits, preparation)
				}
			}
			hash |= (termHash & mask)
		case *Variable:
			if preparation {
				hash = hash | mask
			}
		case Stringer:
			hash = hash | (hashString(t.String()) & mask)
		default:
			msg := Sprintf("Unexpected term type %s\n", t)
			panic(msg)
		}
	}

	return hash
}

// constants for FNV-1a hash algorithm
const (
	offset64 uint64 = 14695981039346656037
	prime64  uint64 = 1099511628211
)

// hashString returns a hash code for a given string
func hashString(x string) uint64 {
	hash := offset64
	for _, codepoint := range x {
		hash ^= uint64(codepoint)
		hash *= prime64
	}
	return hash
}

// Converts a '.'/2 list terminated in []/0 into a slice of the associated
// terms.  Panics if the argument is not a proper list.
func ProperListToTermSlice(x Term) []Term {
	l := make([]Term, 0)
	t := x.(Callable)
	for {
		switch t.Indicator() {
		case "[]/0":
			return l
		case "./2":
			l = append(l, t.Arguments()[0])
			t = t.Arguments()[1].(Callable)
		default:
			panic("Improper list")
		}
	}
}

// Implement sort.Interface for []Term
type TermSlice []Term

func (self *TermSlice) Len() int {
	ts := []Term(*self)
	return len(ts)
}
func (self *TermSlice) Less(i, j int) bool {
	ts := []Term(*self)
	return Precedes(ts[i], ts[j])
}
func (self *TermSlice) Swap(i, j int) {
	ts := []Term(*self)
	tmp := ts[i]
	ts[i] = ts[j]
	ts[j] = tmp
}

func maybePanic(err error) {
	if err != nil {
		panic(err)
	}
}
