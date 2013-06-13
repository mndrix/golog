package golog

// This file runs Go tests for Prolog test files under a 't' directory.
// This gives us an easy way to write many Prolog tests without writing
// a bunch of manual Go tests.  Because the tests are written in Prolog
// they can be reused by other Prolog implementations.

import "os"
import "testing"
import "github.com/mndrix/golog/read"
import "github.com/mndrix/golog/term"
import . "github.com/mndrix/golog/util"

func TestPureProlog(t *testing.T) {
	// find all t/*.pl files
	file, err := os.Open("t")
	MaybePanic(err)
	names, err := file.Readdirnames(-1)

	// run tests found in each file
	for _, name := range names {
		if name[0] == '.' {
			continue // skip hidden files
		}
		openTest := func() *os.File {
			f, err := os.Open("t/" + name)
			MaybePanic(err)
			return f
		}

		// which tests does the file have?
		tests := make([]term.Term, 0)
		terms := read.TermAll_(openTest())
		for _, t := range terms {
			x := t.(term.Callable)
			if x.Indicator() == ":-/2" {
				tests = append(tests, x.Arguments()[0])
			}
		}

		// run each test in this file
		m := NewMachine().Consult(openTest())
		for _, test := range tests {
			x := test.(term.Callable)
			canProve := m.CanProve(test)
			if x.Arity() > 0 && x.Arguments()[0].String() == "fail" {
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
