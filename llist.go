package golog

import "github.com/mndrix/golog/lex"

type LexemeList struct {
    Value   *lex.Eme
    next    *LexemeList
    src     <-chan *lex.Eme
}

// NewLexemList returns a new lexeme list which pulls lexemes from
// the given source channel.  Creating a new list consumes one lexeme
// from the source channel.
func NewLexemeList(src <-chan *lex.Eme) *LexemeList {
    lexeme, ok := <-src
    if !ok {
        lexeme = &lex.Eme{Type: lex.EOF}
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
