package golog

import . "github.com/mndrix/golog/term"

import "fmt"
import "github.com/mndrix/golog/scanner"
import "io"
import "reflect"
import "strings"

// Functions match the regular expression
//
//    ReadTerm(All)?

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


// ReadTerm reads a single term from a term source.  A term source can
// be any of the following:
//
//    * type that implements io.Reader
//    * string
//
// Reading a term may consume more content from the source than is strictly
// necessary.
func ReadTerm(src interface{}) (Term, error) {
    r, err := NewTermReader(src)
    if err != nil {
        return nil, err
    }

    return r.Next()
}

// ReadTerm_ is like ReadTerm but panics instead of returning an error.
// (Too bad Go doesn't allow ! as an identifier character)
func ReadTerm_(src interface{}) Term {
    t, err := ReadTerm(src)
    maybePanic(err)
    return t
}

// ReadTermAll reads all available terms from the source
func ReadTermAll(src interface{}) ([]Term, error) {
    r, err := NewTermReader(src)
    if err != nil {
        return nil, err
    }
    return r.All()
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
    terms       chan Term
}

func NewTermReader(src interface{}) (*TermReader, error) {
    ioReader, err := toReader(src)
    if err != nil {
        return nil, err
    }

    tokens := scanner.Scan(ioReader)
    r := TermReader{terms: make(chan Term)}
    r.ResetOperatorTable()
    go r.start(NewLexemeList(tokens))
    return &r, nil
}

// Next returns the next term available from this reader.
// Returns error NoMoreTerms if the reader can't find any more terms.
var NoMoreTerms = fmt.Errorf("No more terms available")
func (r *TermReader) Next() (Term, error) {
    t, ok := <-r.terms
    if !ok {  // channel closed, no more terms
        return nil, NoMoreTerms
    }
    if IsError(t) {
        return nil, t.Error()
    }
    return t, nil
}

// All returns a slice of all terms available from this reader
func (r *TermReader) All() ([]Term, error) {
    terms := make([]Term, 0)
    for t := range r.terms {
        if IsError(t) {
            return terms, t.Error()
        }
        terms = append(terms, t)
    }
    return terms, nil
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

func (r *TermReader) emit(t Term) {
    r.terms <- t
}

func (r *TermReader) start(ll0 *LexemeList) {
    var t Term
    var ll *LexemeList
    for r.readTerm(1200, ll0, &ll, &t) {
        r.emit(t)
        ll0 = ll
    }

    // we won't generate any more terms
    close(r.terms)
}

// parse a single functor
func (r *TermReader) functor(in *LexemeList, out **LexemeList, f *string) bool {
    if in.Value.Type == scanner.Functor {
        *f = in.Value.Content
        *out = in.Next()  // skip functor we just processed
        return true
    }

    return false
}

// consume a single character token
func (r *TermReader) tok(c rune, in *LexemeList, out **LexemeList) bool {
    if in.Value.Type == c {
        *out = in.Next()
        return true
    }
    return false
}

func (r *TermReader) readTerm(p priority, i *LexemeList, o **LexemeList, t *Term) bool {
    return r.term(p, i, o, t) && r.tok(scanner.FullStop, *o, o)
}

// parse a single term
func (r *TermReader) term(p priority, i *LexemeList, o **LexemeList, t *Term) bool {
    var op, f string
    var t0 Term
    var opP, argP priority

    // prefix operator
    if r.prefix(&op, &opP, &argP, i, o) && opP<=p && r.term(argP, *o, o, &t0) {
        opT := NewTerm(op, t0)
        return r.restTerm(opP, p, *o, o, opT, t)
    }

    switch i.Value.Type {
        case scanner.Int:       // integer term §6.3.1.1
            n := NewInt(i.Value.Content)
            *o = i.Next()
            return r.restTerm(0, p, *o, o, n, t)
        case scanner.Float:     // float term §6.3.1.1
            f := NewFloat(i.Value.Content)
            *o = i.Next()
            return r.restTerm(0, p, *o, o, f, t)
        case scanner.Atom:      // atom term §6.3.1.3
            a := NewTerm(i.Value.Content)
            *o = i.Next()
            return r.restTerm(0, p, *o, o, a, t)
        case scanner.Variable:  // variable term §6.3.2
            v := NewVar(i.Value.Content)
            *o = i.Next()
            return r.restTerm(0, p, *o, o, v, t)
        case scanner.Void:  // variable term §6.3.2
            v := NewVar("_")
            *o = i.Next()
            return r.restTerm(0, p, *o, o, v, t)
    }

    // compound term - functional notation §6.3.3
    if r.functor(i,o,&f) && r.tok('(',*o,o) {
        var args []Term
        var arg Term
        for r.term(999,*o,o,&arg) {  // 999 priority per §6.3.3.1
            args = append(args, arg)
            if r.tok(')', *o, o) { break }
            if r.tok(',', *o, o) { continue }
            panic("Unexpected content inside compound term arguments")
        }
        f := NewTerm(f, args...)
        return r.restTerm(0, p, *o, o, f, t)
    }

    return false
}

func (r *TermReader) restTerm(leftP, p priority, i *LexemeList, o **LexemeList, leftT Term, t *Term) bool {
    var op string
    var rightT Term
    var opP, lap, rap priority
    if r.infix(&op, &opP, &lap, &rap, i, o) && p>=opP && leftP<=lap && r.term(rap, *o, o, &rightT) {
        t0 := NewTerm(op, leftT, rightT)
        return r.restTerm(opP, p, *o, o, t0, t)
    }
    if r.postfix(&op, &opP, &lap, i, o) && opP<=p && leftP<=lap {
        opT := NewTerm(op, leftT)
        return r.restTerm(opP, p, *o, o, opT, t)
    }

    // ε rule can always succeed
    *o = i
    *t = leftT
    return true
}

// consume an infix operator and indicate which one it was along with its priorities
func (r *TermReader) infix(op *string, opP, lap, rap *priority, i *LexemeList, o **LexemeList) bool {
    if i.Value.Type != scanner.Atom && i.Value.Type != ',' {
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
func (r *TermReader) prefix(op *string, opP, argP *priority, i *LexemeList, o **LexemeList) bool {
    if i.Value.Type != scanner.Atom {
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
func (r *TermReader) postfix(op *string, opP, argP *priority, i *LexemeList, o **LexemeList) bool {
    if i.Value.Type != scanner.Atom {
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
