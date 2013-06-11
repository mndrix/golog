% Tests for (\+)/1
:- use_module(library(tap)).

% Tests derived from Prolog: The Standard p. 115
simple_unify :-
    X = 3,
    \+((X=1;X=2)).

failing :-
    \+ fail.

/*
% should bind X to 1
cutting :-
    \+(!); X=1.
*/

disjunction_then_unify(fail) :-
    \+((X=1;X=2)), 3=X.

unify_then_disjunction(fail) :-
    X = 1, \+((X=1;X=2)).
