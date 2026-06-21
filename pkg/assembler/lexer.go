package assembler

import (
	"bufio"
	"fmt"
	// "math/big"

	// "io"
	"math"
	"strconv"
	"unicode"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
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
	TOKEN_TSEMICOLON
	TOKEN_TDOUBLESEMICOLON
	TOKEN_TNEWLINE
	TOKEN_TSINGLEQ
	TOKEN_TAT
	TOKEN_TWEAK
	TOKEN_TBYTE
	TOKEN_THALF
	TOKEN_TWORD
	TOKEN_TQUARTER
	TOKEN_TABSOLUTE
	TOKEN_TEOF
)

var keywords = map[string]int{
	"SIGNED":   TOKEN_TSIGNED,
	"UNSIGNED": TOKEN_TUNSIGNED,
	"FLOAT":    TOKEN_TFLOAT,
	"WEAK":     TOKEN_TWEAK,
	"BYTE":     TOKEN_TBYTE,
	"HALF":     TOKEN_THALF,
	"QUARTER":  TOKEN_TQUARTER,
	"WORD":     TOKEN_TWORD,
	"ABSOLUTE": TOKEN_TABSOLUTE,
}
var keywordNames map[int]string = makeKeywordNames()

func makeKeywordNames() map[int]string {
	names := make(map[int]string)
	for k, v := range keywords {
		names[v] = k
	}
	return names
}

func GetKeywordName(kwdToken int) (string, bool) {
	n, ok := keywordNames[kwdToken]
	return n, ok
}
func GetSizeKeyword(size byte) (string, bool) {
	switch size {
	case vm.SZ_8:
		return GetKeywordName(TOKEN_TBYTE)
	case vm.SZ_16:
		return GetKeywordName(TOKEN_TQUARTER)
	case vm.SZ_32:
		return GetKeywordName(TOKEN_THALF)
	case vm.SZ_64:
		return GetKeywordName(TOKEN_TWORD)
	default:
		return "", false
	}
}
func GetInvalidRegister() RegisterData {
	return RegisterData{
		Reg:  math.MaxInt64,
		Size: math.MaxInt8,
	}
}

type RegisterData struct {
	Reg  int
	Size byte
}

func (r RegisterData) IsInvalidRegister() bool {
	return r.Reg == math.MaxInt64 || r.Size == math.MaxInt8
}
func (r RegisterData) String() string {
	return RegToString(byte(r.Reg), r.Size)
}
func RegToString(reg, sz byte) string {
	switch reg {
	case vm.BP_IDX:
		return "bp"
	case vm.IP_IDX:
		return "ip"
	case vm.SP_IDX:
		return "sp"
	default:
		var suf string = ""
		switch sz {
		case vm.SZ_8:
			suf = "b"
		case vm.SZ_16:
			suf = "q"
		case vm.SZ_32:
			suf = "h"
		}
		return fmt.Sprintf("r%v%s", reg, suf)
	}
}

type Token struct {
	Ty   int
	Val  any
	Col  int
	Line int
}

// this is a helper funcion, it will attempt to present Token.Val as a string
//
// so for example if Val is of rune type it will be printed as a char
func (t Token) ForceValAsString() string {
	if s, ok := t.Val.(string); ok {
		if t.Ty == TOKEN_TSINGLEQ {
			return fmt.Sprintf("%q", s)
		}
		return s
	} else if r, ok := t.Val.(rune); ok {
		return string(r)
	} else if f, ok := t.Val.(uint64); ok && t.Ty == TOKEN_TFLOAT_LIT {
		return fmt.Sprintf("%v", math.Float64frombits(f))
	} else if reg, ok := t.Val.(RegisterData); ok {
		return reg.String()
	}else {
		return fmt.Sprint(t.Val)
	}

}

type Lexer struct {
	line int
	col  int
	// starting line and column for the current token
	sline, scol  int
	lastLine     int
	lastCol      int
	currentToken Token
	unRead       bool
	reader       *bufio.Reader
}

func nilToken() Token {
	return Token{
		Ty: TOKEN_TNIL,
	}
}

func TokenAsIdent(t *Token) (string, bool) {
	if s, ok := t.Val.(string); ok {
		return s, ok
	} else if r, ok := t.Val.(rune); ok {
		return string(r), ok
	} else {
		return "", false
	}
}
func TokenAsSize(t *Token) (byte, bool) {
	switch t.Ty {
	case TOKEN_TBYTE:
		return vm.SZ_8, true
	case TOKEN_TQUARTER:
		return vm.SZ_16, true
	case TOKEN_THALF:
		return vm.SZ_32, true
	case TOKEN_TWORD:
		return vm.SZ_64, true
	default:
		return 0xff, false
	}
}

