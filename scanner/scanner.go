// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2013 Michael Hendricks. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package scanner tokenizes UTF-8-encoded Prolog text.
// It takes an io.Reader providing the source, which then can be tokenized
// through repeated calls to the Scan function.  For compatibility with
// existing tools, the NUL character is not allowed. If the first character
// in the source is a UTF-8 encoded byte order mark (BOM), it is discarded.
//
// Basic usage pattern:
//
//    lexemes := scanner.Scan(file)
//    for lexeme := range lexemes {
//        // do something with lexeme
//    }
//
package scanner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"unicode"
	"unicode/utf8"
)

// A lexeme encapsulating it's type and content
type Lexeme struct {
	Type	rune	// EOF, Atom, Comment, etc.
	Content	string
}

// Scan tokenizes src in a separate goroutine sending lexemes down a
// channel as they become available.  The channel is closed on EOF.
func Scan(src io.Reader) <-chan *Lexeme {
	ch := make(chan *Lexeme)
	go func () {
		s := new(Scanner).Init(src)
		tok := s.Scan()
		for tok != EOF {
			l := &Lexeme{
				Type:		tok,
				Content:	s.TokenText(),
			}
			ch <- l
			tok = s.Scan()
		}
		close(ch)
	}()
	return ch
}

// A source position is represented by a Position value.
// A position is valid if Line > 0.
type Position struct {
	Filename string // filename, if any
	Offset   int    // byte offset, starting at 0
	Line     int    // line number, starting at 1
	Column   int    // column number, starting at 1 (character count per line)
}

// IsValid returns true if the position is valid.
func (pos *Position) IsValid() bool { return pos.Line > 0 }

func (pos Position) String() string {
	s := pos.Filename
	if pos.IsValid() {
		if s != "" {
			s += ":"
		}
		s += fmt.Sprintf("%d:%d", pos.Line, pos.Column)
	}
	if s == "" {
		s = "???"
	}
	return s
}

// The result of Scan is one of these tokens or a Unicode character.
const (
	EOF = -(iota + 1)  // reached end of source
	Atom               // a Prolog atom, possibly quoted
	Comment            // a comment
	Float              // a floating point number
	Functor            // an atom used as a predicate functor
	FullStop           // "." ending a term
	Int                // an integer
	String             // a double-quoted string
	Variable           // a Prolog variable
	Void               // the special "_" variable
)

var tokenString = map[rune]string{
	EOF:       "EOF",
	Atom:      "Atom",
	Comment:   "Comment",
	Float:     "Float",
	Functor:   "Functor",
	FullStop:  "FullStop",
	Int:       "Int",
	String:    "String",
	Variable:  "Variable",
	Void:      "Void",
}

// TokenString returns a printable string for a token or Unicode character.
func TokenString(tok rune) string {
	if s, found := tokenString[tok]; found {
		return s
	}
	return fmt.Sprintf("%q", string(tok))
}

const bufLen = 1024 // at least utf8.UTFMax

// A Scanner implements reading of Unicode characters and tokens from an io.Reader.
type Scanner struct {
	// Input
	src io.Reader

	// Source buffer
	srcBuf [bufLen + 1]byte // +1 for sentinel for common case of s.next()
	srcPos int              // reading position (srcBuf index)
	srcEnd int              // source end (srcBuf index)

	// Source position
	srcBufOffset int // byte offset of srcBuf[0] in source
	line         int // line count
	column       int // character count
	lastLineLen  int // length of last line in characters (for correct column reporting)
	lastCharLen  int // length of last character in bytes

	// Token text buffer
	// Typically, token text is stored completely in srcBuf, but in general
	// the token text's head may be buffered in tokBuf while the token text's
	// tail is stored in srcBuf.
	tokBuf bytes.Buffer // token text head that is not in srcBuf anymore
	tokPos int          // token text tail position (srcBuf index); valid if >= 0
	tokEnd int          // token text tail end (srcBuf index)

	// One character look-ahead
	ch rune // character before current srcPos

	// Error is called for each error encountered. If no Error
	// function is set, the error is reported to os.Stderr.
	Error func(s *Scanner, msg string)

	// ErrorCount is incremented by one for each error encountered.
	ErrorCount int

	// Start position of most recently scanned token; set by Scan.
	// Calling Init or Next invalidates the position (Line == 0).
	// The Filename field is always left untouched by the Scanner.
	// If an error is reported (via Error) and Position is invalid,
	// the scanner is not inside a token. Call Pos to obtain an error
	// position in that case.
	Position
}

