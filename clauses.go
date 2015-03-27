package golog

import (
	"strconv"

	"github.com/mndrix/golog/term"
	"github.com/mndrix/ps"
)

// clauses represents an ordered list of terms with
// cheap insertion at the front
// and back and deletion from anywhere.
// Each clause has a unique identifier
// which can be used for deletions
type clauses struct {
	n         int64 // number of terms in collection
	lowestId  int64
	highestId int64
	terms     ps.Map // maps int64 => Term
}

// newClauses returns a new, empty list of clauses
func newClauses() *clauses {
	var cs clauses
	cs.terms = ps.NewMap()
	// n, lowestId, highestId correctly default to 0
	return &cs
}

// returns the number of terms stored in the list
func (self *clauses) count() int64 {
	return self.n
}

// cons adds a term to the list's front
func (self *clauses) cons(t term.Term) *clauses {
	cs := self.clone()
	cs.n++
	cs.lowestId--
	key := strconv.FormatInt(cs.lowestId, 10)
	cs.terms = self.terms.Set(key, t)
	return cs
}

// cons adds a term to the list's back
func (self *clauses) snoc(t term.Term) *clauses {
	cs := self.clone()
	cs.n++
	cs.highestId++
	key := strconv.FormatInt(cs.highestId, 10)
	cs.terms = self.terms.Set(key, t)
	return cs
}

// all returns a slice of all terms, in order
func (self *clauses) all() []term.Term {
	terms := make([]term.Term, 0)
	if self.count() == 0 {
		return terms
	}

	for i := self.lowestId; i <= self.highestId; i++ {
		key := strconv.FormatInt(i, 10)
		t, ok := self.terms.Lookup(key)
		if ok {
			terms = append(terms, t.(term.Term))
		}
	}
	return terms
}

// invoke a callback on each clause
func (self *clauses) forEach(f func(term.Term)) {
	for _, t := range self.all() {
		f(t)
	}
}

// returns a copy of this clause list
func (self *clauses) clone() *clauses {
	cs := *self
	return &cs
}
