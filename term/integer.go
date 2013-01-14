package term

import "fmt"
import "math/big"

// Integer represents an unbounded, signed integer value
type Integer big.Int

// NewInt parses an integer's string representation to create a new
// integer value. Panics if the string's is not a valid integer
func NewInt(text string) *Integer {
    if len(text) == 0 {
        panic("Empty string is not a valid integer")
    }

    // see §6.4.4 for syntax details
    if text[0] == '0' && len(text) >= 3 {
        switch text[1] {
            case '\'':
                return parseEscape(text[2:])
            case 'b':
                return parseInteger("%b", text[2:])
            case 'o':
                return parseInteger("%o", text[2:])
            case 'x':
                return parseInteger("%x", text[2:])
            default:
                return parseInteger("%d", text)
        }
    }
    return parseInteger("%d", text)
}

func parseInteger(format, text string) *Integer {
    i := new(big.Int)
    n, err := fmt.Sscanf(text, format, i)
    maybePanic(err)
    if n == 0 {
        panic("Parsed no integers")
    }

    return (*Integer)(i)
}

// see "single quoted character" - §6.4.2.1
func parseEscape(text string) *Integer {
    var r rune
    if text[0] == '\\' {
        if len(text) < 2 {
            msg := fmt.Sprintf("Invalid integer character constant: %s", text)
            panic(msg)
        }
        switch text[1] {
            // "meta escape sequence" - §6.4.2.1
            case '\\': r = '\\'
            case '\'': r = '\''
            case '"': r = '"'
            case '`': r = '`'

            // "control escape char" - §6.4.2.1
            case 'a': r = '\a'
            case 'b': r = '\b'
            case 'f': r = '\f'
            case 'n': r = '\n'
            case 'r': r = '\r'
            case 's': r = ' '  // SWI-Prolog extension
            case 't': r = '\t'
            case 'v': r = '\v'

            // "hex escape char" - §6.4.2.1
            case 'x':
                return parseInteger("%x", text[2:len(text)-1])

            // "octal escape char" - §6.4.2.1
            case '0','1','2','3','4','5','6','7':
                return parseInteger("%o", text[1:len(text)-1])

            // unexpected escape sequence
            default:
                msg := fmt.Sprintf("Invalid character escape sequence: %s", text)
                panic(msg)
        }
    } else {
        // "non quote char" - §6.4.2.1
        runes := []rune(text)
        r = runes[0]
    }
    code := int64(r)
    return (*Integer)(big.NewInt(code))
}

func (self *Integer) Value() *big.Int {
    return (*big.Int)(self)
}

func (self *Integer) String() string {
    return self.Value().String()
}

func (self *Integer) Functor() string {
    panic("Variables have no Functor()")
}
func (self *Integer) Arity() int {
    panic("Variables have no Arity()")
}
func (self *Integer) Arguments() []Term {
    panic("Variables have no Arguments()")
}
func (self *Integer) Body() Term {
    panic("Variables have no Body()")
}
func (self *Integer) Head() Term {
    panic("Variables have no Head()")
}
func (self *Integer) IsClause() bool {
    return false
}
func (self *Integer) Indicator() string {
    return self.String()
}
func (self *Integer) Error() error {
    panic("Can't call Error() on a Variable")
}
