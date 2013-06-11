% Tests for memberchk/1
%
% This isn't defined in ISO but many Prologs have it.
:- use_module(library(tap)).

'empty list'(fail) :-
    memberchk(a, []).

'singleton list w/o match'(fail) :-
    memberchk(a, [b]).

'list w/o match'(fail) :-
    memberchk(a, [b, d, c, e]).

'only element' :-
    memberchk(a, [a]).

'first element' :-
    memberchk(c, [c, a, j, k]).

'final element' :-
    memberchk(c, [a, j, k, c]).

typical :-
    memberchk(c, [a, j, c, k]).

'with duplicates' :-
    memberchk(c, [a, b, a, c, d, c]).
