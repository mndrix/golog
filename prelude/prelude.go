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
		Ignore1,
		Length2,
		Memberchk2,
		Phrase2,
		Phrase3,
		Sort2,
	}, "\n\n")
}

var Ignore1 = `
ignore(A) :-
	call(A),
	!.
ignore(_).
`

var Length2 = `
length(Xs, N) :-
	length(Xs, 0, N).

length([], N, N) :- !.
length([_|T], N0, N) :-
	% N0 \= N
	succ(N0, N1),
	length(T, N1, N).
`

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

// sort(+List, -Sorted) is det.
//
// Like msort/2 but removes duplicates.
var Sort2 = `
sort(List, Sorted) :-
	msort(List, Duplicates),
	'$consolidate'(Duplicates, Sorted).

% helper predicate that removes adjacent duplicates from a list
'$consolidate'([], []).
'$consolidate'([X], [X]).
'$consolidate'([X,Y|Rest], Result) :-
	( X = Y ->
		'$consolidate'([Y|Rest], Result)
	; % otherwise ->
		'$consolidate'([Y|Rest], Tail),
		Result = [X|Tail]
	).
`
