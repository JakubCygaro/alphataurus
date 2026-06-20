package assembler

import (
	// "math"
	"bufio"
	"math/rand"
	"strings"
	"testing"
)

// start with a known value e.g 3030
// then do operations on it and recursively build an expression tree
//     add    mul
// 3000-- 1000
//     \- 2000-- 2
//            \- 1000
// after this tree is built it can be used to emit an expression string
// and then the const evaluated expression can be checked for corectness
// also the expression tree can be evaluated

var exprTypes = []int{
	ARTHEXPR_TADD,
	ARTHEXPR_TSUB,
	ARTHEXPR_TMUL,
	ARTHEXPR_TDIV,
}

const (
	// chance to split an expression while recursively expanding the result
	SPLIT_EXPR_REC_CHANCE = 50
)

func abs(in int64) int64 {
	if in < 0 {
		return -in
	}
	return in
}

func buildExprRecurse(result int64, e *Expr) {
	if rand.Intn(101) > SPLIT_EXPR_REC_CHANCE || result == 0 {
		return
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
		a = result * b
	}
	e.Val = ArthExpr{
		Ty: exprToBuild,
		A:  MakeConstexprI64(a),
		B:  MakeConstexprI64(b),
	}
	buildExprRecurse(a, e.Val.(ArthExpr).A)
	buildExprRecurse(b, e.Val.(ArthExpr).B)

}
func TestExpressions(t *testing.T) {
	start := rand.Int63n(50)
	expr := MakeConstexprI64(start)
	buildExprRecurse(start, expr)
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
		t.Errorf("Wrong evaluation result wanted %v , got %v:", start, evalV)
		em, _ := reparsed.Emit()
		t.Error(em)
	}

}
