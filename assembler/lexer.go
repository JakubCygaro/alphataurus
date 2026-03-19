package assembler

import (
	"bufio"
	"fmt"
	// "io"
	"math"
	"strconv"
	"unicode"

	"github.com/JakubCygaro/alphataurus/assembler/errors"
	"github.com/JakubCygaro/alphataurus/internal/vm"
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
	//Ops
	TOKEN_TMINUS
	TOKEN_TPLUS
	TOKEN_TASTERISK
	TOKEN_TSLASH
	//
	TOKEN_TDOT
	TOKEN_TOPENBRACKET
	TOKEN_TCLOSEDBRACKET
	TOKEN_TOPENPAREN
	TOKEN_TCLOSEDPAREN
	TOKEN_TCOLON
	TOKEN_TNEWLINE
	TOKEN_TSINGLEQ
	TOKEN_TAT
	TOKEN_TEOF
)
const (
	TOKEN_OPSTART = TOKEN_TMINUS
	TOKEN_OPEND   = TOKEN_TSLASH
)

var keywords = map[string]int{
	"SIGNED":   TOKEN_TSIGNED,
	"UNSIGNED": TOKEN_TUNSIGNED,
	"FLOAT":    TOKEN_TFLOAT,
}

type Token struct {
	Ty        int
	Val       any
	Col, Line uint64
}

