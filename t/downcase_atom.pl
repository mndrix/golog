% Tests for downcase_atom/2
%
% downcase_atom/2 is an SWI-Prolog extension for converting an
% atom to lowercase.  Other Prolog systems have similar
% predicates, but the semantics seem less useful or more confusing.
:- use_module(library(tap)).

already_lowercase :-
    downcase_atom(foo, foo),
    downcase_atom(foo, Foo),
    Foo = foo.
all_uppercase :-
    downcase_atom('YELL', A),
    A = yell,
    downcase_atom('YELL', yell).
mixed_case :- 
    downcase_atom('Once upon a time...', A),
    A = 'once upon a time...'.
