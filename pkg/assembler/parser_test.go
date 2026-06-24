package assembler

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strings"
	"testing"
)

var exprTypes = []int{
	ARTHEXPR_TADD,
	ARTHEXPR_TSUB,
	ARTHEXPR_TMUL,
	ARTHEXPR_TDIV,
}

const (
	// chance to split an expression while recursively expanding the result
	SPLIT_EXPR_REC_CHANCE              = 50
	SPLIT_MAX_DEPTH                    = 6
	EXPR_PARSE_REPARSE_EVAL_TEST_COUNT = 10
)

func abs(in int64) int64 {
	if in < 0 {
		return -in
	}
	return in
}

func buildExprRecurseInt(result int64, depth uint, e *Expr) uint {
	if rand.Intn(101) > SPLIT_EXPR_REC_CHANCE+int(depth)*5 ||
		result == 0 ||
		depth >= SPLIT_MAX_DEPTH {
		return depth
	}
	e.Ty = EXPR_TARTH
	var a, b int64
	exprToBuild := exprTypes[rand.Int()%len(exprTypes)]
start:
	switch exprToBuild {
	case ARTHEXPR_TADD:
		a = rand.Int63n(abs(result))
		b = result - a
	case ARTHEXPR_TSUB:
		b = rand.Int63n(abs(result))
		a = result + b
	case ARTHEXPR_TMUL:
		b = rand.Int63n(abs(result))
		if b == 0 {
			b++
		}
		if rem := result % b; rem != 0 {
			b -= rem
			retries := 100
			for result%b != 0 && retries >= 0 {
				b++
				retries--
			}
		}
		if result%b != 0 {
			//default to add
			exprToBuild = ARTHEXPR_TADD
			goto start
		}
		a = result / b
	case ARTHEXPR_TDIV:
		b = rand.Int63n(abs(result))
		if b == 0 {
			b++
		}
		retries := 100
		for retries >= 0 {
			a = result * b
			if a%b != 0 {
				b++
				retries--
			} else {
				break
			}
		}
		if retries < 0 {
			//default to add
			exprToBuild = ARTHEXPR_TSUB
			goto start
		}
	}
	e.Val = ArthExpr{
		Ty: exprToBuild,
		A:  MakeConstexprI64(a),
		B:  MakeConstexprI64(b),
	}
	return max(
		buildExprRecurseInt(a, depth+1, e.Val.(ArthExpr).A),
		buildExprRecurseInt(b, depth+1, e.Val.(ArthExpr).B),
	)
}
func buildExprRecurseFloat(result float64, depth uint, e *Expr) uint {
	if rand.Intn(101) > SPLIT_EXPR_REC_CHANCE+int(depth)*5 ||
		result == 0 ||
		depth >= SPLIT_MAX_DEPTH {
		return depth
	}
	e.Ty = EXPR_TARTH
	var a, b float64
	exprToBuild := exprTypes[rand.Int()%len(exprTypes)]
	switch exprToBuild {
	case ARTHEXPR_TADD:
		a = rand.Float64() * 1000
		b = result - a
	case ARTHEXPR_TSUB:
		b = rand.Float64() * 1000
		a = result + b
	case ARTHEXPR_TMUL:
		b = rand.Float64() * 1000
		if b == 0 {
			b++
		}
		a = result / b
	case ARTHEXPR_TDIV:
		b = rand.Float64() * 1000
		if b == 0 {
			b++
		}
		a = result * b
	}
	e.Val = ArthExpr{
		Ty: exprToBuild,
		A:  MakeConstexprF64(a),
		B:  MakeConstexprF64(b),
	}
	return max(
		buildExprRecurseFloat(a, depth+1, e.Val.(ArthExpr).A),
		buildExprRecurseFloat(b, depth+1, e.Val.(ArthExpr).B),
	)

}
func TestExpressionsInt(t *testing.T) {
	for range EXPR_PARSE_REPARSE_EVAL_TEST_COUNT {
		start := rand.Int63n(10_000) - 5_000
		expr := MakeConstexprI64(start)
		buildExprRecurseInt(start, 0, expr)
		em, _ := expr.Emit()
		if eval, ok := TryConstEvaluateExpression(expr); !ok {
			t.Errorf("TryConstEvaluateExpression failed for expression:")
			t.Error(em)
		} else if evalV := int64(eval.Val); evalV != start {
			t.Errorf("Wrong evaluation result wanted %v , got %v:", start, evalV)
			t.Error(em)
		}
		p := NewParser(bufio.NewReader(strings.NewReader(em)))
		if reparsed, err := p.ParseExpression(); err != nil {
			t.Errorf("Failed to reparse emitted expression")
			t.Error(err)
			t.Error(em)
		} else if eval, ok := TryConstEvaluateExpression(expr); !ok {
			t.Errorf("TryConstEvaluateExpression failed for reparsed expression:")
			em, _ := reparsed.Emit()
			t.Error(em)
		} else if evalV := int64(eval.Val); evalV != start {
			t.Errorf("Wrong reparse evaluation result wanted %v , got %v:", start, evalV)
			em, _ := reparsed.Emit()
			t.Error(em)
		}
	}
}
func TestExpressionsFloat(t *testing.T) {
	for range EXPR_PARSE_REPARSE_EVAL_TEST_COUNT {
		start := (rand.Float64() * 10_000) - 5_000
		expr := MakeConstexprF64(start)
		buildExprRecurseFloat(start, 0, expr)
		em, _ := expr.Emit()
		if eval, ok := TryConstEvaluateExpression(expr); !ok {
			t.Errorf("TryConstEvaluateExpression failed for expression:")
			t.Error(em)
		} else if evalV := math.Float64frombits(eval.Val); math.Abs(evalV-start) > 0.1 {
			t.Errorf("Wrong evaluation result wanted %v , got %v:", start, evalV)
			t.Errorf("Diff: %v", math.Abs(evalV-start))
			t.Error(em)
		}
		p := NewParser(bufio.NewReader(strings.NewReader(em)))
		if reparsed, err := p.ParseExpression(); err != nil {
			t.Errorf("Failed to reparse emitted expression")
			t.Error(err)
			t.Error(em)
		} else if eval, ok := TryConstEvaluateExpression(expr); !ok {
			t.Errorf("TryConstEvaluateExpression failed for reparsed expression:")
			em, _ := reparsed.Emit()
			t.Error(em)
		} else if evalV := math.Float64frombits(eval.Val); math.Abs(evalV-start) > 0.1 {
			t.Errorf("Wrong reparse evaluation result wanted %v , got %v:", start, evalV)
			t.Errorf("Diff: %v", math.Abs(evalV-start))
			em, _ := reparsed.Emit()
			t.Error(em)
		}
	}
}

type instErrP struct {
	inst, err string
	col       int
}

func asInstErrP(inst, err string, col int) instErrP {
	return instErrP{inst: inst, err: err, col: col}
}

var instWithError = []instErrP{
	asInstErrP("mov r0, 100 these are extra tokens",
		".*Extra tokens the on line `these`", 13),
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
	asInstErrP("add 105",
		".*First operand to instruction must be a valid register, got `105`", 5),
	asInstErrP("add 'text'",
		".*Disallowed token in expression `'text'`", 5),
	asInstErrP("add ip, 100",
		".*Disallowed destination register `ip`", 5),
	asInstErrP("add r0 @ 100",
		".*Instruction missing a comma, got `@` instead", 8),
	// asInstErrP("add r0, noncomptime",
	// 	".*Second operand to instruction has to be a valid register "+
	// 	"or a compile time expression, got `noncomptime`", 9),
	// asInstErrP("add r0, ip",
	// 	".*Disallowed source register `ip`", 9),
	// asInstErrP("add r0, noncomptime",
	// 	"Second operand to instruction has to be a valid register "+
	// 		"or a compile time expression got `noncomptime`", 9),

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
