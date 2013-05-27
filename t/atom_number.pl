% Tests for atom_number/2
%
% atom_number/2 is an SWI-Prolog extension.  We follow its semantics.

empty_atom(fails) :-
    atom_number('', _).
nil(fails) :-
    atom_number([], _).

atom_to_int :-
    atom_number('13', N),
    N = 13.
int_to_atom :-
    atom_number(A, 7),
    A = '7'.

atom_to_float :-
    atom_number('7.23', N),
    N = 7.23.
float_to_atom :-
    atom_number(A, 3.1415),
    A = '3.1415'.

hex :-
    atom_number('0xa9', N),
    169 = N.

bound_variables :-
    A = '4',
    atom_number(A, N),
    4 = N.
