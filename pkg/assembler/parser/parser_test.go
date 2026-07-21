package assembler

import (
	"bufio"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"testing"
)

type instErrP struct {
	inst, err string
	col       int
}

func asInstErrP(inst, err string, col int) instErrP {
	return instErrP{inst: inst, err: err, col: col}
}

var instWithError = []instErrP{
	asInstErrP("mov r0, 100 these are extra tokens",
		".*Extra tokens on the line `these`", 13),
	asInstErrP("zupaeaea",
		".*Unknown identifier `zupaeaea`", 1),
	asInstErrP("section '.nothing'",
		".*Unknown section name `.nothing`", 9),
	asInstErrP("section 123123",
		".*Bad section type argument `123123`", 9),
	asInstErrP("import 123123",
		".*Expected a single quoted string parameter, got `123123`", 8),
	asInstErrP("import 'asdasd asdasd'",
		".*`asdasd asdasd` is not a valid identifier", 8),
	asInstErrP("add 105, r0",
		".*Disallowed destination `105`", 5),
	asInstErrP("add 'text'",
		".*Disallowed token in expression `'text'`", 5),
	asInstErrP("add ip, 100",
		".*Disallowed destination register `ip`", 5),
	asInstErrP("add r0 @ 100",
		".*Instruction missing a comma, got `@` instead", 8),
	asInstErrP("add r0, noncomptime",
		".*Disallowed source `noncomptime`", 9),
	asInstErrP("add r0, ip",
		".*Disallowed source register `ip`", 9),
	asInstErrP("add noncomptime, r0",
		".*Disallowed destination `noncomptime`.", 5),
	asInstErrP("call [ip+r0b]",
		".*Bad register size `r0b`, expected a WORD sized register.", 6),

}

func TestParsingErrorsF(t *testing.T) {
	for _, e := range instWithError {
		linesCount := rand.Intn(20) + 5
		errorLine := rand.Intn(linesCount)
		lines := make([]string, 0)
		lines = append(lines,
			"section '.code'",
			"@entry",
		)
		for i := range linesCount {
			if i == errorLine {
				lines = append(lines, e.inst)
			} else {
				lines = append(lines, "nop")
			}
		}
		errorLine += 3
		asm := strings.Join(lines, "\n")
		p := NewParser(bufio.NewReader(strings.NewReader(asm)))
		p.ForceCoalesceGenerics = true
		var err error
		for {
			var ok bool
			ok, err = p.ParseNext()
			if !ok || err != nil {
				break
			}
		}
		if err == nil {
			t.Errorf("Expected parsing error, got no error")
			t.Errorf("Expected: `%s`", e.err)
			t.Errorf("Assembly:\n%s", asm)
		} else if matched, rerr := regexp.MatchString(e.err, err.Error()); !matched {
			t.Errorf("Parsing error does not match expected error")
			t.Errorf("Expected: `%s`", e.err)
			t.Errorf("Got: `%s`", err.Error())
			t.Errorf("Assembly:\n%s", asm)
		} else if rerr != nil {
			t.Errorf("Regexp error:")
			t.Error(rerr.Error())
		} else if m, rerr := regexp.
			MatchString(fmt.Sprintf("(%v:%v)", errorLine, e.col), err.Error()); !m {

			t.Errorf("Parsing error line or column does not match")
			t.Errorf("Expected (%v:%v)", errorLine, e.col)
			t.Errorf("Got: `%s`", err.Error())
			t.Errorf("Assembly:\n%s", asm)
		} else if rerr != nil {
			t.Errorf("Regexp error:")
			t.Error(rerr.Error())
		}
	}
}
