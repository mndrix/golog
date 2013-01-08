package golog

import "github.com/mndrix/golog/scanner"

type LexemeList struct {
    Value   *scanner.Lexeme
    next    *LexemeList
    src     <-chan *scanner.Lexeme
}

// NewLexemList returns a new lexeme list which pulls lexemes from
// the given source channel.  Creating a new list consumes one lexeme
// from the source channel.
func NewLexemeList(src <-chan *scanner.Lexeme) *LexemeList {
    lexeme, ok := <-src
    if !ok {
        lexeme = &scanner.Lexeme{Type: scanner.EOF}
    }
    return &LexemeList{
        Value:  lexeme,
        next:   nil,
        src:    src,
    }
}

// Next returns the next element in the lexeme list, pulling a lexeme
// from the source channel, if necessary
func (self *LexemeList) Next() *LexemeList {
    if self.next == nil {
        next := NewLexemeList(self.src)
        self.next = next
    }
    return self.next
}