func NewLexer(reader *bufio.Reader) Lexer {
	return Lexer{
		currentToken: nilToken(),
		reader:       reader,
		unRead:       false,
		col:          0,
		line:         1,
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
func IdentCheck(b byte) bool {
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
func (l *Lexer) readByte() (byte, bool) {
	b, err := l.reader.ReadByte()
	l.lastLine, l.lastCol = l.line, l.col
	if b == '\n' {
		l.line++
		l.col = 0
	} else {
		l.col++
	}
	return b, err == nil
}
func (l *Lexer) unreadByte() error {
	l.line, l.col = l.lastLine, l.lastCol
	return l.reader.UnreadByte()
}
func (l *Lexer) UnreadToken() {
	l.unRead = true
}

// Reads tokens into an array untill an EOF is encountered, EOF is included at the end
// of the array
func (l *Lexer) ReadTokensTillEof() ([]Token, error) {
	tokens := make([]Token, 0, 8)
	var eof bool
	var err error
	for {
		tokens, eof, err = l.ReadTokensInto(tokens)
		if eof || err != nil {
			return tokens, err
		}
		tmp := make([]Token, 0, cap(tokens)*2)
		tmp = append(tmp, tokens...)
		tokens = tmp
	}
}

// Reads tokens into a slice, appends the tokens to the end of the slice and up to the
// capacity of the slice.
//
// Returns the new slice, a boolean that indicates whether EOF has been reached and an
// optional error
func (l *Lexer) ReadTokensInto(out []Token) (ret []Token, eof bool, err error) {
	var tok Token
	for {
		if len(out) == cap(out) {
			return out, false, err
		}
		tok, err = l.ReadNextTokenReturn()
		if err != nil {
			break
		}
		out = append(out, tok)
		if tok.Ty == TOKEN_TEOF {
			break
		}
	}
	return out, tok.Ty == TOKEN_TEOF, err
}

// Reads the next token and returns it along with a possible error
func (l *Lexer) ReadNextTokenReturn() (Token, error) {
	if err := l.ReadNextToken(); err != nil {
		return Token{}, err
	} else {
		return l.currentToken, nil
	}
}

// Reads the next token and returns a possible error
//
// The read token can be accessed with CurrentToken()
func (l *Lexer) ReadNextToken() error {
	if l.unRead {
		l.unRead = false
		return nil
	}
	var b byte
	for {
		var ok bool
		b, ok = l.readByte()
		if !ok {
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
	l.scol, l.sline = l.col, l.line
	switch {
	case b == '\'':
		if err := l.readSingleQuoted(); err != nil {
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
	case b == ';':
		if b, ok := l.readByte(); ok && b == ';' {
			l.currentToken = Token{
				Ty:  TOKEN_TDOUBLESEMICOLON,
				Val: rune(b),
			}
		} else {
			if ok {
				l.unreadByte()
			}
			l.currentToken = Token{
				Ty:  TOKEN_TSEMICOLON,
				Val: rune(b),
			}
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
		next, ok := l.readByte()
		if ok && numberCheck(next) {
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
	case IdentCheck(b):
		buf := make([]byte, 0, 16)
		buf = append(buf, b)
		for {
			next, ok := l.readByte()
			if !ok {
				break
			}
			if IdentCheck(next) || numberCheck(next) {
				buf = append(buf, next)
			} else {
				l.unreadByte()
				break
			}
		}
		val := string(buf)
		if reg, sz, ok := RecognizeRegister(val); ok {
			l.currentToken = Token{
				Ty: TOKEN_TREG,
				Val: RegisterData{
					Reg:  reg,
					Size: sz,
				},
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
	l.currentToken.Line, l.currentToken.Col = l.sline, l.scol
	return nil
}
func (l *Lexer) readSingleQuoted() error {
	buf := make([]byte, 0, 64)
	for {
		next, ok := l.readByte()
		if !ok {
			return errors.LUnclosedSingleQuote(l.line, l.col)
		}
		if next == '\'' {
			break
		} else if next == '\\' {
			if next, ok := l.readByte(); !ok {
				return errors.LPrematureEndOfInput(l.line, l.col)
			} else if next == '\'' || next == '\\' {
				buf = append(buf, next)
			} else {
				return errors.UnsupportedEscape(l.line, l.col, rune(next))
			}
		} else if next == '\n' {
			return errors.LSingleQuoteNewline(l.line, l.col)
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
	const MAX_LIT_LEN int = 64
	const MAX_EXP int = 4
	expBufC := 0
	buf := [MAX_LIT_LEN]byte{0}
	bufC := 0
	binary := false
	dot := b == '.'
	appendToBuf := func(b byte) error {
		if (bufC >= MAX_LIT_LEN) ||
			(!binary && !dot && bufC > 20) {
			return errors.DigitLiteralTooLong(l.line, l.col, string(buf[:bufC]))
		}
		buf[bufC] = b
		bufC++
		return nil
	}
	if err := appendToBuf(b); err != nil {
		return err
	}
	e := false
	startedWithZero := b == '0'
	hex := false
	if startedWithZero {
		next, ok := l.readByte()
		switch {
		case next == 'x' && ok:
			hex = true
			dot = true
		case next == 'b' && ok:
			binary = true
			dot = true
		default:
			l.unreadByte()
		}
	}
	for {
		next, ok := l.readByte()
		if !ok {
			break
		}
		if !hex && !binary && numberCheck(next) {
			if err := appendToBuf(next); err != nil {
				return err
			}
		} else if hex && hexNumberCheck(next) {
			if err := appendToBuf(next); err != nil {
				return err
			}
		} else if binary && binaryNumberCheck(next) {
			if err := appendToBuf(next); err != nil {
				return err
			}
		} else if next == '.' && !dot {
			if err := appendToBuf(next); err != nil {
				return err
			}
			dot = true
		} else if (next == 'e' || next == 'E') && !e && !hex && !binary {
			if err := appendToBuf(next); err != nil {
				return err
			}
			next, ok := l.readByte()
			if !ok {
				return errors.LPrematureEndOfInput(l.line, l.col)
			}
			if next == '-' || next == '+' {
				if err := appendToBuf(next); err != nil {
					return err
				}
				next, ok = l.readByte()
				if !ok {
					return errors.LPrematureEndOfInput(l.line, l.col)
				}
			}
			for {
				if numberCheck(next) {
					if expBufC >= MAX_EXP {
						return errors.
							MalformedFloatLit(l.line, l.col, string(buf[:bufC]))
					} else if err := appendToBuf(next); err != nil {
						return err
					}
					next, _ = l.readByte()
				} else {
					l.unreadByte()
					break
				}
			}
		} else if unicode.IsSpace(rune(next)) || !IdentCheck(next) {
			l.unreadByte()
			break
		} else {
			return errors.MalformedIntegerLit(l.line, l.col, string(buf[:bufC]))
		}
	}
	lit := string(buf[:bufC])
	if !dot {
		val, err := strconv.ParseUint(lit, 10, 64)
		if err != nil {
			return errors.MalformedIntegerLit(l.line, l.col, lit)
		}
		l.currentToken = Token{
			Ty:  TOKEN_TINTEGER_LIT,
			Val: uint64(val),
		}
	} else if hex {
		val, err := strconv.ParseUint(lit, 16, 64)
		if err != nil {
			return errors.MalformedIntegerLit(l.line, l.col, lit)
		}
		l.currentToken = Token{
			Ty:  TOKEN_TINTEGER_LIT,
			Val: uint64(val),
		}
	} else if binary {
		val, err := strconv.ParseUint(lit, 2, 64)
		if err != nil {
			return errors.MalformedIntegerLit(l.line, l.col, lit)
		}
		l.currentToken = Token{
			Ty:  TOKEN_TINTEGER_LIT,
			Val: uint64(val),
		}
	} else {
		val, err := strconv.ParseFloat(lit, 64)
		if err != nil {
			return errors.MalformedFloatLit(l.line, l.col, string(buf[:bufC]))
		}
		// if e {
		// 	if i, err := strconv.ParseInt(string(expBuf[:expBufC]), 10, 64); err != nil {
		// 		return err
		// 	} else {
		// 		exp := math.Pow10(int(i))
		// 		val = val * exp
		// 	}
		// }
		l.currentToken = Token{
			Ty:  TOKEN_TFLOAT_LIT,
			Val: uint64(math.Float64bits(val)),
		}
	}
	return nil
}

func RecognizeRegister(s string) (int, byte, bool) {
	var sz byte = vm.SZ_64
	if len(s) < 2 || len(s) > 3 {
		return -1, sz, false
	}
	rx := s[1] - '0'
	// GP case
	if s[0] == 'r' && numberCheck(s[1]) && rx <= vm.GP_REG_MAX {
		if len(s) == 3 {
			switch s[2] {
			case 'b':
				sz = vm.SZ_8
			case 'q':
				sz = vm.SZ_16
			case 'h':
				sz = vm.SZ_32
			}
		}
		return int(rx), sz, true
	}
	switch s {
	case "sp":
		return vm.SP_IDX, sz, true
	case "bp":
		return vm.BP_IDX, sz, true
	case "ip":
		return vm.IP_IDX, sz, true
	}
	return -1, sz, false
}
