package assembler

import (
	"bufio"
	"fmt"
	"github.com/JakubCygaro/alphataurus/internal/vm"
	"math"
	"strconv"
	"unicode"
)

const (
	TOKEN_TNIL = iota
	TOKEN_TIDENT
	TOKEN_TCOMMA
	TOKEN_TREG
	TOKEN_TINTEGER_LIT
	TOKEN_TFLOAT_LIT
	TOKEN_TUNSIGNED
	TOKEN_TSIGNED
	TOKEN_TFLOAT
	TOKEN_TMINUS
	TOKEN_TDOT
	TOKEN_TEOF
)

var keywords = map[string]int{
	"SIGNED":   TOKEN_TSIGNED,
	"UNSIGNED": TOKEN_TUNSIGNED,
	"FLOAT":    TOKEN_TFLOAT,
}

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
func (l *Lexer) CurrentPosition() string {
	return fmt.Sprintf("(line: %d, column: %d)", l.line, l.col)
}
func (l *Lexer) CurrentToken() Token {
	return l.currentToken
}
func numberCheck(b byte) bool {
	return b-'0' <= 9
}
func identCheck(b byte) bool {
	return b == '_' ||
		b-'a' <= 'z'-'a' ||
		b-'A' <= 'Z'-'A'
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
	switch {
	case b == ',':
		l.currentToken = Token{
			Ty: TOKEN_TCOMMA,
		}
	case b == '.':
		next, err := l.reader.ReadByte()
		if err != nil {
			return err
		}
		if numberCheck(next) {
			l.reader.UnreadByte()
			err := l.readDigit(b)
			if err != nil {
				return err
			}
		} else {
			l.currentToken = Token {
				Ty: TOKEN_TDOT,
			}
		}
	case b == '-':
		next, err := l.reader.ReadByte()
		if err != nil {
			return err
		}
		if numberCheck(next) {
			l.reader.UnreadByte()
			err := l.readDigit(b)
			if err != nil {
				return err
			}
		} else {
			l.currentToken = Token {
				Ty: TOKEN_TMINUS,
			}
		}
	case numberCheck(b):
		err := l.readDigit(b)
		if err != nil {
			return err
		}
	case identCheck(b):
		buf := make([]byte, 0, 16)
		buf = append(buf, b)
		for {
			next, err := l.reader.ReadByte()
			l.col++
			if err != nil {
				break
			}
			if identCheck(next) || numberCheck(next) {
				buf = append(buf, next)
			} else if unicode.IsSpace(rune(next)) || unicode.IsPunct(rune(next)) {
				l.reader.UnreadByte()
				l.col--
				break
			} else {
				return fmt.Errorf("Malformed identifier %s", l.CurrentPosition())
			}
		}
		val := string(buf)
		if reg, ok := recognizeRegister(val); ok {
			l.currentToken = Token{
				Ty:  TOKEN_TREG,
				val: reg,
			}
		} else if kwd, ok := keywords[val]; ok {
			l.currentToken = Token{
				Ty: kwd,
			}
		} else {
			l.currentToken = Token{
				Ty:  TOKEN_TIDENT,
				val: val,
			}
		}
	default:
		return fmt.Errorf("Unrecognized character %s", l.CurrentPosition())
	}
	return nil
}

func (l *Lexer) readDigit(b byte) error {
	buf := make([]byte, 0, 16)
	buf = append(buf, b)
	dot := b == '.'
	minus := b == '-'
	for {
		next, err := l.reader.ReadByte()
		l.col++
		if err != nil {
			break
		}
		if numberCheck(next) {
			buf = append(buf, next)
		} else if next == '.' && !dot {
			buf = append(buf, next)
			dot = true
		} else if unicode.IsSpace(rune(next)) || !identCheck(next) {
			l.reader.UnreadByte()
			l.col--
			break
		} else {
			return fmt.Errorf("Malformed integer literal %s", l.CurrentPosition())
		}
	}
	if !dot {
		if !minus {
			val, err := strconv.ParseUint(string(buf), 10, 64)
			if err != nil {
				return err
			}
			l.currentToken = Token{
				Ty:  TOKEN_TINTEGER_LIT,
				val: uint64(val),
			}
		} else {
			val, err := strconv.ParseInt(string(buf), 10, 64)
			if err != nil {
				return err
			}
			l.currentToken = Token{
				Ty:  TOKEN_TINTEGER_LIT,
				val: uint64(val),
			}
		}
	} else {
		val, err := strconv.ParseFloat(string(buf), 64)
		if err != nil {
			return fmt.Errorf("Malformed 64-bit floating point digit literal %s", l.CurrentPosition())
		}
		l.currentToken = Token{
			Ty:  TOKEN_TFLOAT_LIT,
			val: uint64(math.Float64bits(val)),
		}
	}
	return nil
}

func recognizeRegister(s string) (int, bool) {
	if len(s) < 2 {
		return -1, false
	}
	if s[0] == 'r' && numberCheck(s[1]) && s[1]-'0' <= vm.GP_REG_MAX {
		return int(s[1] - '0'), true
	}
	switch s {
	case "sp":
		return vm.SP_IDX, true
	case "bp":
		return vm.BP_IDX, true
	case "ip":
		return vm.IP_IDX, true
	}
	return -1, false
}
