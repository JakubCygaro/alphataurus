package assembler

import (
	"math"

	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
)

type VariableValueProvider interface {
	// Get variable value for given name, either uint64 or float64
	Get(name string) any
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
	case pr.ConstExprIden:
		if ev.Ctx.VarProvider == nil {
			return nil
		}
		f := ev.Ctx.VarProvider.Get(v.Ident)
		return f
	case pr.ConstExprILit:
		return v.Integer
	case pr.ConstExprFLit:
		return v.Float
	}
	return nil
}
func (ev *ExpressionEvaluator) performBinary(
	a, b *pr.Expr, op func(any, any) pr.ConstExpr) (res, bestA, bestB *pr.Expr, ok bool) {
	var canPerform bool = true
	if !a.IsConstexpr() || a.IsArthexpr() {
		tmp, ok := ev.TryEvaluateExpression(a)
		canPerform = canPerform && ok
		a = tmp
	}
	if !b.IsConstexpr() || b.IsArthexpr() {
		tmp, ok := ev.TryEvaluateExpression(b)
		canPerform = canPerform && ok
		b = tmp
	}
	if !canPerform {
		return nil, a, b, false
	}
	aV, bV := ev.extractValue(a.Val.(pr.ConstExpr)),
		ev.extractValue(b.Val.(pr.ConstExpr))
	if aV == nil || bV == nil {
		return nil, a, b, false
	} else {
		return &pr.Expr{Val: op(aV, bV)}, nil, nil, true
	}
}

func (ev *ExpressionEvaluator) performOperation(arth pr.ArthExpr) (*pr.Expr, bool) {
	switch ar := arth.Val.(type) {
	case pr.ArthExprAdd:
		res, a, b, ok := ev.performBinary(ar.A, ar.B, func(a, b any) pr.ConstExpr {
			switch aV := a.(type) {
			case uint64:
				switch bV := b.(type) {
				case uint64:
					return pr.ConstExpr{
						Val: pr.ConstExprILit{
							Integer: aV + bV,
						},
					}
				case float64:
					return pr.ConstExpr{
						Val: pr.ConstExprFLit{
							Float: math.Float64bits(float64(aV) + bV),
						},
					}
				}
			case float64:
				switch bV := b.(type) {
				case uint64:
					return pr.ConstExpr{
						Val: pr.ConstExprFLit{
							Float: math.Float64bits(aV + float64(bV)),
						},
					}
				case float64:
					return pr.ConstExpr{
						Val: pr.ConstExprFLit{
							Float: math.Float64bits(float64(aV) + bV),
						},
					}
				}
			}
			return pr.ConstExpr{}
		})
		if !ok {
			return pr.MakeArth(
				pr.ArthExprAdd{
					A: a,
					B: b,
				},
			), false
		} else {
			return res, true
		}
	case pr.ArthExprSub:
		res, a, b, ok := ev.performBinary(ar.A, ar.B, func(a, b any) pr.ConstExpr {
			switch aV := a.(type) {
			case uint64:
				switch bV := b.(type) {
				case uint64:
					return pr.ConstExpr{
						Val: pr.ConstExprILit{
							Integer: aV - bV,
						},
					}
				case float64:
					return pr.ConstExpr{
						Val: pr.ConstExprFLit{
							Float: math.Float64bits(float64(aV) - bV),
						},
					}
				}
			case float64:
				switch bV := b.(type) {
				case uint64:
					return pr.ConstExpr{
						Val: pr.ConstExprFLit{
							Float: math.Float64bits(aV - float64(bV)),
						},
					}
				case float64:
					return pr.ConstExpr{
						Val: pr.ConstExprFLit{
							Float: math.Float64bits(float64(aV) - bV),
						},
					}
				}
			}
			return pr.ConstExpr{}
		})
		if !ok {
			return pr.MakeArth(
				pr.ArthExprAdd{
					A: a,
					B: b,
				},
			), false
		} else {
			return res, true
		}
	case pr.ArthExprMul:
		res, a, b, ok := ev.performBinary(ar.A, ar.B, func(a, b any) pr.ConstExpr {
			switch aV := a.(type) {
			case uint64:
				switch bV := b.(type) {
				case uint64:
					return pr.ConstExpr{
						Val: pr.ConstExprILit{
							Integer: aV * bV,
						},
					}
				case float64:
					return pr.ConstExpr{
						Val: pr.ConstExprFLit{
							Float: math.Float64bits(float64(aV) * bV),
						},
					}
				}
			case float64:
				switch bV := b.(type) {
				case uint64:
					return pr.ConstExpr{
						Val: pr.ConstExprFLit{
							Float: math.Float64bits(aV * float64(bV)),
						},
					}
				case float64:
					return pr.ConstExpr{
						Val: pr.ConstExprFLit{
							Float: math.Float64bits(float64(aV) * bV),
						},
					}
				}
			}
			return pr.ConstExpr{}
		})
		if !ok {
			return pr.MakeArth(
				pr.ArthExprAdd{
					A: a,
					B: b,
				},
			), false
		} else {
			return res, true
		}
	case pr.ArthExprDiv:
		res, a, b, ok := ev.performBinary(ar.A, ar.B, func(a, b any) pr.ConstExpr {
			switch aV := a.(type) {
			case uint64:
				switch bV := b.(type) {
				case uint64:
					return pr.ConstExpr{
						Val: pr.ConstExprILit{
							Integer: aV / bV,
						},
					}
				case float64:
					return pr.ConstExpr{
						Val: pr.ConstExprFLit{
							Float: math.Float64bits(float64(aV) / bV),
						},
					}
				}
			case float64:
				switch bV := b.(type) {
				case uint64:
					return pr.ConstExpr{
						Val: pr.ConstExprFLit{
							Float: math.Float64bits(aV / float64(bV)),
						},
					}
				case float64:
					return pr.ConstExpr{
						Val: pr.ConstExprFLit{
							Float: math.Float64bits(float64(aV) / bV),
						},
					}
				}
			}
			return pr.ConstExpr{}
		})
		if !ok {
			return pr.MakeArth(
				pr.ArthExprAdd{
					A: a,
					B: b,
				},
			), false
		} else {
			return res, true
		}
	}
	return &pr.Expr{Val: arth}, false
}