type Lexer struct {
	head, col, line   uint64
	currentToken      Token
	unRead            bool
	reader            bufio.Reader
	lastCol, lastLine uint64
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
		unRead:       false,
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
func hexNumberCheck(b byte) bool {
	return numberCheck(b) || b-'a' <= 'f'-'a' || b-'A' <= 'F'-'A'
}
func binaryNumberCheck(b byte) bool {
	return b-'0' <= 1
}
func identCheck(b byte) bool {
	return b == '_' ||
		b-'a' <= 'z'-'a' ||
		b-'A' <= 'Z'-'A'
}
func (l *Lexer) Expect(tokenType int) (Token, bool) {
	l.ReadNextToken()
	if l.currentToken.Ty != tokenType {
		return l.currentToken, false
	}
	return l.currentToken, true
}
func (l *Lexer) readByte() (byte, error) {
	b, err := l.reader.ReadByte()
	l.lastLine, l.lastCol = l.line, l.col
	if b == '\n' {
		l.line++
		l.col = 0
	} else {
		l.col++
	}
	return b, err
}
func (l *Lexer) unreadByte() error {
	l.line, l.col = l.lastLine, l.lastCol
	return l.reader.UnreadByte()
}
func (l *Lexer) UnreadToken() {
	l.unRead = true
}
func (l *Lexer) ReadNextToken() error {
	if l.unRead {
		l.unRead = false
		return nil
	}
	var b byte
	for {
		var err error
		b, err = l.readByte()
		if err != nil {
			l.currentToken = Token{
				Ty: TOKEN_TEOF,
			}
			return nil
		}
		if b == '\n' {
			break
		}
		if !unicode.IsSpace(rune(b)) {
			break
		}
	}
	l.currentToken.Col, l.currentToken.Line = l.col, l.line
	switch {
	case b == '\'':
		if err := l.readSingleQuoted(); err != nil{
			return err
		}
	case b == '@':
		l.currentToken = Token{
			Ty:  TOKEN_TAT,
			Val: rune(b),
		}
	case b == ':':
		l.currentToken = Token{
			Ty:  TOKEN_TCOLON,
			Val: rune(b),
		}
	case b == '(':
		l.currentToken = Token{
			Ty:  TOKEN_TOPENPAREN,
			Val: rune(b),
		}
	case b == ')':
		l.currentToken = Token{
			Ty:  TOKEN_TCLOSEDPAREN,
			Val: rune(b),
		}
	case b == '[':
		l.currentToken = Token{
			Ty:  TOKEN_TOPENBRACKET,
			Val: rune(b),
		}
	case b == ']':
		l.currentToken = Token{
			Ty:  TOKEN_TCLOSEDBRACKET,
			Val: rune(b),
		}
	case b == '\n':
		l.currentToken = Token{
			Ty:  TOKEN_TNEWLINE,
			Val: rune(b),
		}
	case b == ',':
		l.currentToken = Token{
			Ty:  TOKEN_TCOMMA,
			Val: rune(b),
		}
	case b == '+':
		l.currentToken = Token{
			Ty:  TOKEN_TPLUS,
			Val: rune(b),
		}
	case b == '-':
		l.currentToken = Token{
			Ty:  TOKEN_TMINUS,
			Val: rune(b),
		}
	case b == '*':
		l.currentToken = Token{
			Ty:  TOKEN_TASTERISK,
			Val: rune(b),
		}
	case b == '/':
		l.currentToken = Token{
			Ty:  TOKEN_TSLASH,
			Val: rune(b),
		}
	case b == '.':
		next, err := l.readByte()
		if err != nil {
			return err
		}
		if numberCheck(next) {
			l.unreadByte()
			err := l.readDigit(b)
			if err != nil {
				return err
			}
		} else {
			l.currentToken = Token{
				Ty:  TOKEN_TDOT,
				Val: rune(b),
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
			next, err := l.readByte()
			if err != nil {
				break
			}
			if identCheck(next) || numberCheck(next) {
				buf = append(buf, next)
			} else {
				l.unreadByte()
				break
			}
		}
		val := string(buf)
		if reg, ok := recognizeRegister(val); ok {
			l.currentToken = Token{
				Ty:  TOKEN_TREG,
				Val: reg,
			}
		} else if kwd, ok := keywords[val]; ok {
			l.currentToken = Token{
				Ty:  kwd,
				Val: val,
			}
		} else {
			l.currentToken = Token{
				Ty:  TOKEN_TIDENT,
				Val: val,
			}
		}
	default:
		return errors.UnrecognizedChar(rune(b), l.line, l.col)
	}
	return nil
}
func (l *Lexer) readSingleQuoted() error {
	buf := make([]byte, 0, 64)
	for {
		next, err := l.readByte()
		if err != nil {
			return errors.UnclosedSingleQuote(l.line, l.col)
		}
		if next == '\'' {
			break
		} else {
			buf = append(buf, next)
		}
	}
	l.currentToken = Token{
		Ty:  TOKEN_TSINGLEQ,
		Val: string(buf),
	}
	return nil
}

func (l *Lexer) readDigit(b byte) error {
	buf := make([]byte, 0, 16)
	buf = append(buf, b)
	dot := b == '.'
	e := false
	startedWithZero := b == '0'
	hex := false
	binary := false
	if startedWithZero {
		next, err := l.readByte()
		if err != nil {
			return err
		}
		switch next {
		case 'x':
			hex = true
			dot = true
		case 'b':
			binary = true
			dot = true
		default:
			l.unreadByte()
		}
	}
	for {
		next, err := l.readByte()
		if err != nil {
			break
		}
		if !hex && !binary && numberCheck(next) {
			buf = append(buf, next)
		} else if hex && hexNumberCheck(next) {
			buf = append(buf, next)
		} else if binary && binaryNumberCheck(next) {
			buf = append(buf, next)
		} else if next == '.' && !dot {
			buf = append(buf, next)
			dot = true
		} else if (next == 'e' || next == 'E') && !e && !hex && !binary {
			buf = append(buf, next)
			e = true
			next, err := l.readByte()
			if err != nil {
				return err
			}
			if next == '-' || next == '+' {
				buf = append(buf, next)
				next, err = l.readByte()
				if err != nil {
					return err
				}
			}
			if numberCheck(next) {
				l.unreadByte()
			} else {
				return errors.MalformedFloatLit(l.line, l.col)
			}
		} else if unicode.IsSpace(rune(next)) || !identCheck(next) {
			l.unreadByte()
			break
		} else {
			return errors.MalformedIntegerLit(l.line, l.col)
		}
	}
	if !dot {
		val, err := strconv.ParseUint(string(buf), 10, 64)
		if err != nil {
			return errors.MalformedIntegerLit(l.line, l.col)
		}
		l.currentToken = Token{
			Ty:  TOKEN_TINTEGER_LIT,
			Val: uint64(val),
		}
	} else if hex {
		val, err := strconv.ParseUint(string(buf), 16, 64)
		if err != nil {
			return errors.MalformedIntegerLit(l.line, l.col)
		}
		l.currentToken = Token{
			Ty:  TOKEN_TINTEGER_LIT,
			Val: uint64(val),
		}
	} else if binary {
		val, err := strconv.ParseUint(string(buf), 2, 64)
		if err != nil {
			return errors.MalformedIntegerLit(l.line, l.col)
		}
		l.currentToken = Token{
			Ty:  TOKEN_TINTEGER_LIT,
			Val: uint64(val),
		}
	} else {
		val, err := strconv.ParseFloat(string(buf), 64)
		if err != nil {
			return errors.MalformedFloatLit(l.line, l.col)
		}
		l.currentToken = Token{
			Ty:  TOKEN_TFLOAT_LIT,
			Val: uint64(math.Float64bits(val)),
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
