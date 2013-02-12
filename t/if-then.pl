% Tests for ->/2
%
% ->/2 is defined in ISO §7.8.7

% Tests derived from examples in ISO §7.8.7.4
true_true :-
    (true -> true).
true_fail(fails) :-
    (true -> fail).
fail_true(fails) :-
    (fail -> true).
true_bind :-
    (true -> X=1),
    1 = X.
true_bind_before_fail(fails) :-
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
failing_if_clause(fails) :-
    (fail -> (true;true)).
