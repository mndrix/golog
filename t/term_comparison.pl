% Tests for ==/2 and other term comparison operators
%
% As defined in ISO §8.4
:- use_module(library(tap)).

% examples taken from ISO §8.4.1.4
1.0 @=< 1.
1.0 @<  1.
'1 \\== 1'(fail) :-
    1 \== 1.
aardvark @=< zebra.
short @=< short.
short @=< shorter.
'short @>= shorter'(fail) :-
    short @>= shorter.
'foo(a,b) @< north(a).'(fail) :-
    foo(a,b) @< north(a).
foo(b) @> foo(a).
foo(a,_X) @< foo(b,_Y).
'foo(X,a) @< foo(Y,b)'(todo('implementation dependent')) :-
    foo(_X,a) @< foo(_Y,b).
X @=< X.
X == X.
'X @=< Y'(todo('implementation dependent')) :-
    _X @=< _Y.
'X == Y'(fail) :-
    _X == _Y.
_ \== _.
'_ == _'(fail) :-
    _ == _.
'_ @=< _'(todo('implementation dependent')) :-
    _ @=< _.
'foo(X,a) @=< foo(Y,b)'(todo('implementation dependent')) :-
    foo(_X,a) @=< foo(_Y,b).