// Does what TryEvaluateExpression does, but at the end verifies that the expression is constant
func (ev *ExpressionEvaluator) TryConstEvaluateExpression(e *pr.Expr) (pr.ConstExpr, bool) {
	if expr, ok := ev.TryEvaluateExpression(e); ok && expr.IsConstexpr() {
		return expr.Val.(pr.ConstExpr), true
	}
	return pr.ConstExpr{}, false
}
func TryConstEvalExprT[CE pr.CExpr](
	ev *ExpressionEvaluator, e *pr.Expr) (any, bool) {
	if expr, ok := ev.TryEvaluateExpression(e); ok && expr.IsConstexpr() {
		ce, ok := expr.Val.(pr.ConstExpr).Val.(CE), true
		return ce, ok
	}
	return nil, false
}
// Basically try to evalueate an expression at compile time.
// Does a best effor evaluation - tries to evaluate all expressions that involve constant expressions,
// otherwise returns them as is
func (ev *ExpressionEvaluator) TryEvaluateExpression(e *pr.Expr) (*pr.Expr, bool) {
	switch expr := e.Val.(type) {
	//if this is a constant expression, pass it on
	case pr.ConstExpr:
		if val := ev.extractValue(expr); val == nil {
			return e, false
		} else {
			switch v := val.(type) {
			case uint64:
				return pr.MakeConstexprU64(v), true
			case float64:
				return pr.MakeConstexprF64(v), true
			default:
				panic("Unsupported variable value extracted from evaluation context")
			}
		}
		//if this is an arthmetic expression, atttempt to evaluate it
	case pr.ArthExpr:
		return ev.performOperation(expr)
	case pr.RegExpr:
		return e, true
	case pr.DerefExpr:
		inner, ok := ev.TryEvaluateExpression(expr.Inner)
		return pr.MakeDeref(inner), ok
	case pr.OneRegOffsetExpr:
		var offset *pr.Expr
		var ok bool
		if expr.Offset != nil {
			offset, ok = ev.TryEvaluateExpression(expr.Offset)
			expr.Offset = offset
		}
		return &pr.Expr{Val: expr}, ok
	case pr.TwoRegOffsetExpr:
		var offset *pr.Expr
		var ok bool
		if expr.Offset != nil {
			offset, ok = ev.TryEvaluateExpression(expr.Offset)
			expr.Offset = offset
		}
		return &pr.Expr{Val: expr}, ok
	}
	return e, false
}
