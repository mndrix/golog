// Defines the Prolog standard library which can be implemented
// in Prolog itself.  Each variable in this package is a predicate
// definition.  At init time, they're combined into a single string
// in the Prelude var.
package prelude

import "strings"

var Prelude string
func init() {
    Prelude = strings.Join([]string{
        Phrase3,
    }, "\n\n")
}

// phrase(:DCGBody, ?List, ?Rest) is nondet.
//
// True when DCGBody applies to the difference List/Rest.
var Phrase3 = `
phrase(Dcg, Head, Tail) :-
    call(Dcg, Head, Tail).
`

