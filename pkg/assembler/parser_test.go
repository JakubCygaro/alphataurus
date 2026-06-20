package assembler

import (
	"bufio"
	"math"
	"math/rand"
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
	SPLIT_EXPR_REC_CHANCE = 50
	SPLIT_MAX_DEPTH       = 6
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
	start := rand.Int63n(6000) - 3000
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
func TestExpressionsFloat(t *testing.T) {
	start := rand.Float64() * 10_000
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
