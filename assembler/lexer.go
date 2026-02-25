package assembler

import (
	"bufio"
	"fmt"
	"math"
	"unicode"
)

const (
	TOKEN_TNIL = iota
	TOKEN_TIDENT
	TOKEN_TCOMMA
	TOKEN_TREG
	TOKEN_TINTEGER
	TOKEN_TEOF
)

type Token struct {
	Ty  int
	val any
}

type Lexer struct {
	head, col, line uint64
	currentToken    Token
	reader          bufio.Reader
}

func nilToken() Token {
	return Token{
		Ty: TOKEN_TNIL,
	}
}

func NewLexer(reader bufio.Reader) Lexer {
	return Lexer{
		currentToken: nilToken(),
		reader:       reader,
	}
}
func (l *Lexer) currentPosition() string {
	return fmt.Sprintf("(line: %d, column: %d)", l.line, l.col)
}
func (l *Lexer) CurrentToken() Token {
	return l.currentToken
}
func (l *Lexer) ReadNextToken() error {
	var b byte
	for {
		var err error
		b, err = l.reader.ReadByte()
		l.col++
		if err != nil {
			l.currentToken = Token{
				Ty: TOKEN_TEOF,
			}
			return nil
		}
		if b == '\n' {
			l.col = 0
			l.line++
		}
		if !unicode.IsSpace(rune(b)) {
			break
		}
	}
	numberCheck := func(b byte) bool {
		return b-'0' <= 9
	}
	identCheck := func(b byte) bool {
		return b == '_' ||
			b-'a' <= 'z'-'a' ||
			b-'A' <= 'Z'-'A'
	}
	switch {
	case b == ',':
		l.currentToken = Token{
			Ty: TOKEN_TCOMMA,
		}
	case numberCheck(b):
		buf := make([]byte, 0)
		buf = append(buf, b)
		for {
			next, err := l.reader.ReadByte()
			l.col++
			if err != nil {
				break
			}
			if numberCheck(next) {
				buf = append(buf, next)
			} else if unicode.IsSpace(rune(next)) {
				l.reader.UnreadByte()
				l.col--
				break
			} else {
				return fmt.Errorf("Malformed integer literal %s", l.currentPosition())
			}
		}
		val, idx := 0, len(buf)-1
		for _, elem := range buf {
			mul := int(math.Pow(10, float64(idx)))
			val += int(elem-'0') * mul
			idx--
		}
		l.currentToken = Token{
			Ty:  TOKEN_TINTEGER,
			val: val,
		}
	case identCheck(b):
		buf := make([]byte, 16)
		buf = append(buf, b)
		for {
			next, err := l.reader.ReadByte()
			l.col++
			if err != nil {
				break
			}
			if identCheck(next) || numberCheck(next) {
				buf = append(buf, next)
			} else if unicode.IsSpace(rune(next)) {
				l.reader.UnreadByte()
				l.col--
				break
			} else {
				return fmt.Errorf("Malformed identifier %s", l.currentPosition())
			}
		}
		val := string(buf[:])
		l.currentToken = Token{
			Ty:  TOKEN_TIDENT,
			val: val,
		}
	default:
		return fmt.Errorf("Unrecognized character %s", l.currentPosition())
	}
	return nil
}
