package golog

import . "fmt"

import "bytes"
import "github.com/mndrix/ps"

// Database is an immutable Prolog database.  All write operations on the
// database produce a new database without affecting the previous one.
type Database interface {
    // Asserta adds a term to the database at the start of any existing
    // terms with the same name and arity.
    Asserta(Term) Database

    // ClauseCount returns the number of clauses in the database
    ClauseCount() int

    // Prove determines whether a given term is true for this database
    Prove(Term) bool

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
    predicates  *ps.Map  // term indicator => ps.List of terms
}
func (self *mapDb) Asserta(term Term) Database {
    var newMapDb    mapDb
    var newClauses  *ps.List

    indicator := term.Indicator()
    oldClauses, ok := self.predicates.Lookup(indicator)
    if ok {  // clauses exist for this predicate
        newClauses = oldClauses.(*ps.List).Cons(term)
    } else {  // brand new predicate
        newClauses = ps.NewList().Cons(term)
    }

    newMapDb.clauseCount = self.clauseCount + 1
    newMapDb.predicates = self.predicates.Set(indicator, newClauses)
    return &newMapDb
}
func (self *mapDb) ClauseCount() int {
    return self.clauseCount
}
func (self *mapDb) Prove(term Term) bool {
    // TODO
    return true
}
func (self *mapDb) String() string {
    var buf bytes.Buffer

    keys := self.predicates.Keys()
    for _, key := range keys {
        clauses, _ := self.predicates.Lookup(key)
        clauses.(*ps.List).ForEach( func (v ps.Any) { Fprintf(&buf, "%s.\n", v) } )
    }
    return buf.String()
}
