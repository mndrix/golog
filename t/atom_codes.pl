% Tests for atom_codes/2
%
% atom_codes/2 is defined in ISO §8.6.5

% Tests derived from examples in ISO §8.6.5.4
empty_atom :-
    atom_codes('', L),
    L = [].
nil :-
    atom_codes([], L),
    L = [0'[, 0']].
%single_quote :-
%    atom_codes('''', L),
%    L = [0''']. % '] for syntax
ant :-
    atom_codes(ant, L),
    L = [0'a, 0'n, 0't]. % ' for syntax
sop :-
    atom_codes(Str, [0's, 0'o, 0'p]),  % ' for syntax
    Str = sop.
partial_list :-
    atom_codes('North', [0'N | X]), % ' for syntax
    X = [0'o, 0'r, 0't, 0'h].
missing_a_code(fails) :-
    atom_codes(soap, [0's, 0'o, 0'p]). % ' for syntax
%all_variables(throws(instantiation_error)) :-
%    atom_codes(_,_).


% Tests derived from Prolog: The Standard p. 53
anna_var :-
    atom_codes(X, [0'a, 0'n, 0'n, 0'a]),
    X = anna.
anna_ground :-
    atom_codes(anna, [0'a, 0'n, 0'n, 0'a]).
%var_and_list_with_var(throws(instantiation_error)) :-
%    atom_codes(_, [0'a | _]).  % ' for syntax


% Tests covering edge cases I've encountered
bound_variables :-
    L = "mario",
    atom_codes(A, L),
    A = mario.
