% Tests for phrase/2 and phrase/3
%
% This isn't defined in ISO but most Prologs have it.

% helper predicates
alphabet([a,b,c,d|X], X).

triplets([X,X,X|Xs], Xs).

greeting(Whom, [hello,Whom|Xs], Xs).

:- use_module(library(tap)).

'phrase/2 alphabet' :-
    phrase(alphabet, [a,b,c,d]).

'phrase/3 alphabet' :-
    phrase(alphabet, [a,b,c,d,e,f], Rest),
    Rest = [e,f].

'phrase/2 triplets' :-
    phrase(triplets, [j,j,j]).

'phrase/3 triplets' :-
    phrase(triplets, [j,j,j,k], Rest),
    Rest = [k].

'phrase/2 greeting' :-
    phrase(greeting(Whom), [hello,michael]),
    Whom = michael.

'phrase/3 greeting' :-
    phrase(greeting(Whom), [hello,john,doe], Rest),
    Whom = john,
    Rest = [doe].