// Init initializes a Scanner with a new source and returns s.
// Error is set to nil, ErrorCount is set to 0
func (s *Scanner) Init(src io.Reader) *Scanner {
	s.src = src

	// initialize source buffer
	// (the first call to next() will fill it by calling src.Read)
	s.srcBuf[0] = utf8.RuneSelf // sentinel
	s.srcPos = 0
	s.srcEnd = 0

	// initialize source position
	s.srcBufOffset = 0
	s.line = 1
	s.column = 0
	s.lastLineLen = 0
	s.lastCharLen = 0

	// initialize token text buffer
	// (required for first call to next()).
	s.tokPos = -1

	// initialize one character look-ahead
	s.ch = -1 // no char read yet

	// initialize public fields
	s.Error = nil
	s.ErrorCount = 0
	s.Line = 0 // invalidate token position

	return s
}

// next reads and returns the next Unicode character. It is designed such
// that only a minimal amount of work needs to be done in the common ASCII
// case (one test to check for both ASCII and end-of-buffer, and one test
// to check for newlines).
func (s *Scanner) next() rune {
	ch, width := rune(s.srcBuf[s.srcPos]), 1

	if ch >= utf8.RuneSelf {
		// uncommon case: not ASCII or not enough bytes
		for s.srcPos+utf8.UTFMax > s.srcEnd && !utf8.FullRune(s.srcBuf[s.srcPos:s.srcEnd]) {
			// not enough bytes: read some more, but first
			// save away token text if any
			if s.tokPos >= 0 {
				s.tokBuf.Write(s.srcBuf[s.tokPos:s.srcPos])
				s.tokPos = 0
				// s.tokEnd is set by Scan()
			}
			// move unread bytes to beginning of buffer
			copy(s.srcBuf[0:], s.srcBuf[s.srcPos:s.srcEnd])
			s.srcBufOffset += s.srcPos
			// read more bytes
			// (an io.Reader must return io.EOF when it reaches
			// the end of what it is reading - simply returning
			// n == 0 will make this loop retry forever; but the
			// error is in the reader implementation in that case)
			i := s.srcEnd - s.srcPos
			n, err := s.src.Read(s.srcBuf[i:bufLen])
			s.srcPos = 0
			s.srcEnd = i + n
			s.srcBuf[s.srcEnd] = utf8.RuneSelf // sentinel
			if err != nil {
				if s.srcEnd == 0 {
					if s.lastCharLen > 0 {
						// previous character was not EOF
						s.column++
					}
					s.lastCharLen = 0
					return EOF
				}
				if err != io.EOF {
					s.error(err.Error())
				}
				// If err == EOF, we won't be getting more
				// bytes; break to avoid infinite loop. If
				// err is something else, we don't know if
				// we can get more bytes; thus also break.
				break
			}
		}
		// at least one byte
		ch = rune(s.srcBuf[s.srcPos])
		if ch >= utf8.RuneSelf {
			// uncommon case: not ASCII
			ch, width = utf8.DecodeRune(s.srcBuf[s.srcPos:s.srcEnd])
			if ch == utf8.RuneError && width == 1 {
				// advance for correct error position
				s.srcPos += width
				s.lastCharLen = width
				s.column++
				s.error("illegal UTF-8 encoding")
				return ch
			}
		}
	}

	// advance
	s.srcPos += width
	s.lastCharLen = width
	s.column++

	// special situations
	switch ch {
	case 0:
		// for compatibility with other tools
		s.error("illegal character NUL")
	case '\n':
		s.line++
		s.lastLineLen = s.column
		s.column = 0
	}

	return ch
}

// Next reads and returns the next Unicode character.
// It returns EOF at the end of the source. It reports
// a read error by calling s.Error, if not nil; otherwise
// it prints an error message to os.Stderr. Next does not
// update the Scanner's Position field; use Pos() to
// get the current position.
func (s *Scanner) Next() rune {
	s.tokPos = -1 // don't collect token text
	s.Line = 0    // invalidate token position
	ch := s.Peek()
	s.ch = s.next()
	return ch
}

// Peek returns the next Unicode character in the source without advancing
// the scanner. It returns EOF if the scanner's position is at the last
// character of the source.
func (s *Scanner) Peek() rune {
	if s.ch < 0 {
		// this code is only run for the very first character
		s.ch = s.next()
		if s.ch == '\uFEFF' {
			s.ch = s.next() // ignore BOM
		}
	}
	return s.ch
}

func (s *Scanner) error(msg string) {
	s.ErrorCount++
	if s.Error != nil {
		s.Error(s, msg)
		return
	}
	pos := s.Position
	if !pos.IsValid() {
		pos = s.Pos()
	}
	fmt.Fprintf(os.Stderr, "%s: %s\n", pos, msg)
}

func (s *Scanner) scanAlphanumeric(ch rune) rune {
	for isAlphanumeric(ch) {
		ch = s.next()
	}
	return ch
}

