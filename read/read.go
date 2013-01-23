// Read Prolog terms.
package read

import "github.com/mndrix/golog/term"

import "fmt"
import "github.com/mndrix/golog/lex"
import "io"
import "reflect"
import "strings"

// ISO operator specifiers per §6.3.4, table 4
type specifier  int // xf, yf, xfy, etc.
const (
    fx  specifier = iota
    fy
    xfx
    xfy
    yfx
    xf
    yf
)

// ISO operator priorities per §6.3.4
type priority   int     // between 1 and 1200, inclusive


// Term reads a single term from a term source.  A term source can
// be any of the following:
//
//    * type that implements io.Reader
//    * string
//
// Reading a term may consume more content from the source than is strictly
// necessary.
func Term(src interface{}) (term.Term, error) {
    r, err := NewTermReader(src)
    if err != nil {
        return nil, err
    }

    return r.Next()
}

// Term_ is like Term but panics instead of returning an error.
// (Too bad Go doesn't allow ! as an identifier character)
func Term_(src interface{}) term.Term {
    t, err := Term(src)
    maybePanic(err)
    return t
}

// TermAll reads all available terms from the source
func TermAll(src interface{}) ([]term.Term, error) {
    r, err := NewTermReader(src)
    if err != nil {
        return nil, err
    }
    return r.All()
}

// TermAll_ is like TermAll but panics instead of returning an error.
func TermAll_(src interface{}) []term.Term {
    ts, err := TermAll(src)
    maybePanic(err)
    return ts
}

// toReader tries to convert a source into something that implements io.Reader
var _ReaderT reflect.Type = reflect.TypeOf((*io.Reader)(nil)).Elem()
func toReader(src interface{}) (io.Reader, error) {
    if reflect.TypeOf(src).Implements(_ReaderT) {
        return src.(io.Reader), nil
    }
    switch x := src.(type) {
        case string:
            return strings.NewReader(x), nil
    }

    return nil, fmt.Errorf("Can't convert %#v into io.Reader\n", src)
}

type TermReader struct {
    operators   map[string]*[7]priority
    ll          *lex.List
}

func NewTermReader(src interface{}) (*TermReader, error) {
    ioReader, err := toReader(src)
    if err != nil {
        return nil, err
    }

    tokens := lex.Scan(ioReader)
    r := TermReader{ll: lex.NewList(tokens)}
    r.ResetOperatorTable()
    return &r, nil
}

// Next returns the next term available from this reader.
// Returns error NoMoreTerms if the reader can't find any more terms.
var NoMoreTerms = fmt.Errorf("No more terms available")
func (r *TermReader) Next() (term.Term, error) {
    var t term.Term
    var ll *lex.List
    if r.readTerm(1200, r.ll, &ll, &t) {
        if term.IsError(t) {
            return nil, t.Error()
        }
        r.ll = ll
        return term.RenameVariables(t), nil
    }

    return nil, NoMoreTerms
}

// All returns a slice of all terms available from this reader
func (r *TermReader) All() ([]term.Term, error) {
    terms := make([]term.Term, 0)

    t, err := r.Next()
    for err == nil {
        terms = append(terms, t)
        t, err = r.Next()
    }

    if err == NoMoreTerms {
        err = nil
    }
    return terms, err
}

