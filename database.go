package golog

import . "fmt"
import . "github.com/mndrix/golog/term"

import "bytes"
import "github.com/mndrix/ps"

// Database is an immutable Prolog database.  All write operations on the
// database produce a new database without affecting the previous one.
type Database interface {
    // Asserta adds a term to the database at the start of any existing
    // terms with the same name and arity.
    Asserta(Term) Database

    // Assertz adds a term to the database at the end of any existing
    // terms with the same name and arity.
    Assertz(Term) Database

    // Candidates() returns a list of clauses that might match a term
    Candidates(Term) []Term

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
    clauseCount int     // number of clauses in the database
    predicates  *ps.Map  // term indicator => *clauses
}

func (self *mapDb) Asserta(term Term) Database {
    return self.assert('a', term)
}

func (self *mapDb) Assertz(term Term) Database {
    return self.assert('z', term)
}

func (self *mapDb) assert(side rune, term Term) Database {
    var newMapDb    mapDb
    var cs  *clauses

    // find the indicator under which this term is classified
    indicator := term.Indicator()
    if term.IsClause() {
        // ':-' uses the indicator of its head term
        indicator = term.Head().Indicator()
    }

    oldClauses, ok := self.predicates.Lookup(indicator)
    if ok {  // clauses exist for this predicate
        switch side {
            case 'a':
                cs = oldClauses.(*clauses).cons(term)
            case 'z':
                cs = oldClauses.(*clauses).snoc(term)
        }
    } else {  // brand new predicate
        cs = newClauses().snoc(term)
    }

    newMapDb.clauseCount = self.clauseCount + 1
    newMapDb.predicates = self.predicates.Set(indicator, cs)
    return &newMapDb
}

func (self *mapDb) Candidates(t Term) []Term {
    indicator := t.Indicator()
    cs, ok := self.predicates.Lookup(indicator)
    if !ok {  // no clauses for this predicate
        terms := make([]Term, 0)
        return terms
    }

    return cs.(*clauses).all()
}

func (self *mapDb) ClauseCount() int {
    return self.clauseCount
}

func (self *mapDb) String() string {
    var buf bytes.Buffer

    keys := self.predicates.Keys()
    for _, key := range keys {
        cs, _ := self.predicates.Lookup(key)
        cs.(*clauses).forEach( func (t Term) { Fprintf(&buf, "%s.\n", t) } )
    }
    return buf.String()
}
