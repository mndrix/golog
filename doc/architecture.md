Architecture
============

This document describes how the Golog interpreter is designed. Understanding this document is not necessary for running Prolog code.  It's intended to help those who want to hack on the interpreter.

A Golog machine tracks the following state:

  * database
  * foreign predicates
  * environment
  * disjunctions (called disjs)
  * conjunctions (called conjs)

Database
--------

The database holds all predicates defined in Prolog.  It's conceptually a map from predicate indicators (foo/2) to a list of terms.  Those terms define the predicate's clauses.  A database may support indexing.  It may represent clauses internally using whatever means seems reasonble.  The database is encouraged to inspect all clauses, their shape and number when deciding how to represent clauses internally.

Eventually, a Golog machine will map atoms (module names) to databases.  In this scenario each database represents a module.  Databases might also become first class values that are garbage collected like other values.

Foreign Predicates
------------------

This is similar to a database and may eventually be merged with it.  Conceptually, it's a map from predicate indicators to native Go functions.  These Go functions implement the associated predicates.

Environment
-----------

An environment encapsulates variable bindings.  Unification occurs in the presence of this environment.  At the moment, unification doesn't replace variables in terms, it just adds more bindings to the environment.  It might be reasonable to "compact" the environment occassionally by discarding unneeded bindings.

Disjunctions
------------

This is a stack of choice points.  Each time Golog encounters nondeterminism, it pushes some choice points onto the disjunction stack.  Special choice points, called cut barriers, act as sentinels on this stack.  A cut removes all choice points stacked on top of one of these barriers.  Backtracking pops a choice point of this stack and follows it to produce a new machine.  That machine is typically a snapshot of the Golog machine as it existed when we first encountered the choice point.  Following the choice point replaces the current Golog machine with that one.  A side effect is discarding all other state that has accumulated since.

This design makes backtracking very easy to understand.  We just revert back to the state of the machine as it existed before.  Go's garbage collector takes care of any state that's no longer needed.

The disjunction stack can be thought of as "computations we haven't tried yet."  Each of those computations is represented as a choice point.  A choice point is just a function which returns a machine.  That machine could come from anywhere.  It could be a snapshot of a machine we saw earlier.  It could be the result of executing a machine in parallel.  It could be the result of executing a machine on several servers, etc.

Conjunctions
------------

TODO


Immutability
============

TODO

  * what is Step()?
