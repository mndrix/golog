% Tests for msort/2
%
% This isn't defined in ISO but many Prologs have it.
:- use_module(library(tap)).

empty :-
    msort([], L),
    L = [].
single :-
    msort([a], L),
    L = [a].
duplicates :-
    msort([a,a], L),
    L = [a,a].
realistic :-
    msort([a,9,hi(world),4.32,a(0)], L),
    L = [4.32, 9, a, a(0), hi(world)].
wrong(fail) :-
    msort([3,2,1], [3,2,1]).
