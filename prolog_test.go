package golog

// This file runs Go tests for Prolog test files under a 't' directory.
// This gives us any easy way to write many Prolog tests without the
// overhead of direct method calls on a testing.T object.  These Prolog
// tests can also be reused by other test implementations.

import "os"
import "testing"
import "github.com/mndrix/golog/read"
import "github.com/mndrix/golog/term"


func TestPureProlog (t *testing.T) {
    // find all t/*.pl files
    file, err := os.Open("t")
    maybePanic(err)
    names, err := file.Readdirnames(-1)

    // run tests found in each file
    for _, name := range names {
        openTest := func () *os.File {
            f, err := os.Open("t/" + name)
            maybePanic(err)
            return f
        }

        // which tests does the file have?
        tests := make([]term.Term, 0)
        terms := read.TermAll_(openTest())
        for _, t := range terms {
            if t.Indicator() == ":-/2" {
                tests = append(tests, t.Arguments()[0])
            }
        }

        // run each test in this file
        m := NewMachine().Consult(openTest())
        for _, test := range tests {
            canProve := m.CanProve(test)
            if test.Arity() > 0 && test.Arguments()[0].String() == "fails" {
                if canProve {
                    t.Errorf("%s: %s should fail", name, test)
                }
            } else {
                if !canProve {
                    t.Errorf("%s: %s failed", name, test)
                }
            }
        }
    }
}