func (s *Scanner) scanGraphic(ch rune) rune {
	for isGraphic(ch) {
		ch = s.next()
	}
	return ch
}

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= ch && ch <= 'f':
		return int(ch - 'a' + 10)
	case 'A' <= ch && ch <= 'F':
		return int(ch - 'A' + 10)
	}
	return 16 // larger than any legal digit val
}

func isDecimal(ch rune) bool { return '0' <= ch && ch <= '9' }

// true if the rune is a graphic token char per ISO §6.4.2
func isGraphic(ch rune) bool {
	return isOneOf(ch, `#$&*+-./:<=>?@^\~`)
}

// ISO §6.5.2 "alphanumeric char" extended to Unicode
func isAlphanumeric(ch rune) bool {
	if ch == '_' || unicode.IsLetter(ch) || unicode.IsDigit(ch) {
		return true
	}
	return false
}

// true if the rune is a valid start for a variable
func isVariableStart(ch rune) bool {
	return ch == '_' || unicode.IsUpper(ch)
}

func isOneOf(ch rune, chars string) bool {
	for _, allowed := range chars {
		if ch == allowed {
			return true
		}
	}
	return false
}

func isSolo(ch rune) bool { return ch == '!' || ch == ';' }

func (s *Scanner) scanMantissa(ch rune) rune {
	for isDecimal(ch) {
		ch = s.next()
	}
	return ch
}

func (s *Scanner) scanFraction(ch rune) rune {
	if ch == '.' {
		ch = s.scanMantissa(s.next())
	}
	return ch
}

func (s *Scanner) scanExponent(ch rune) rune {
	if ch == 'e' || ch == 'E' {
		ch = s.next()
		if ch == '-' || ch == '+' {
			ch = s.next()
		}
		ch = s.scanMantissa(ch)
	}
	return ch
}

func (s *Scanner) scanNumber(ch rune) (rune, rune) {
	// isDecimal(ch)
	if ch == '0' {
		// int or float
		ch = s.next()
		switch ch {
		case 'x', 'X':
			// hexadecimal int
			ch = s.next()
			hasMantissa := false
			for digitVal(ch) < 16 {
				ch = s.next()
				hasMantissa = true
			}
			if !hasMantissa {
				s.error("illegal hexadecimal number")
			}
		case '\'':
			ch = s.next()
			if ch == '\\' {
				ch = s.scanEscape('\'')
			} else {
				ch = s.next()
			}
		default:
			// octal int or float
			has8or9 := false
			for isDecimal(ch) {
				if ch > '7' {
					has8or9 = true
				}
				ch = s.next()
			}
			if ch == '.' || ch == 'e' || ch == 'E' {
				// float
				ch = s.scanFraction(ch)
				ch = s.scanExponent(ch)
				return Float, ch
			}
			// octal int
			if has8or9 {
				s.error("illegal octal number")
			}
		}
		return Int, ch
	}
	// decimal int or float
	ch = s.scanMantissa(ch)
	if ch == '.' || ch == 'e' || ch == 'E' {
		// float
		ch = s.scanFraction(ch)
		ch = s.scanExponent(ch)
		return Float, ch
	}
	return Int, ch
}

func (s *Scanner) scanDigits(ch rune, base, n int) rune {
	for n > 0 && digitVal(ch) < base {
		ch = s.next()
		n--
	}
	if n > 0 {
		s.error("illegal char escape")
	}
	return ch
}

func (s *Scanner) scanEscape(quote rune) rune {
	ch := s.next() // read character after '/'
	switch ch {
	case 'a', 'b', 'f', 'n', 'r', 's', 't', 'v', '\\', quote:
		// nothing to do
		ch = s.next()
	case '0', '1', '2', '3', '4', '5', '6', '7':
		ch = s.scanDigits(ch, 8, 3)
	case 'x':
		ch = s.scanDigits(s.next(), 16, 2)
	case 'u':
		ch = s.scanDigits(s.next(), 16, 4)
	case 'U':
		ch = s.scanDigits(s.next(), 16, 8)
	default:
		s.error("illegal char escape")
	}
	return ch
}

func (s *Scanner) scanString(quote rune) (n int) {
	ch := s.next() // read character after quote
	for ch != quote {
		if ch == '\n' || ch < 0 {
			s.error("literal not terminated")
			return
		}
		if ch == '\\' {
			ch = s.scanEscape(quote)
		} else {
			ch = s.next()
		}
		n++
	}
	return
}

