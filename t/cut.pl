

% Tests that caught me at least once
multiple_cuts :-
    findall(X, ((X=1;X=2),!,!), Xs),
    Xs = [1].
