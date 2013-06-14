# Common Features

This is a list of features that are common to Prolog implementations.
Golog supports them all.

TODO

# Rare Features

This is a list of features which are found in some Prolog
implementations, but they're not commonly available.

## Automatic, arbitrary precision numbers

Most programming languages treat the text "1.23" as a floating point
number.  Calculations with that value are subject to approximation and
rounding errors.  When Golog encounters a number, it uses the most
accurate representation possible.  In this case, it's stored as the
rational number 123/100.

Arithmetic calculations try to maintain the most accurate
representation at all times.  Numbers are only converted to floating
point representation as a last resort.

All this behavior is strictly an implementation detail.  One can use
Golog as if it only had integer and float numbers.  However, the extra
precision is especially helpful for working with currency amounts,
etc.
