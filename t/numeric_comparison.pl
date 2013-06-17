% Tests for =:=/2 and other numeric comparison operators
%
% As defined in ISO §8.7
:- use_module(library(tap)).

'integer equality' :-
    7 =:= 7.

'integer equality failing'(fail) :-
    97 =:= 2.

'float equality' :-
    1.0 =:= 1.0000.

'float equality failing'(fail) :-
    1.7 =:= 2.0.

'mixed equality' :-
    9 =:= 9.00.

'arithmetic equality' :-
    X = 1 + 2,
    X + 6 =:= X * 3.

'more arithmetic equality' :-
    3*2 =:= 7-1.

'integer quotient equality' :-
    1/3 =:= 2/6.

'zero is not one'(fail) :-
    0 =:= 1.
