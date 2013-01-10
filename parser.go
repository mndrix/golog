package golog

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
type operator   string

type reader struct {
    operators   map[operator]*[7]priority
    dst         chan<- Term
}

// ReadTerm reads a single term from a term source.  A term source can
// be any of the following:
//
//    * type that implements io.Reader
//    * string
//
// Reading a term may consume more content from the source than is strictly
// necessary.
func ReadTerm(src interface{}) (Term, error) {
    r, err := toReader(src)
    if err != nil {
        return nil, err
    }

    return one(readTermsTokens(scanner.Scan(r)))
}

func ReadTermAll(src interface{}) ([]Term, error) {
    r, err := toReader(src)
    if err != nil {
        return nil, err
    }
    return all(readTermsTokens(scanner.Scan(r)))
}

func readTermsTokens(tokens <-chan *scanner.Lexeme) <-chan Term {
    ch := make(chan Term)
    ll := NewLexemeList(tokens)
    r := newReader(ch)
    go r.start(ll)
    return ch
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

// one takes a single term out of a term channel
func one(ch <-chan Term) (Term, error) {
    t := <-ch
    if t == nil {  // channel closed right away
        return nil, fmt.Errorf("No terms found in term channel")
    }
    if IsError(t) {
        return nil, t.Error()
    }
    return t, nil
}

// all takes all terms out of a term channel and puts them in a slice
func all(ch <-chan Term) ([]Term, error) {
    terms := make([]Term, 0)
    for t := range ch {
        if IsError(t) {
            return terms, t.Error()
        }
        terms = append(terms, t)
    }
    return terms, nil
}


func newReader(ch chan<- Term) *reader {
    r := reader{dst: ch}
    r.resetOperatorTable()
    return &r
}

// resetOperatorTable replaces the reader's current operator table
// with the default table specified in §6.3.4.4, table 7
func (r *reader) resetOperatorTable() {
    r.operators = make(map[operator]*[7]priority)
    r.op(1200,  xfx, []operator{`:-`, `-->`})
    r.op(1200,   fx, []operator{`:-`, `?-`})
    r.op(1100,  xfy, []operator{`;`})
    r.op(1050,  xfy, []operator{`->`})
    r.op(1000,  xfy, []operator{`,`})
    r.op( 900,   fy, []operator{`\+`})
    r.op( 700,  xfx, []operator{`=`, `\=`})
    r.op( 700,  xfx, []operator{`==`, `\==`, `@<`, `@=<`, `@>`, `@>=`})
    r.op( 700,  xfx, []operator{`=..`})
    r.op( 700,  xfx, []operator{`is`, `=:=`, `=\=`, `<`, `=<`, `>`, `>=`})
    r.op( 500,  yfx, []operator{`+`, `-`, `/\`, `\/`}) // syntax highlighter `
    r.op( 400,  yfx, []operator{`*`, `/`, `//`, `rem`, `mod`, `<<`, `<<`})
    r.op( 200,  xfx, []operator{`**`})
    r.op( 200,  xfy, []operator{`^`})
    r.op( 200,   fy, []operator{`-`, `\`})             // syntax highlighter `
}
func (r *reader) op(p priority, s specifier, os []operator) {
    for _, o := range os {
        priorities, ok := r.operators[o]
        if !ok {
            priorities = new([7]priority)
            r.operators[o] = priorities
        }
        priorities[s] = p
    }
}
func (r *reader) emit(t Term) {
    r.dst <- t
}
func (r *reader) start(ll0 *LexemeList) {
    var t Term
    var ll *LexemeList
    for r.readTerm(1200, ll0, &ll, &t) {
        r.emit(t)
        ll0 = ll
    }

    // we won't generate any more terms
    close(r.dst)
}

// parse a single functor
func (r *reader) functor(in *LexemeList, out **LexemeList, f *string) bool {
    if in.Value.Type == scanner.Functor {
        *f = in.Value.Content
        *out = in.Next()  // skip functor we just processed
        return true
    }

    return false
}

// consume a single character token
func (r *reader) tok(c rune, in *LexemeList, out **LexemeList) bool {
    if in.Value.Type == c {
        *out = in.Next()
        return true
    }
    return false
}

func (r *reader) readTerm(p priority, i *LexemeList, o **LexemeList, t *Term) bool {
    return r.term(p, i, o, t) && r.tok(scanner.FullStop, *o, o)
}

// parse a single term
func (r *reader) term(p priority, i *LexemeList, o **LexemeList, t *Term) bool {
    var f string
    var op, t0 Term
    var opP, argP priority

    // prefix operator
    if r.prefix(&op, &opP, &argP, i, o) && opP<=p && r.term(argP, *o, o, &t0) {
        opT := NewTerm(op.Functor(), t0)
        return r.restTerm(opP, p, *o, o, opT, t)
    }

    switch i.Value.Type {
        case scanner.Atom:      // atom term §6.3.1.3
            a := NewTerm(i.Value.Content)
            *o = i.Next()
            return r.restTerm(0, p, *o, o, a, t)
        default:
    }

    // compound term - functional notation §6.3.3
    if r.functor(i,o,&f) && r.tok('(',*o,o) && r.term(1200,*o,o,&t0) && r.tok(')',*o,o) {
        *t = NewTerm(f, t0)
        return true
    }

    return false
}

func (r *reader) restTerm(leftP, p priority, i *LexemeList, o **LexemeList, leftT Term, t *Term) bool {
    var op, rightT Term
    var opP, lap, rap priority
    if r.infix(&op, &opP, &lap, &rap, i, o) && p>=opP && leftP<=lap && r.term(rap, *o, o, &rightT) {
        t0 := NewTerm(op.Functor(), leftT, rightT)
        return r.restTerm(opP, p, *o, o, t0, t)
    }

    // ε rule can always succeed
    *o = i
    *t = leftT
    return true
}

// consume an infix operator and indicate which one it was along with its priorities
func (r *reader) infix(op *Term, opP, lap, rap *priority, i *LexemeList, o **LexemeList) bool {
    if i.Value.Type != scanner.Atom && i.Value.Type != ',' {
        return false
    }

    // is this an operator at all?
    name := operator(i.Value.Content)
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

    *op = NewTerm(string(name))
    *o = i.Next()
    return true
}

// consume a prefix operator. indicate which one it was along with its priority
func (r *reader) prefix(op *Term, opP, argP *priority, i *LexemeList, o **LexemeList) bool {
    if i.Value.Type != scanner.Atom {
        return false
    }

    // is this an operator at all?
    name := operator(i.Value.Content)
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
        default:    // wasn't an infix operator after all
            return false
    }

    *op = NewTerm(string(name))
    *o = i.Next()
    return true
}
