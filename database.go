package golog

import . "fmt"
import . "github.com/mndrix/golog/term"
import . "github.com/mndrix/golog/util"

import "bytes"
import "github.com/mndrix/ps"

// Database is an immutable Prolog database.  All write operations on the
// database produce a new database without affecting the previous one.
// A database is a mapping from predicate indicators (foo/3) to clauses.
// The database may or may not implement indexing.  It's unusual to
// interact with databases directly.  One usually calls methods on Machine
// instead.
type Database interface {
	// Asserta adds a term to the database at the start of any existing
	// terms with the same name and arity.
	Asserta(Term) Database

	// Assertz adds a term to the database at the end of any existing
	// terms with the same name and arity.
	Assertz(Term) Database

	// Candidates() returns a list of clauses that might unify with a term.
	// Returns error if no predicate with appropriate
	// name and arity has been defined.
	Candidates(Term) ([]Term, error)

	// Candidates_() is like Candidates() but panics on error.
	Candidates_(Term) []Term

	// ClauseCount returns the number of clauses in the database
	ClauseCount() int

	// String returns a string representation of the entire database
	String() string
}

// NewDatabase returns a new, empty database
func NewDatabase() Database {
	var db mapDb
	db.clauseCount = 0
	db.predicates = ps.NewMap()
	return &db
}

type mapDb struct {
	clauseCount int    // number of clauses in the database
	predicates  ps.Map // term indicator => *clauses
}

func (self *mapDb) Asserta(term Term) Database {
	return self.assert('a', term)
}

func (self *mapDb) Assertz(term Term) Database {
	return self.assert('z', term)
}

func (self *mapDb) assert(side rune, term Term) Database {
	var newMapDb mapDb
	var cs *clauses

	// find the indicator under which this term is classified
	indicator := term.Indicator()
	if term.IsClause() {
		// ':-' uses the indicator of its head term
		indicator = term.Head().Indicator()
	}

	oldClauses, ok := self.predicates.Lookup(indicator)
	if ok { // clauses exist for this predicate
		switch side {
		case 'a':
			cs = oldClauses.(*clauses).cons(term)
		case 'z':
			cs = oldClauses.(*clauses).snoc(term)
		}
	} else { // brand new predicate
		cs = newClauses().snoc(term)
	}

	newMapDb.clauseCount = self.clauseCount + 1
	newMapDb.predicates = self.predicates.Set(indicator, cs)
	return &newMapDb
}

func (self *mapDb) Candidates_(t Term) []Term {
	ts, err := self.Candidates(t)
	if err != nil {
		panic(err)
	}
	return ts
}

func (self *mapDb) Candidates(t Term) ([]Term, error) {
	indicator := t.Indicator()
	cs, ok := self.predicates.Lookup(indicator)
	if !ok { // this predicate hasn't been defined
		return nil, Errorf("Undefined predicate: %s", indicator)
	}

	// quick return for an atom term
	if !IsCompound(t) {
		return cs.(*clauses).all(), nil
	}

	// ignore clauses that can't possibly unify with our term
	candidates := make([]Term, 0)
	cs.(*clauses).forEach(func(clause Term) {
		if !IsCompound(clause) {
			Debugf("    ... discarding. Not compound term\n")
			return
		}
		head := clause
		if clause.IsClause() {
			head = clause.Head()
		}
		if t.(*Compound).MightUnify(head.(*Compound)) {
			Debugf("    ... adding to candidates: %s\n", clause)
			candidates = append(candidates, clause)
		}
	})
	Debugf("  final candidates = %s\n", candidates)
	return candidates, nil
}

func (self *mapDb) ClauseCount() int {
	return self.clauseCount
}

func (self *mapDb) String() string {
	var buf bytes.Buffer

	keys := self.predicates.Keys()
	for _, key := range keys {
		cs, _ := self.predicates.Lookup(key)
		cs.(*clauses).forEach(func(t Term) { Fprintf(&buf, "%s.\n", t) })
	}
	return buf.String()
}
