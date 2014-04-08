% Tests for var/1

:- use_module(library(tap)).

% Tests derived from Prolog: The Standard p. 181
simple_var :-
    var(_).

entangled_var :-
    X = _Y,
    var(X).

term_with_vars(fail) :-
    X = f(_),
    var(X).

not_a_var(fail) :-
    var(hi).
