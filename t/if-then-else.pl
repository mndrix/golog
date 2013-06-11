% Tests for ;/2 (if-then-else)
%
% ;/2 is defined in ISO §7.8.8
:- use_module(library(tap)).

% Tests derived from examples in ISO §7.8.8.4
true_true_fail :-
    (true -> true; fail).
fail_true_true :-
    (fail -> true; true).
true_fail_fail(fail) :-
    (true -> fail; fail).
fail_true_fail(fail) :-
    (fail -> true; fail).
then :-
    findall(X, (true -> X=1; X=2), Xs),
    Xs = [1].
else :-
    findall(X, (fail -> X=1; X=2), Xs),
    Xs = [2].
disjunction_in_then :-
    findall(X, (true -> (X=1;X=2); true), Xs),
    Xs = [1,2].
disjunction_in_if :-
    findall(X, ((X=1;X=2) -> true; true), Xs),
    Xs = [1].
cut_if :-  % cut inside if shouldn't trim else clause
    ( !,fail -> true; true ).

% Tests derived from Prolog: The Standard p. 107
cut_opaque :-
    (((!,X=1,fail) -> true; fail); X=2),
    2 = X.
