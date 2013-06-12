% Tests for unification
%
% These tests are only those cases which have caused Golog trouble
% in the past.  Eventually I hope to have many more unification
% tests here.
:- use_module(library(tap)).

% This exact unification failure arose in a production language model
'product/3 with many anonymous variables' :-
    product(_,'wham-shell', _) = product(_,_,[]).

% If _ creates a distinct variable, this unification succeeds.  If
% all _ variables are the same, it fails.  This failure was the
% root cause of the 'product/3 with many anonymous variables' failure.
'anonymous variables are distinct' :-
    _ = 3,
    _ = 2.
