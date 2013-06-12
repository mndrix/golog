package lex

// An immutable list of lexemes which populates its tail by reading
// lexemes from a channel, such as that provided by Scan()
type List struct {
	Value *Eme
	next  *List
	src   <-chan *Eme
}

// NewLexemList returns a new lexeme list which pulls lexemes from
// the given source channel.  Creating a new list consumes one lexeme
// from the source channel.
func NewList(src <-chan *Eme) *List {
	lexeme, ok := <-src
	if !ok {
		lexeme = &Eme{Type: EOF}
	}
	return &List{
		Value: lexeme,
		next:  nil,
		src:   src,
	}
}

// Next returns the next element in the lexeme list, pulling a lexeme
// from the source channel, if necessary
func (self *List) Next() *List {
	if self.next == nil {
		next := NewList(self.src)
		self.next = next
	}
	return self.next
}
