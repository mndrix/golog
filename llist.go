package golog

import "fmt"
import "github.com/mndrix/golog/scanner"

type LexemeList struct {
    Value   *scanner.Lexeme
    next    *LexemeList
    src     <-chan *scanner.Lexeme
}

// ChannelClosed is returned to indicate that the channel backing
// a LexemeList has been closed
var ChannelClosed = fmt.Errorf("Source channel is closed")

// NewLexemList returns a new lexeme list which pulls lexemes from
// the given source channel.  Creating a new list consumes one lexeme
// from the source channel.
// Can return error with ChannelClosed if the
// channel can provide no more lexemes
func NewLexemeList(src <-chan *scanner.Lexeme) (*LexemeList, error) {
    lexeme, ok := <-src
    if !ok {
        return nil, ChannelClosed
    }
    return &LexemeList{
        Value:  lexeme,
        next:   nil,
        src:    src,
    }, nil
}

// Next returns the next element in the lexeme list, pulling a lexeme
// from the source channel, if necessary
func (self *LexemeList) Next() (*LexemeList, error) {
    if self.next == nil {
        next, err := NewLexemeList(self.src)
        if err != nil {
            return nil, err
        }
        self.next = next
    }
    return self.next, nil
}
