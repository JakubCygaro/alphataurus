package assembler

import (
	"fmt"
	"math"

	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
)

type VariableValueProvider interface {
	// Get variable value for given name, either uint64 or float64
	Get(name string) any
	// Clear internal cache used for storing failed variable access
	ClearFailedExtractCache()
}

type EvaluationContext struct {
	VarProvider VariableValueProvider
}

type ExpressionEvaluator struct {
	Ctx EvaluationContext
}

// Extract the value of the constant expression into either uint64 or float64
// if the value cannot be extracted it is returned as nil
func (ev *ExpressionEvaluator) extractValue(cexpr pr.ConstExpr) any {
	switch v := cexpr.Val.(type) {
	case pr.ExprIden:
		if ev.Ctx.VarProvider == nil {
			return nil
		}
		f := ev.Ctx.VarProvider.Get(v.Ident)
		return f
	case pr.ConstExprILit:
		return v.Integer
	case pr.ConstExprFLit:
		return v.ToFloat()
	}
	return nil
}

type binopFnU func(a, b uint64) (uint64, error)
type binopFnF func(a, b float64) (float64, error)

// perform binary operation between two expressions
func (ev *ExpressionEvaluator) performBinary(
	a, b *pr.Expr, opU binopFnU, opF binopFnF) (
	res *pr.Expr, err error) {
	if !a.IsConstexpr() || a.IsArthexpr() {
		if tmp, err := ev.tryEvaluateExpression(a); err != nil {
			return tmp, err
		} else if tmp == nil {
			return nil, nil
		} else {
			a = tmp
		}
	}
	if !b.IsConstexpr() || b.IsArthexpr() {
		if tmp, err := ev.tryEvaluateExpression(b); err != nil {
			return tmp, err
		} else if tmp == nil {
			return nil, nil
		} else {
			b = tmp
		}
	}
	aV, bV := ev.extractValue(a.Val.(pr.ConstExpr)),
		ev.extractValue(b.Val.(pr.ConstExpr))
	if aV == nil || bV == nil {
		return nil, fmt.Errorf("Could not extract value of either aV or bV")
	} else if res, err := driver(aV, bV, opU, opF); err != nil {
		return nil, err
	} else {
		return &pr.Expr{Val: res}, nil
	}
}

// this function deduces the types of a and b and then performs an op
func driver(a, b any, opU binopFnU, opF binopFnF) (pr.ConstExpr, error) {
	var def pr.ConstExpr
	switch aV := a.(type) {
	case uint64:
		switch bV := b.(type) {
		case uint64:
			if r, e := opU(aV, bV); e != nil {
				return def, e
			} else {
				return pr.ConstExpr{
					Val: pr.ConstExprILit{
						Integer: r,
					},
				}, nil
			}
		case float64:
			if r, e := opF(float64(aV), bV); e != nil {
				return def, e
			} else {
				return pr.ConstExpr{
					Val: pr.ConstExprFLit{
						Float: math.Float64bits(r),
					},
				}, nil
			}
		}
	case float64:
		switch bV := b.(type) {
		case uint64:
			if r, e := opF(float64(aV), float64(bV)); e != nil {
				return def, e
			} else {
				return pr.ConstExpr{
					Val: pr.ConstExprFLit{
						Float: math.Float64bits(r),
					},
				}, nil
			}
		case float64:
			if r, e := opF(float64(aV), float64(bV)); e != nil {
				return def, e
			} else {
				return pr.ConstExpr{
					Val: pr.ConstExprFLit{
						Float: math.Float64bits(r),
					},
				}, nil
			}
		}
	}
	return def, fmt.Errorf("TODO Operation not performed -")
}

func (ev *ExpressionEvaluator) performOperation(arth pr.ArthExpr) (*pr.Expr, error) {
	switch ar := arth.Val.(type) {
	case pr.ArthExprAdd:
		if res, err := ev.performBinary(
			ar.A,
			ar.B,
			func(a, b uint64) (uint64, error) { return a + b, nil },
			func(a, b float64) (float64, error) { return a + b, nil },
		); err != nil {
			return nil, err
		} else {
			return res, nil
		}
	case pr.ArthExprSub:
		if res, err := ev.performBinary(
			ar.A,
			ar.B,
			func(a, b uint64) (uint64, error) { return a - b, nil },
			func(a, b float64) (float64, error) { return a - b, nil },
		); err != nil {
			return nil, err
		} else {
			return res, nil
		}
	case pr.ArthExprMul:
		if res, err := ev.performBinary(
			ar.A,
			ar.B,
			func(a, b uint64) (uint64, error) { return a * b, nil },
			func(a, b float64) (float64, error) { return a * b, nil },
		); err != nil {
			return nil, err
		} else {
			return res, nil
		}
	case pr.ArthExprDiv:
		if res, err := ev.performBinary(
			ar.A,
			ar.B,
			func(a, b uint64) (uint64, error) {
				if b == 0 {
					return 0, fmt.Errorf("TODO: binary op error, zero division")
				}
				return uint64(int64(a) / int64(b)), nil
			},
			func(a, b float64) (float64, error) {
				if b == 0.0 {
					return 0, fmt.Errorf("TODO: binary op error, zero division")
				}
				return a / b, nil
			},
		); err != nil {
			return nil, err
		} else {
			return res, nil
		}
	}
	return nil, fmt.
		Errorf("TODO Unrecognized binary operation expression typ")
}

