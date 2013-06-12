// Defines the Prolog standard library. All prelude predicates are
// implemented in pure Prolog.  Each variable in this package is a predicate
// definition.  At init time, they're combined into a single string
// in the Prelude var.
package prelude

import "strings"

// After init(), Prelude contains all prelude predicates combined into
// a single large string.  One rarely addresses this variable directly
// because golog.NewMachine() does it for you.
var Prelude string

func init() {
	Prelude = strings.Join([]string{
		Memberchk2,
		Phrase2,
		Phrase3,
	}, "\n\n")
}

var Memberchk2 = `
memberchk(X,[X|_]) :- !.
memberchk(X,[_|T]) :-
    memberchk(X,T).
`

// phrase(:DCGBody, ?List, ?Rest) is nondet.
//
// True when DCGBody applies to the difference List/Rest.
var Phrase3 = `
phrase(Dcg, Head, Tail) :-
    call(Dcg, Head, Tail).
`

// phrase(:DCGBody, ?List) is nondet.
//
// Like phrase(DCG,List,[]).
var Phrase2 = `
phrase(Dcg, List) :-
    call(Dcg, List, []).
`
