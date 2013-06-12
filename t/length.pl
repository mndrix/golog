% Tests for sort/2
%
% Part of the de facto standard.
:- use_module(library(tap)).

typical :-
    length([one,two,three], 3).

empty :-
    length([], 0).

singleton :-
    length([x], 1).

'build a list' :-
    length(Xs, 3),
    Xs = [_,_,_].
