% Tests for ignore/1

true_cp.
true_cp.

:- use_module(library(tap)).

goal_succeeds :-
    ignore(true).

goal_fails :-
    ignore(fail).

goal_succeeds_with_choicepoints :-
    ignore(true_cp).