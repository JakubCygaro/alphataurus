package assembler

import (
	"bufio"
	"fmt"
	// "math"
	"math/rand"
	"strings"
	"testing"

	"github.com/JakubCygaro/alphataurus/internal/pkg/tests_commons"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

type pair0 struct {
	One int
	Two any
}

func asPair(o int, t any) pair0 {
	return pair0{One: o, Two: t}
}

var (
	identChars    string = "_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numChars      string = "0123456789"
	allIdentChars string = strings.Join([]string{identChars, numChars}, "")
	tokenList            = []pair0{
		asPair(TOKEN_TIDENT, func() string {
			first := identChars[rand.Intn(len(identChars)-1)]
			var sB strings.Builder
			for range rand.Intn(50) {
				sB.WriteByte(byte(allIdentChars[rand.Intn(len(identChars)-1)]))
			}
			return strings.
				Join([]string{string(first), sB.String()}, "")
		}),
		asPair(TOKEN_TCOMMA, ","),
		asPair(TOKEN_TREG, func() string {
			reg := tests_commons.RandomRegisterWord()
			return tests_commons.RegStr(reg, byte(rand.Int()%vm.SZ_64+1))
		}),
		asPair(TOKEN_TINTEGER_LIT, func() string {
			return fmt.Sprintf("%d", rand.Int())
		}),
		asPair(TOKEN_TFLOAT_LIT, func() string {
			return fmt.Sprintf("%f", rand.Float64())
		}),
		asPair(TOKEN_TUNSIGNED, "UNSIGNED"),
		asPair(TOKEN_TSIGNED, "SIGNED"),
		asPair(TOKEN_TFLOAT, "FLOAT"),
		asPair(TOKEN_TMINUS, "-"),
		asPair(TOKEN_TPLUS, "+"),
		asPair(TOKEN_TASTERISK, "*"),
		asPair(TOKEN_TSLASH, "/"),
		asPair(TOKEN_TDOT, "."),
		asPair(TOKEN_TOPENBRACKET, "["),
		asPair(TOKEN_TCLOSEDBRACKET, "]"),
		asPair(TOKEN_TOPENPAREN, "("),
		asPair(TOKEN_TCLOSEDPAREN, ")"),
		asPair(TOKEN_TCOLON, ":"),
		asPair(TOKEN_TSEMICOLON, ";"),
		asPair(TOKEN_TDOUBLESEMICOLON, ";;"),
		// asPair(TOKEN_TSINGLEQ, "'"),
		asPair(TOKEN_TAT, "@"),
		asPair(TOKEN_TWEAK, "WEAK"),
		asPair(TOKEN_TBYTE, "BYTE"),
		asPair(TOKEN_THALF, "HALF"),
		asPair(TOKEN_TWORD, "WORD"),
		asPair(TOKEN_TQUARTER, "QUARTER"),
		asPair(TOKEN_TABSOLUTE, "ABSOLUTE"),
	}
)

func TestTokensLineAndCol(t *testing.T) {
	sB := strings.Builder{}
	tokenCount := rand.Intn(100)
	sB.Grow(tokenCount)
	line, column := 1, 1
	nlChance := 0
	type tokInfo struct {
		ty   int
		s    string
		c, l int
	}
	inputTokens := make([]tokInfo, 0)
	for range tokenCount {
		if rand.Intn(100) < nlChance {
			sB.WriteByte('\n')
			line++
			column = 1
			nlChance = 0
			continue
		}
		whiteSpace := strings.Repeat(" ", 1+rand.Intn(9))
		column += len(whiteSpace)-1
		sB.WriteString(whiteSpace)

		nextTok := tokenList[rand.Int()%len(tokenList)]
		var stringRep string
		if s, ok := nextTok.Two.(string); ok {
			stringRep = s
		} else if f, ok := nextTok.Two.(func() string); ok {
			stringRep = f()
		} else {
			t.Errorf("Unsupported token generator value type (%+v)", nextTok.Two)
			return
		}
		sB.WriteString(stringRep)

		column++
		inputTokens = append(inputTokens, tokInfo{
			l:  line,
			c:  column,
			s:  stringRep,
			ty: nextTok.One,
		})

		column += len(stringRep)

		nlChance += 5
	}
	reader := bufio.NewReader(strings.NewReader(sB.String()))
	l := NewLexer(reader)
	var tok Token
	var e error
	outputTokens := make([]Token, 0)
	tok, e = l.ReadNextTokenReturn()
	for ; tok.Ty != TOKEN_TEOF && e == nil; tok, e = l.ReadNextTokenReturn() {
		if tok.Ty != TOKEN_TNEWLINE {
			outputTokens = append(outputTokens, tok)
		}
	}
	if e != nil {
		t.Errorf("Tokenizer error")
		t.Error(e)
		t.Error(sB.String())
		return
	}
	if len(inputTokens) != len(outputTokens) {
		t.Errorf("Input tokens and output tokens count mismatch")
		t.Errorf("IN = %v, OUT = %v", len(inputTokens), len(outputTokens))
		for i := range tests_commons.Max(len(inputTokens), len(outputTokens)) {
			t.Errorf("INPUT\t\t\t\tOUTPUT\n")
			if i < len(inputTokens) {
				t.Errorf("%+v\t\t", inputTokens[i])
			}
			if i < len(outputTokens) {
				t.Errorf("%+v", outputTokens[i])
			}
			t.Errorf("\n")
		}
		// t.Errorf("INPUT: %+v", inputTokens)
		// t.Errorf("OUTPUT: %+v", outputTokens)
		t.Error(sB.String())
		return
	}
	for i, oT := range outputTokens {
		iT := inputTokens[i]
		if iT.ty != oT.Ty || iT.l != int(oT.Line) || iT.c != int(oT.Col) {
			t.Errorf("Token mismatch")
			t.Errorf("\nin: %+v\nout: %+v", iT, oT)
			t.Errorf("\n|(1:1) ~ under this pipe\n%s", sB.String())
			return
		}
	}
}
