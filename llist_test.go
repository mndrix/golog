package golog

import "github.com/mndrix/golog/lex"
import "testing"

func TestLLBasic(t *testing.T) {
    ch := make(chan *lex.Eme)
    go func() {
        ch <- &lex.Eme{lex.Atom, "1"}
        ch <- &lex.Eme{lex.Atom, "2"}
        ch <- &lex.Eme{lex.Atom, "3"}
        close(ch)
    }()

    l := NewLexemeList(ch)
    if l.Value.Type != lex.Atom || l.Value.Content != "1" {
        t.Errorf("Wrong lexeme: %s", l.Value.Content)
    }

    l = l.Next()
    if l.Value.Type != lex.Atom || l.Value.Content != "2" {
        t.Errorf("Wrong lexeme: %s", l.Value.Content)
    }

    l = l.Next()
    if l.Value.Type != lex.Atom || l.Value.Content != "3" {
        t.Errorf("Wrong lexeme: %s", l.Value.Content)
    }

    l = l.Next()
    if l.Value.Type != lex.EOF {
        t.Errorf("Backing channel not closed")
    }

    l = l.Next()
    if l.Value.Type != lex.EOF {
        t.Errorf("Backing channel still not closed")
    }
}
