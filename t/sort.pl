% Tests for sort/2
%
% Part of the defacto standard.
:- use_module(library(tap)).

typical :-
    sort([a,d,c,e], [a,c,d,e]).

'already in order' :-
    sort([c,d,e,f,g], [c,d,e,f,g]).

'input has duplicates' :-
    sort([a,b,a,c,d], [a,b,c,d]).

'empty list' :-
    sort([], []).

'complex terms' :-
    sort([sue, hello(world), 9, 42.95], [9, 42.95, sue, hello(world)]).
