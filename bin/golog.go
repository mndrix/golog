// A primitive Prolog top level for Golog.  See golog.sh in the repository
// for recommended usage.
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mndrix/golog"
	"github.com/mndrix/golog/read"
	"github.com/mndrix/golog/term"
	"github.com/mndrix/ps"
)

func main() {
	// create a Golog machine
	m := initMachine()

	// ?- do(stuff).
	in := bufio.NewReader(os.Stdin)
	for {
		warnf("?- ")

		// process the user's query
		query, err := in.ReadString('\n')
		if err == io.EOF {
			warnf("\n")
			os.Exit(0)
		}
		if err != nil {
			warnf("Trouble reading from stdin: %s\n", err)
			os.Exit(1)
		}
		goal, err := read.Term(query)
		if err != nil {
			warnf("Problem parsing the query: %s\n", err)
			continue
		}

		// execute user's query
		variables := term.Variables(goal)
		answers := m.ProveAll(goal)

		// showing 0 results is easy and fun!
		if len(answers) == 0 {
			warnf("no.\n\n")
			continue
		}

		// show each answer in turn
		if variables.Size() == 0 {
			warnf("yes.\n")
			continue
		}
		for i, answer := range answers {
			lines := make([]string, 0)
			variables.ForEach(func(name string, variable interface{}) {
				v := variable.(*term.Variable)
				val := answer.Resolve_(v)
				line := fmt.Sprintf("%s = %s", name, val)
				lines = append(lines, line)
			})

			warnf("%s", strings.Join(lines, "\n"))
			if i == len(answers)-1 {
				warnf("\t.\n\n")
			} else {
				warnf("\t;\n")
			}
		}
	}
}

// warnf generates formatted output on stderr
func warnf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Stderr.Sync()
}

// initMachine creates a new Golog machine based on command line arguments
func initMachine() golog.Machine {
	m := golog.NewMachine()

	// are we supposed to load some code into the machine?
	if len(os.Args) > 1 {
		filename := os.Args[1]
		warnf("Opening %s ...\n", filename)
		file, err := os.Open(filename)
		if err != nil {
			warnf("Can't open file: %s\n", err)
			os.Exit(1)
		}
		m = m.Consult(file)
	}

	return m
}