// ResetOperatorTable replaces the reader's current operator table
// with the default table specified in ISO Prolog §6.3.4.4, table 7
func (r *TermReader) ResetOperatorTable() {
    r.operators = make(map[string]*[7]priority)
    r.Op(1200,  xfx, `:-`, `-->`    )
    r.Op(1200,   fx, `:-`, `?-`     )
    r.Op(1100,  xfy, `;`            )
    r.Op(1050,  xfy, `->`           )
    r.Op(1000,  xfy, `,`            )
    r.Op( 900,   fy, `\+`           )
    r.Op( 700,  xfx, `=`, `\=`      )
    r.Op( 700,  xfx, `==`, `\==`, `@<`, `@=<`, `@>`, `@>=`)
    r.Op( 700,  xfx, `=..`)
    r.Op( 700,  xfx, `is`, `=:=`, `=\=`, `<`, `=<`, `>`, `>=`)
    r.Op( 500,  yfx, `+`, `-`, `/\`, `\/`) // syntax highlighter `
    r.Op( 400,  yfx, `*`, `/`, `//`, `rem`, `mod`, `<<`, `<<`)
    r.Op( 200,  xfx, `**`       )
    r.Op( 200,  xfy, `^`        )
    r.Op( 200,   fy, `-`, `\`   ) // syntax highlighter `
}

// Op creates or changes the parsing behavior of a Prolog operator.
// It's equivalent to op/3
func (r *TermReader) Op(p priority, s specifier, os... string) {
    for _, o := range os {
        priorities, ok := r.operators[o]
        if !ok {
            priorities = new([7]priority)
            r.operators[o] = priorities
        }
        priorities[s] = p
    }
}
// parse a single functor
func (r *TermReader) functor(in *lex.List, out **lex.List, f *string) bool {
    if in.Value.Type == lex.Functor {
        *f = in.Value.Content
        *out = in.Next()  // skip functor we just processed
        return true
    }

    return false
}

// parse all list items after the first one
func (r *TermReader) listItems(i *lex.List, o **lex.List, t *term.Term) bool {
    var arg, rest term.Term
    if r.tok(',', i, o) && r.term(999, *o, o, &arg) && r.listItems(*o, o, &rest) {
        *t = term.NewTerm(".", arg, rest)
        return true
    }
    if r.tok('|', i, o) && r.term(999, *o, o, &arg) && r.tok(']', *o, o) {
        *t = arg
        return true
    }
    if r.tok(']', i, o) {
        *t = term.NewTerm("[]")
        return true
    }
    return false
}

// consume a single character token
func (r *TermReader) tok(c rune, in *lex.List, out **lex.List) bool {
    if in.Value.Type == c {
        *out = in.Next()
        return true
    }
    return false
}

func (r *TermReader) readTerm(p priority, i *lex.List, o **lex.List, t *term.Term) bool {
    return r.term(p, i, o, t) && r.tok(lex.FullStop, *o, o)
}

// parse a single term
func (r *TermReader) term(p priority, i *lex.List, o **lex.List, t *term.Term) bool {
    var op, f string
    var t0, t1 term.Term
    var opP, argP priority

    // prefix operator
    if r.prefix(&op, &opP, &argP, i, o) && opP<=p && r.term(argP, *o, o, &t0) {
        opT := term.NewTerm(op, t0)
        return r.restTerm(opP, p, *o, o, opT, t)
    }

    // list notation for compound terms §6.3.5
    if r.tok('[', i, o) && r.term(999, *o, o, &t0) && r.listItems(*o, o, &t1) {
        list := term.NewTerm(".", t0, t1)
        return r.restTerm(0, p, *o, o, list, t)
    }
    if r.tok('[', i, o) && r.tok(']', *o, o) {
        list := term.NewTerm("[]")
        return r.restTerm(0, p, *o, o, list, t)
    }

    switch i.Value.Type {
        case lex.Int:       // integer term §6.3.1.1
            n := term.NewInt(i.Value.Content)
            *o = i.Next()
            return r.restTerm(0, p, *o, o, n, t)
        case lex.Float:     // float term §6.3.1.1
            f := term.NewFloat(i.Value.Content)
            *o = i.Next()
            return r.restTerm(0, p, *o, o, f, t)
        case lex.Atom:      // atom term §6.3.1.3
            a := term.NewTerm(i.Value.Content)
            *o = i.Next()
            return r.restTerm(0, p, *o, o, a, t)
        case lex.Variable:  // variable term §6.3.2
            v := term.NewVar(i.Value.Content)
            *o = i.Next()
            return r.restTerm(0, p, *o, o, v, t)
        case lex.Void:  // variable term §6.3.2
            v := term.NewVar("_")
            *o = i.Next()
            return r.restTerm(0, p, *o, o, v, t)
        case lex.Comment:
            *o = i.Next()               // skip the comment
            return r.term(p, *o, o, t)  // ... and try again
    }

    // compound term - functional notation §6.3.3
    if r.functor(i,o,&f) && r.tok('(',*o,o) {
        var args []term.Term
        var arg term.Term
        for r.term(999,*o,o,&arg) {  // 999 priority per §6.3.3.1
            args = append(args, arg)
            if r.tok(')', *o, o) { break }
            if r.tok(',', *o, o) { continue }
            panic("Unexpected content inside compound term arguments")
        }
        f := term.NewTerm(f, args...)
        return r.restTerm(0, p, *o, o, f, t)
    }

    return false
}

func (r *TermReader) restTerm(leftP, p priority, i *lex.List, o **lex.List, leftT term.Term, t *term.Term) bool {
    var op string
    var rightT term.Term
    var opP, lap, rap priority
    if r.infix(&op, &opP, &lap, &rap, i, o) && p>=opP && leftP<=lap && r.term(rap, *o, o, &rightT) {
        t0 := term.NewTerm(op, leftT, rightT)
        return r.restTerm(opP, p, *o, o, t0, t)
    }
    if r.postfix(&op, &opP, &lap, i, o) && opP<=p && leftP<=lap {
        opT := term.NewTerm(op, leftT)
        return r.restTerm(opP, p, *o, o, opT, t)
    }

    // ε rule can always succeed
    *o = i
    *t = leftT
    return true
}

// consume an infix operator and indicate which one it was along with its priorities
func (r *TermReader) infix(op *string, opP, lap, rap *priority, i *lex.List, o **lex.List) bool {
    if i.Value.Type != lex.Atom && i.Value.Type != ',' {
        return false
    }

    // is this an operator at all?
    name := i.Value.Content
    priorities, ok := r.operators[name]
    if !ok {
        return false
    }

    // what class of operator is it?
    switch {
        case priorities[yfx] > 0:
            *opP = priorities[yfx]
            *lap = *opP
            *rap = *opP - 1
        case priorities[xfy] > 0:
            *opP = priorities[xfy]
            *lap = *opP - 1
            *rap = *opP
        case priorities[xfx] > 0:
            *opP = priorities[xfx]
            *lap = *opP - 1
            *rap = *opP - 1
        default:    // wasn't an infix operator after all
            return false
    }

    *op = name
    *o = i.Next()
    return true
}

// consume a prefix operator. indicate which one it was along with its priority
func (r *TermReader) prefix(op *string, opP, argP *priority, i *lex.List, o **lex.List) bool {
    if i.Value.Type != lex.Atom {
        return false
    }

    // is this an operator at all?
    name := i.Value.Content
    priorities, ok := r.operators[name]
    if !ok {
        return false
    }

    // what class of operator is it?
    switch {
        case priorities[fx] > 0:
            *opP = priorities[fx]
            *argP = *opP - 1
        case priorities[fy] > 0:
            *opP = priorities[fy]
            *argP = *opP
        default:    // wasn't a prefix operator after all
            return false
    }

    *op = name
    *o = i.Next()
    return true
}

// consume a postfix operator. indicate which one it was along with its priority
func (r *TermReader) postfix(op *string, opP, argP *priority, i *lex.List, o **lex.List) bool {
    if i.Value.Type != lex.Atom {
        return false
    }

    // is this an operator at all?
    name := i.Value.Content
    priorities, ok := r.operators[name]
    if !ok {
        return false
    }

    // what class of operator is it?
    switch {
        case priorities[xf] > 0:
            *opP = priorities[xf]
            *argP = *opP - 1
        case priorities[yf] > 0:
            *opP = priorities[yf]
            *argP = *opP
        default:    // wasn't a postfix operator after all
            return false
    }

    *op = name
    *o = i.Next()
    return true
}

func maybePanic(err error) {
    if err != nil {
        panic(err)
    }
}