// Does what TryEvaluateExpression does, but at the end verifies that the expression is constant
func (ev *ExpressionEvaluator) TryConstEvaluateExpression(e *pr.Expr) (
	pr.ConstExpr, bool, error) {
	var def pr.ConstExpr
	if expr, err := ev.TryEvaluateExpression(e); err != nil {
		return def, false, err
	} else if expr == nil {
		return def, false, nil
	}else if expr.IsConstexpr() {
		return expr.Val.(pr.ConstExpr), true, nil
	}
	return def, false, nil
}
func TryConstEvalExprT[CE pr.CExpr](
	ev *ExpressionEvaluator, e *pr.Expr) (CE, bool, error) {
	var def CE
	if expr, err := ev.TryEvaluateExpression(e); err != nil {
		return def, false, err
	} else if expr.IsConstexpr() {
		ce, ok := expr.Val.(pr.ConstExpr).Val.(CE), true
		return ce, ok, nil
	}
	return def, false, nil
}
// Basically try to evalueate an expression at compile time.
// Attenpts to evaluate the whole expression and either returns it or returns a nil pointer
// error contains all potential errors encountered
//
// nil, nil -> unable to evaluate, no errors
// nil, error -> unable to evaluate, has errors
// *pr.Expr, nil -> evaluated, no errors
//
// This function is a wrapper tryEvaluateExpression internally
// that first calls VariableValueProvider.ClearFailedExtractCache() before evaluation logic.
//
// This is done so that previous failed evaluation data is not clogging up the cache.
// the cache is VariableValueProvider implementation specific and might as well be a no-op.
// Also the access to the cache is done through the VariableValueProvider and not the evaluator.
func (ev *ExpressionEvaluator) TryEvaluateExpression(e *pr.Expr) (*pr.Expr, error) {
	if ev.Ctx.VarProvider != nil {
		ev.Ctx.VarProvider.ClearFailedExtractCache()
	}
	return ev.tryEvaluateExpression(e)
}

func (ev *ExpressionEvaluator) tryEvaluateExpression(e *pr.Expr) (*pr.Expr, error) {
	switch expr := e.Val.(type) {
	case pr.NegExpr:
		if inner, err := ev.tryEvaluateExpression(expr.Inner); err != nil {
			return nil, err
		} else if inner == nil {
			return nil, nil
		} else if cexpr, ok := inner.Val.(pr.ConstExpr); !ok {
			return nil, fmt.Errorf("TODO: cannot negate non const expression")
		} else if val := ev.extractValue(cexpr); val == nil{
			return nil, nil
		} else {
			switch v := val.(type) {
			case uint64:
				return pr.MakeConstexprI64(-int64(v)), nil
			case float64:
				return pr.MakeConstexprF64(-v), nil
			default:
				panic("Unsupported variable value extracted from evaluation context")
			}
		}
	case pr.ConstExpr:
		if val := ev.extractValue(expr); val == nil {
			return nil, nil
		} else {
			switch v := val.(type) {
			case uint64:
				return pr.MakeConstexprU64(v), nil
			case float64:
				return pr.MakeConstexprF64(v), nil
			default:
				panic("Unsupported variable value extracted from evaluation context")
			}
		}
		//if this is an arthmetic expression, atttempt to evaluate it
	case pr.ArthExpr:
		return ev.performOperation(expr)
	case pr.RegExpr:
		return e, nil
	case pr.DerefExpr:
		if inner, err := ev.tryEvaluateExpression(expr.Inner); err != nil {
			return nil, err
		} else if inner == nil {
			return nil, nil
		} else {
			return pr.MakeDeref(inner), nil
		}
	case pr.OneRegOffsetExpr:
		if expr.Offset != nil {
			if offset, err := ev.tryEvaluateExpression(expr.Offset); err != nil {
				return nil, err
			} else if offset == nil {
				return nil, nil
			} else {
				expr.Offset = offset
				return &pr.Expr{Val: expr}, nil
			}
		} else {
			return &pr.Expr{Val: expr}, nil
		}
	case pr.TwoRegOffsetExpr:
		if expr.Offset != nil {
			if offset, err := ev.tryEvaluateExpression(expr.Offset); err != nil {
				return nil, err
			} else if offset == nil {
				return nil, nil
			} else {
				expr.Offset = offset
				return &pr.Expr{Val: expr}, nil
			}
		} else {
			return &pr.Expr{Val: expr}, nil
		}
	}
	return nil, fmt.Errorf("TODO: expression cannot be evaluated")
}
