% Tests for ->/2
%
% ->/2 is defined in ISO §7.8.7
:- use_module(library(tap)).

% Tests derived from examples in ISO §7.8.7.4
true_true :-
    (true -> true).
true_fail(fail) :-
    (true -> fail).
fail_true(fail) :-
    (fail -> true).
true_bind :-
    (true -> X=1),
    1 = X.
true_bind_before_fail(fail) :-
    (true -> X=1),
    1 = X,
    fail.
cut_works :-
    ((X=1;X=2) -> true),
    1=X.
cut_if_clause :-
    findall(X, ((X=1;X=2) -> true), Xs),
    [1] = Xs.
dont_cut_then_clause :-
    findall(X, (true -> (X=1;X=2)), Xs),
    [1,2] = Xs.

% Tests derived from Prolog: The Standard p. 107
failing_if_clause(fail) :-
    (fail -> (true;true)).
