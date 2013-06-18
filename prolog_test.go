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

	useModule := read.Term_(`:- use_module(library(tap)).`)
	env := term.NewBindings()

	// run tests found in each file
	for _, name := range names {
		if name[0] == '.' {
			continue // skip hidden files
		}
		//t.Logf("-------------- %s", name)
		openTest := func() *os.File {
			f, err := os.Open("t/" + name)
			MaybePanic(err)
			return f
		}

		// which tests does the file have?
		pastUseModule := false
		tests := make([]term.Term, 0)
		terms := read.TermAll_(openTest())
		for _, s := range terms {
			x := s.(term.Callable)
			if pastUseModule {
				if x.Arity() == 2 && x.Name() == ":-" {
					tests = append(tests, x.Arguments()[0])
				} else {
					tests = append(tests, x)
				}
			} else {
				// look for use_module(library(tap)) declaration
				_, err := s.Unify(env, useModule)
				if err == nil {
					pastUseModule = true
					//t.Logf("found use_module directive")
				}
			}
		}

		// run each test in this file
		m := NewMachine().Consult(openTest())
		for _, test := range tests {
			x := test.(term.Callable)
			//t.Logf("proving: %s", test)
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
