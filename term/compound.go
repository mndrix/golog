package term

import . "fmt"

import "bytes"

// NewTerm creates a new term with the given functor and optional arguments
func NewTerm(functor string, arguments ...Term) Term {
	if len(arguments) == 0 {
		return NewAtom(functor)
	}
	return &Compound{
		Func:   functor,
		Args:   arguments,
		ucache: &unificationCache{},
	}
}

// Unlikely to be useful outside of the parser
func NewTermFromLexeme(possiblyQuotedName string, arguments ...Term) Term {
	a := NewAtomFromLexeme(possiblyQuotedName)
	return NewTerm(a.Functor(), arguments...)
}

// NewCodeList returns a compound term consisting of the character codes
// of the given string.  The internal representation may eventually optimize
// for storing character codes.
func NewCodeList(s string) Term {
	runes := []rune(s)
	list := NewAtom("[]")
	for i := len(runes) - 1; i >= 0; i-- {
		list = NewTerm(".", NewCode(runes[i]), list)
	}
	return list
}

// NewTermList returns a list term consisting of each term from the slice.
// A future implementation may optimize the data structure that's returned.
func NewTermList(terms []Term) Term {
	list := NewAtom("[]")
	for i := len(terms) - 1; i >= 0; i-- {
		list = NewTerm(".", terms[i], list)
	}
	return list
}

// ISO calls this a "compound term" see §6.1.2(e)
// We currently use this type to cover atoms defined in §6.1.2(b)
// by treating atoms as compound terms with 0 arity.
type Compound struct {
	Func   string
	Args   []Term
	ucache *unificationCache
}
type unificationCache struct {
	// 0 means UnificationHash hasn't been calculated yet
	phash uint64 // prepared hash
	qhash uint64 // query hash
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
func (self *Compound) IsClause() bool {
	return self.Arity() == 2 && self.Functor() == ":-"
}
func (self *Compound) String() string {
	quotedFunctor := QuoteFunctor(self.Functor())

	var buf bytes.Buffer
	Fprintf(&buf, "%s(", quotedFunctor)
	arity := self.Arity()
	for i := 0; i < arity; i++ {
		if i > 0 {
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

func (self *Compound) ReplaceVariables(env Bindings) Term {
	args := self.Arguments()
	for i, arg := range args {
		newArg := arg.ReplaceVariables(env)
		if arg != newArg { // argument changed. build a new compound term
			newArgs := make([]Term, self.Arity())
			for j, arg := range args {
				if j < i {
					newArgs[j] = arg
				} else {
					if j == i {
						newArgs[j] = newArg
					} else {
						newArgs[j] = arg.ReplaceVariables(env)
					}
				}
			}
			newTerm := NewTerm(self.Functor(), newArgs...)
			return newTerm
		}
	}

	// no variables were replaced.  reuse the same compound term
	return self
}

func (a *Compound) Unify(e Bindings, b Term) (Bindings, error) {
	if IsVariable(b) {
		return b.Unify(e, a)
	}
	if !IsCompound(b) {
		return e, CantUnify
	}

	// functor and arity must match for unification to work
	arity := a.Arity()
	if arity != b.Arity() {
		return e, CantUnify
	}
	if a.Functor() != b.Functor() {
		return e, CantUnify
	}

	// try unifying each subterm
	var err error
	env := e
	aArgs := a.Arguments()
	bArgs := b.Arguments()
	for i := 0; i < arity; i++ {
		env, err = aArgs[i].Unify(env, bArgs[i])
		if err != nil {
			return e, err // return original environment along with error
		}
	}

	// unification succeeded
	return env, nil
}

// Univ is just like =../2 in ISO Prolog
func (self *Compound) Univ() []Term {
	ts := make([]Term, 0)
	ts = append(ts, NewAtom(self.Functor()))
	ts = append(ts, self.Arguments()...)
	return ts
}

// Returns true if a and b might unify.  This is an optimization
// for times when a and b are frequently unified with other
// compound terms.  For example, goals and clause heads.
func (a *Compound) MightUnify(b *Compound) bool {
	if a.ucache.qhash == 0 {
		a.ucache.qhash = UnificationHash([]Term{a}, 64, false)
	}
	if b.ucache.phash == 0 {
		b.ucache.phash = UnificationHash([]Term{b}, 64, true)
	}

	return (a.ucache.qhash & b.ucache.phash) == a.ucache.qhash
}
