package term

import "fmt"

// Returned by Unify() if the unification fails
var CantUnify error = fmt.Errorf("Can't unify the given terms")

func Unify(e Bindings, a, b Term) (Bindings, error) {
    return a.Unify(e, b)
}
