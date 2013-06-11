% Tests for ground/1
%
% This isn't defined in ISO but many Prologs have it.
:- use_module(library(tap)).

atom :-
    ground(here_is_an_atom).

integer :-
    ground(19).

float :-
    ground(7.321).

'compound term' :-
    ground(hi(one, two, three)).

nil :-
    ground([]).

'complete list' :-
    ground([a, b, c]).

'bound variable' :-
    X = something,
    ground(X).

'unbound variable'(fail) :-
    ground(_).

'improper list'(fail) :-
    ground([a,b|_]).

'list pattern'(fail) :-
    ground([_|_]).

'compound term with variables'(fail) :-
    ground(a_term(one, X, three, X)).