func (s *Scanner) scanComment(ch rune) rune {
	// ch == '%' || ch == '*'
	if ch == '%' {
		// line comment
		ch = s.next() // read character after "%"
		for ch != '\n' && ch >= 0 {
			ch = s.next()
		}
		return ch
	}

	// general comment.  See Note1
	depth := 1
	ch = s.next() // read character after "/*"
	for depth > 0 {
		if ch < 0 {
			s.error("comment not terminated")
			break
		}
		ch0 := ch
		ch = s.next()
		if ch0 == '*' && ch == '/' {
			ch = s.next()
			depth--
		} else if ch0 == '/' && ch == '*' {
			ch = s.next()
			depth++
		}
	}
	return ch
}
// Note1: Nested comments are prohibited by ISO Prolog §6.4.1.  To wit,
// "The comment text of a bracketed comment shall not contain the comment
// close sequence."  However, nested comments are ridiculously practical
// during debugging and development, so I've chosen to deviate by being
// more permissive than is strictly allowed.  SWI-Prolog does the same thing.

// Scan reads the next token or Unicode character from source and returns it.
// It returns EOF at the end of the source. It reports scanner errors (read and
// token errors) by calling s.Error, if not nil; otherwise it prints an error
// message to os.Stderr.
func (s *Scanner) Scan() rune {
	ch := s.Peek()

	// reset token text position
	s.tokPos = -1
	s.Line = 0

	// skip white space
	for unicode.IsSpace(ch) {
		ch = s.next()
	}

	// start collecting token text
	s.tokBuf.Reset()
	s.tokPos = s.srcPos - s.lastCharLen

	// set token position
	// (this is a slightly optimized version of the code in Pos())
	s.Offset = s.srcBufOffset + s.tokPos
	if s.column > 0 {
		// common case: last character was not a '\n'
		s.Line = s.line
		s.Column = s.column
	} else {
		// last character was a '\n'
		// (we cannot be at the beginning of the source
		// since we have called next() at least once)
		s.Line = s.line - 1
		s.Column = s.lastLineLen
	}

	// determine token value
	tok := ch
	switch {
	case ch == '/':		// '/' can start a comment or an atom
		ch = s.next()
		if ch == '*' {
			ch = s.scanComment(ch)
			tok = Comment
		} else {
			tok = Atom
			ch = s.scanGraphic(ch)
			if ch == '(' { tok = Functor }
		}
	case isGraphic(ch):
		ch = s.next()
		tok = Atom
		ch = s.scanGraphic(ch)
		if ch == '(' { tok = Functor }
	case isSolo(ch):
		tok = Atom
		ch = s.next()
	case unicode.IsLower(ch):  // name by "letter digit token" rule §6.4.2 w/ Unicode
		tok = Atom
		ch = s.next()
		ch = s.scanAlphanumeric(ch)
		if ch == '(' { tok = Functor }
	case isVariableStart(ch):
		tok = Variable
		ch = s.next()
		ch = s.scanAlphanumeric(ch)  // variables look like atoms after the start
	case isDecimal(ch):
		tok, ch = s.scanNumber(ch)
	default:
		switch ch {
		case '"':
			s.scanString('"')
			tok = String
			ch = s.next()
		case '\'':
			s.scanString('\'')
			tok = Atom
			ch = s.next()
		case '%':
			ch = s.scanComment(ch)
			tok = Comment
		default:
			ch = s.next()
		}
	}

	// end of token text
	s.tokEnd = s.srcPos - s.lastCharLen

	s.ch = ch

	// last minute specializations
	switch tok {
		case Atom:
			switch s.TokenText() {
				case ".": return FullStop
			}
		case Variable:
			switch s.TokenText() {
				case "_": return Void
			}
	}
	return tok
}

// Pos returns the position of the character immediately after
// the character or token returned by the last call to Next or Scan.
func (s *Scanner) Pos() (pos Position) {
	pos.Filename = s.Filename
	pos.Offset = s.srcBufOffset + s.srcPos - s.lastCharLen
	switch {
	case s.column > 0:
		// common case: last character was not a '\n'
		pos.Line = s.line
		pos.Column = s.column
	case s.lastLineLen > 0:
		// last character was a '\n'
		pos.Line = s.line - 1
		pos.Column = s.lastLineLen
	default:
		// at the beginning of the source
		pos.Line = 1
		pos.Column = 1
	}
	return
}

// TokenText returns the string corresponding to the most recently scanned token.
// Valid after calling Scan().
func (s *Scanner) TokenText() string {
	if s.tokPos < 0 {
		// no token text
		return ""
	}

	if s.tokEnd < 0 {
		// if EOF was reached, s.tokEnd is set to -1 (s.srcPos == 0)
		s.tokEnd = s.tokPos
	}

	if s.tokBuf.Len() == 0 {
		// common case: the entire token text is still in srcBuf
		return string(s.srcBuf[s.tokPos:s.tokEnd])
	}

	// part of the token text was saved in tokBuf: save the rest in
	// tokBuf as well and return its content
	s.tokBuf.Write(s.srcBuf[s.tokPos:s.tokEnd])
	s.tokPos = s.tokEnd // ensure idempotency of TokenText() call
	return s.tokBuf.String()
}
