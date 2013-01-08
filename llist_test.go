package golog

import "github.com/mndrix/golog/scanner"
import "testing"

func TestLLBasic(t *testing.T) {
    ch := make(chan *scanner.Lexeme)
    go func() {
        ch <- &scanner.Lexeme{scanner.Atom, "1"}
        ch <- &scanner.Lexeme{scanner.Atom, "2"}
        ch <- &scanner.Lexeme{scanner.Atom, "3"}
        close(ch)
    }()

    l := NewLexemeList(ch)
    if l.Value.Type != scanner.Atom || l.Value.Content != "1" {
        t.Errorf("Wrong lexeme: %s", l.Value.Content)
    }

    l = l.Next()
    if l.Value.Type != scanner.Atom || l.Value.Content != "2" {
        t.Errorf("Wrong lexeme: %s", l.Value.Content)
    }

    l = l.Next()
    if l.Value.Type != scanner.Atom || l.Value.Content != "3" {
        t.Errorf("Wrong lexeme: %s", l.Value.Content)
    }

    l = l.Next()
    if l.Value.Type != scanner.EOF {
        t.Errorf("Backing channel not closed")
    }

    l = l.Next()
    if l.Value.Type != scanner.EOF {
        t.Errorf("Backing channel still not closed")
    }
}
