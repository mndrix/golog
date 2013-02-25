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

This is a stack of choice points.  Each time Golog encounters nondeterminism, it pushes some choice points onto the disjunction stack.  Special choice points, called cut barriers, act as sentinels on this stack.  A cut removes all choice points stacked on top of one of these barriers.  Backtracking pops a choice point off this stack and follows it to produce a new machine.  That machine is typically a snapshot of the Golog machine as it existed when we first encountered the choice point.  Following the choice point replaces the current Golog machine with that one.  A side effect is discarding all other state that has accumulated since.

This design makes backtracking very easy to understand.  We just revert back to the state of the machine as it existed before.  Go's garbage collector takes care of any state that's no longer needed.

The disjunction stack can be thought of as "computations we haven't tried yet."  Each of those computations is represented as a choice point.  A choice point is just a function which returns a machine.  That machine could come from anywhere.  It could be a snapshot of a machine we saw earlier.  It could be the result of executing a machine in parallel.  It could be the result of executing a machine on several servers, etc.

Conjunctions
------------

This is a stack of goals yet to be proven.  It's the machine's continuation.  A new goal is pushed onto this empty stack.  Executing the top goal of this stack replaces the goal with its corresponding clause body.

This design makes it easy to add call-with-current-continuation to Golog at some point.

Execution
---------

Take a goal off the conjunction stack.  If the goal matches a clause head, push the clause's body onto the conjunction stack.  If the goal might match other clause heads, push those other clauses onto the disjunction stack.  If the goal fails, take a choice point off the disjunction stack and follow it to produce a new machine.  Continue execution on this new machine.


Immutability
============

All data structures in a Golog machine are immutable.  Operations on a Golog machine produce a new machine, leaving the old one completely intact.  I initially chose this approach because it makes backtracking trivial.  It also makes it easy to build a Golog machine during Go's init() and then use that machine in many different web requests without affecting the original machine.

It looks like this design might also make or-parallel and distributed execution easy to implement.  Time and experimentation will tell.
