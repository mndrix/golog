package golog

import "fmt"
import "github.com/mndrix/golog/scanner"
import "strings"

// Functions match the regular expression
//
//    ReadTerm(String)?(One|All)?

type ReaderMode int
const (
    Read    = iota
    Consult
)

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
    mode        ReaderMode
    operators   map[operator]*[7]priority
    dst         chan<- Term
}

func ReadTermTokens(tokens <-chan *scanner.Lexeme, mode ReaderMode) <-chan Term {
    ch := make(chan Term)
    ll := NewLexemeList(tokens)
    r := newReader(mode, ch)
    go r.start(ll)
    return ch
}
func ReadTermString(s string, mode ReaderMode) <-chan Term {
    r := strings.NewReader(s)
    return ReadTermTokens(scanner.Scan(r), mode)
}
func ReadTermStringOne(s string, mode ReaderMode) (Term, error) {
    ch := ReadTermString(s, mode)
    t := <-ch
    if t == nil {  // channel closed right away
        return nil, fmt.Errorf("No terms found in `%s`", s)
    }
    if IsError(t) {
        return nil, t.Error()
    }
    return t, nil
}
func ReadTermStringAll(s string, mode ReaderMode) ([]Term, error) {
    ch := ReadTermString(s, mode)
    return readAll(ch)
}

func readAll(ch <-chan Term) ([]Term, error) {
    terms := make([]Term, 0)
    for t := range ch {
        if IsError(t) {
            return terms, t.Error()
        }
        terms = append(terms, t)
    }
    return terms, nil
}


// resetOperatorTable replaces the reader's current operator table
// with the default table specified in §6.3.4.4, table 7
func newReader(mode ReaderMode, ch chan<- Term) *reader {
    r := reader{mode: mode, dst: ch}
    r.resetOperatorTable()
    return &r
}
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
    var t0 Term

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
