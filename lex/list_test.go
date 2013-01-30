package lex

import "testing"

func TestLLBasic(t *testing.T) {
    ch := make(chan *Eme)
    go func() {
        ch <- &Eme{Atom, "1", nil}
        ch <- &Eme{Atom, "2", nil}
        ch <- &Eme{Atom, "3", nil}
        close(ch)
    }()

    l := NewList(ch)
    if l.Value.Type != Atom || l.Value.Content != "1" {
        t.Errorf("Wrong lexeme: %s", l.Value.Content)
    }

    l = l.Next()
    if l.Value.Type != Atom || l.Value.Content != "2" {
        t.Errorf("Wrong lexeme: %s", l.Value.Content)
    }

    l = l.Next()
    if l.Value.Type != Atom || l.Value.Content != "3" {
        t.Errorf("Wrong lexeme: %s", l.Value.Content)
    }

    l = l.Next()
    if l.Value.Type != EOF {
        t.Errorf("Backing channel not closed")
    }

    l = l.Next()
    if l.Value.Type != EOF {
        t.Errorf("Backing channel still not closed")
    }
}
