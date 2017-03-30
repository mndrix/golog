package golog

import (
	"bytes"
	"fmt"
	"os"
	"regexp"

	"github.com/mndrix/golog/term"
)

func RegisterHelp(m Machine, h map[string]string) {
	rh := m.(*machine).help
	for k, v := range h {
		rh[k] = v
	}
}

func NewInteractiveMachine() Machine {
	res := NewMachine()
	res.(*machine).help = builtInHelp()
	return res.RegisterForeign(map[string]ForeignPredicate{
		"help/0":    InteractiveHelp0,
		"help/1":    InteractiveHelp1,
		"apropos/1": InteractiveApropos1,
	})
}

func apropos(m Machine, pattern string) string {
	var keys []string
	pat := regexp.MustCompile(pattern)
	buf := bytes.NewBuffer([]byte{})
	if np, ok := m.(*machine); ok {
		for i := 0; i < smallThreshold; i++ {
			for _, k := range np.smallForeign[i].Keys() {
				keys = append(keys, fmt.Sprintf("%s/%d", k, i))
			}
		}
		keys = append(keys, np.largeForeign.Keys()...)

		for _, k := range keys {
			if pat.MatchString(k) {
				_, _ = buf.WriteString(k)
				_, _ = buf.WriteRune('\n')
			}
		}
		return string(buf.Bytes())
	}
	return "% no results."
}

func builtInHelp() map[string]string {
	return map[string]string{
		"help/0": `Prints help usage.`,
		"help/1": `Prints the usage of the given predicate.`,
		"apropos/1": `Looks up the database for the predicates 
matching regexp and prints them.`,
		"!/0":    `Cut operator, prevents backtracking beyond this point.`,
		",/2":    `Conjunction operator.`,
		"->/2":   `Implication operator.`,
		";/2":    `Disjunction operator.`,
		"=/2":    `Unification operator.`,
		"=:=/2":  `Numeric equality operator.`,
		"==/2":   `Equality operator.`,
		"\\==/2": `Equality negation operator.`,
		"@</2":   `Less than operator.`,
		"@=</2":  `Less than or equal operator`,
		"@>/2":   `Greater than operator.`,
		"@>=/2":  `Greater than or equal operator.`,
		`\+/1`:   `Negation operator.`,
		"atom_codes/2": `Second argument is the list containing the character
codes of the name of the first argument.`,
		"atom_number/2": `Second argument is the number represented by the name
of the first argument.`,
		"call/1": `Evaluates its argument.`,
		"call/2": `Constructs term from its arguments and evaluates it.`,
		"call/3": `Constructs term from its arguments and evaluates it.`,
		"call/4": `Constructs term from its arguments and evaluates it.`,
		"call/5": `Constructs term from its arguments and evaluates it.`,
		"call/6": `Constructs term from its arguments and evaluates it.`,
		"downcase_atom/2": `Second argument is the atom with the name made up of
all the same characters of the first atom, just in lower case`,
		"fail/0": `Fail unconditionaly.`,
		"findall/3": `Generate variables from template (first argument),
bind them in the second argument, then collect the bindings in the third argument.`,
		"ground/1": `Succeeds if the argument is ground.`,
		"is/2": `Succeeds if the numerical expressions on both sides
evaluate to the same number.`,
		"listing/0": `Prints all predicates known to this interpreter.`,
		"msort/2":   `Sorts list.`,
		"printf/1":  `Prints its first argument.`,
		"printf/2": `Populates the template in the first argument with
the printable representations of its second argument (which must be a list)
and prints it.`,
		"printf/3": `Same as printf/2, but prints into a stream given
in the first argument.`,
		"succ/2": `True if its second argument is one greater than its
first argument.`,
		"var/1": `True if its argument is a variable.`,
	}
}

func InteractiveHelp0(m Machine, args []term.Term) ForeignReturn {
	_, _ = fmt.Fprintf(os.Stderr, `
Use:
?- help(predicate).
to print documentation of the predicate.
?- apropos("regexp").
to look for predicates matching regexp.
`)
	return ForeignTrue()
}

func InteractiveHelp1(m Machine, args []term.Term) ForeignReturn {
	var subj string
	if term.IsAtom(args[0]) {
		subj = args[0].(*term.Atom).Name()
	} else if term.IsString(args[0]) {
		subj = term.RawString(args[0])
	} else {
		panic(fmt.Sprintf("Illegal argument to help/1: %s", args[0]))
	}
	rh := m.(*machine).help
	help := rh[subj]

	if help == "" {
		_, _ = fmt.Fprintf(os.Stderr, "No help on %s\n", subj)
		help = apropos(m, subj)
		if help != "" {
			_, _ = fmt.Fprintf(os.Stderr, "Maybe you meant\n%s\n", help)
		}
	} else {
		_, _ = fmt.Fprintf(os.Stderr, help+"\n")
	}
	return ForeignTrue()
}

func InteractiveApropos1(m Machine, args []term.Term) ForeignReturn {
	var subj string
	if term.IsAtom(args[0]) {
		subj = args[0].(*term.Atom).Name()
	} else if term.IsString(args[0]) {
		subj = term.RawString(args[0])
	} else {
		panic(fmt.Sprintf("Illegal argument to apropos/1: %s", args[0]))
	}
	_, _ = fmt.Fprintf(os.Stderr, apropos(m, subj))
	return ForeignTrue()
}
