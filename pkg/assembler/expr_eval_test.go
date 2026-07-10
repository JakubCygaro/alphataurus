package assembler

import (
	"bufio"
	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
	"math"
	"math/rand"
	"strings"
	"testing"
)

var exprTypes = []int{
	pr.ARTHEXPR_TADD,
	pr.ARTHEXPR_TSUB,
	pr.ARTHEXPR_TMUL,
	pr.ARTHEXPR_TDIV,
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

func buildExprRecurseInt(result int64, depth uint, e *pr.Expr) uint {
	if rand.Intn(101) > SPLIT_EXPR_REC_CHANCE+int(depth)*5 ||
		result == 0 ||
		depth >= SPLIT_MAX_DEPTH {
		return depth
	}
	var a, b int64
	exprToBuild := exprTypes[rand.Int()%len(exprTypes)]
start:
	switch exprToBuild {
	case pr.ARTHEXPR_TADD:
		a = rand.Int63n(abs(result))
		b = result - a
	case pr.ARTHEXPR_TSUB:
		b = rand.Int63n(abs(result))
		a = result + b
	case pr.ARTHEXPR_TMUL:
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
			exprToBuild = pr.ARTHEXPR_TADD
			goto start
		}
		a = result / b
	case pr.ARTHEXPR_TDIV:
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
			exprToBuild = pr.ARTHEXPR_TSUB
			goto start
		}
	}
	A := pr.MakeConstexprI64(a)
	B := pr.MakeConstexprI64(b)
	switch exprToBuild {
	case pr.ARTHEXPR_TADD:
		e = pr.MakeArth(
			pr.ArthExprAdd{
				A: A,
				B: B,
			},
		)
	case pr.ARTHEXPR_TSUB:
		e = pr.MakeArth(
			pr.ArthExprSub{
				A: A,
				B: B,
			},
		)
	case pr.ARTHEXPR_TMUL:
		e = pr.MakeArth(
			pr.ArthExprMul{
				A: A,
				B: B,
			},
		)
	case pr.ARTHEXPR_TDIV:
		e = pr.MakeArth(
			pr.ArthExprMul{
				A: A,
				B: B,
			},
		)
	}
	return max(
		buildExprRecurseInt(a, depth+1, A),
		buildExprRecurseInt(b, depth+1, B),
	)
}
func buildExprRecurseFloat(result float64, depth uint, e *pr.Expr) uint {
	if rand.Intn(101) > SPLIT_EXPR_REC_CHANCE+int(depth)*5 ||
		result == 0 ||
		depth >= SPLIT_MAX_DEPTH {
		return depth
	}
	var a, b float64
	exprToBuild := exprTypes[rand.Int()%len(exprTypes)]
	switch exprToBuild {
	case pr.ARTHEXPR_TADD:
		a = rand.Float64() * 1000
		b = result - a
	case pr.ARTHEXPR_TSUB:
		b = rand.Float64() * 1000
		a = result + b
	case pr.ARTHEXPR_TMUL:
		b = rand.Float64() * 1000
		if b == 0 {
			b++
		}
		a = result / b
	case pr.ARTHEXPR_TDIV:
		b = rand.Float64() * 1000
		if b == 0 {
			b++
		}
		a = result * b
	}
	A := pr.MakeConstexprF64(a)
	B := pr.MakeConstexprF64(b)
	switch exprToBuild {
	case pr.ARTHEXPR_TADD:
		e = pr.MakeArth(
			pr.ArthExprAdd{
				A: A,
				B: B,
			},
		)
	case pr.ARTHEXPR_TSUB:
		e = pr.MakeArth(
			pr.ArthExprSub{
				A: A,
				B: B,
			},
		)
	case pr.ARTHEXPR_TMUL:
		e = pr.MakeArth(
			pr.ArthExprMul{
				A: A,
				B: B,
			},
		)
	case pr.ARTHEXPR_TDIV:
		e = pr.MakeArth(
			pr.ArthExprMul{
				A: A,
				B: B,
			},
		)
	}
	return max(
		buildExprRecurseFloat(a, depth+1, A),
		buildExprRecurseFloat(b, depth+1, B),
	)

}
func TestExpressionsInt(t *testing.T) {
	ev := ExpressionEvaluator{}
	for range EXPR_PARSE_REPARSE_EVAL_TEST_COUNT {
		start := rand.Int63n(10_000) - 5_000
		expr := pr.MakeConstexprI64(start)
		buildExprRecurseInt(start, 0, expr)
		em, _ := expr.Emit()
		if eval, ok, err := TryConstEvalExprT[pr.ConstExprILit](&ev, expr); err != nil {
			t.Errorf("TryConstEvaluateExpression failed for expression:")
			t.Error(em)
			t.Error(err)
		} else if !ok {
			t.Errorf("TryConstEvaluateExpression failed for expression:")
			t.Error(em)
		} else if evalV := eval.Signed(); evalV != start {
			t.Errorf("Wrong evaluation result wanted %v , got %v:", start, evalV)
			t.Error(em)
		}
		p := pr.NewParser(bufio.NewReader(strings.NewReader(em)))
		if reparsed, err := p.ParseExpression(); err != nil {
			t.Errorf("Failed to reparse emitted expression")
			t.Error(err)
			t.Error(em)
		} else if eval, ok, err :=
			TryConstEvalExprT[pr.ConstExprILit](&ev, expr); err != nil {
			t.Errorf("TryConstEvaluateExpression failed for expression:")
			t.Error(em)
			t.Error(err)
		} else if !ok {
			t.Errorf("TryConstEvaluateExpression failed for reparsed expression:")
			em, _ := reparsed.Emit()
			t.Error(em)
		} else if evalV := eval.Signed(); evalV != start {
			t.Errorf("Wrong reparse evaluation result wanted %v , got %v:", start, evalV)
			em, _ := reparsed.Emit()
			t.Error(em)
		}
	}
}
func TestExpressionsFloat(t *testing.T) {
	ev := ExpressionEvaluator{}
	for range EXPR_PARSE_REPARSE_EVAL_TEST_COUNT {
		start := (rand.Float64() * 10_000) - 5_000
		expr := pr.MakeConstexprF64(start)
		buildExprRecurseFloat(start, 0, expr)
		em, _ := expr.Emit()
		if eval, ok, err := TryConstEvalExprT[pr.ConstExprFLit](&ev, expr); err != nil {
			t.Errorf("TryConstEvaluateExpression failed for expression:")
			t.Error(em)
			t.Error(err)
		} else if !ok {
			t.Errorf("TryConstEvaluateExpression failed for expression:")
			t.Error(em)
		} else if evalV := eval.ToFloat(); math.Abs(evalV-start) > 0.1 {
			t.Errorf("Wrong evaluation result wanted %v , got %v:", start, evalV)
			t.Errorf("Diff: %v", math.Abs(evalV-start))
			t.Error(em)
		}
		p := pr.NewParser(bufio.NewReader(strings.NewReader(em)))
		if reparsed, err := p.ParseExpression(); err != nil {
			t.Errorf("Failed to reparse emitted expression")
			t.Error(err)
			t.Error(em)
		} else if eval, ok, err :=
			TryConstEvalExprT[pr.ConstExprFLit](&ev, expr); err != nil {
			t.Errorf("Failed to reparse emitted expression")
			t.Error(err)
			t.Error(em)
		} else if !ok {
			t.Errorf("TryConstEvaluateExpression failed for reparsed expression:")
			em, _ := reparsed.Emit()
			t.Error(em)
		} else if evalV := eval.ToFloat(); math.Abs(evalV-start) > 0.1 {
			t.Errorf("Wrong reparse evaluation result wanted %v , got %v:", start, evalV)
			t.Errorf("Diff: %v", math.Abs(evalV-start))
			em, _ := reparsed.Emit()
			t.Error(em)
		}
	}
}
