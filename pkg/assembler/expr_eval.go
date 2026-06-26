package assembler

import (
	"fmt"
	"math"

	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

type EvaluationContext struct {
	Variables map[string]any
}

type ExpressionEvaluator struct {
	ctx EvaluationContext
}

// Extract the value of the constant expression into either uint64 or float64
// if the value cannot be extracted it is returned as nil
func (ev *ExpressionEvaluator) extractValue(cexpr pr.ConstExpr) any {
	switch v := cexpr.Val.(type) {
	case pr.ConstExprIden:
		f, _ := ev.ctx.Variables[v.Ident]
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
			return pr.MakeArth[pr.ArthExprAdd](
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
			return pr.MakeArth[pr.ArthExprAdd](
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
			return pr.MakeArth[pr.ArthExprAdd](
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
			return pr.MakeArth[pr.ArthExprAdd](
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

// func TryEvaluatePruneExpression(e *pr.Expr) (*pr.Expr, bool) {
// 	ret1 := e
// 	var ret2 bool
// 	for pruned := true; pruned; {
// 		ret1, ret2 = TryEvaluateExpression(ret1)
// 		pruned = PruneExpression(ret1, nil, 0)
// 	}
// 	return ret1, ret2
// }
//
// func TryConstEvaluatePruneExpression(e *pr.Expr) (pr.ConstExpr, bool) {
// 	if expr, ok := TryEvaluatePruneExpression(e); ok && expr.Ty == EXPR_TCONST {
// 		return expr.Val.(pr.ConstExpr), true
// 	}
// 	return pr.ConstExpr{Ty: INVALID}, false
// }

// Does what TryEvaluateExpression does, but at the end verifies that the expression is constant
func (ev *ExpressionEvaluator) TryConstEvaluateExpression(e *pr.Expr) (pr.ConstExpr, bool) {
	if expr, ok := ev.TryEvaluateExpression(e); ok && expr.IsConstexpr() {
		return expr.Val.(pr.ConstExpr), true
	}
	return pr.ConstExpr{}, false
}

// func (ev *ExpressionEvaluator) tryEvaluateBinaryExpression(
// 	a, b *pr.Expr) (*pr.Expr, bool) {
// 	evalA, okA := ev.TryEvaluateExpression(a)
// 	evalB, okB := ev.TryEvaluateExpression(b)
// 	// in case both have been succesfully evaluated and both are const,
// 	// try to evaluate the arthmetic expression into a constant expression
// 	bothOk := okA && okB
// 	bothConst := evalA.IsConstexpr() && evalB.IsConstexpr()
// 	if bothOk && bothConst {
// 		valA := evalA.Val.(pr.ConstExpr)
// 		valB := evalB.Val.(pr.ConstExpr)
// 		res, ok := PerformOperation(&valA, &valB, eAsArth.Ty)
// 		if ok {
// 			return MakeConstexpr(res), true
// 		} else {
// 			return e, false
// 		}
// 		// in case only the left one is const and the right is an Arth expr
// 	} else if okB && IsArthexpr(evalA, ARTHEXPR_TADD) &&
// 		(IsConstexpr(evalB, CONSTEXPR_TILIT) || IsConstexpr(evalB, CONSTEXPR_TFLIT)) {
// 		leftArth := evalA.Val.(pr.ArthExpr)
// 		rightConst := evalB.Val.(pr.ConstExpr)
// 		if IsConstexpr(leftArth.A, CONSTEXPR_TILIT) ||
// 			IsConstexpr(leftArth.A, CONSTEXPR_TFLIT) {
// 			leftArthA := leftArth.A.Val.(pr.ConstExpr)
// 			res, ok := rightConst.Add(&leftArthA)
// 			if !ok {
// 				return e, false
// 			}
// 			return MakeArth(
// 				leftArth.B,
// 				MakeConstexpr(res),
// 				ARTHEXPR_TADD,
// 			), true
// 		} else if IsConstexpr(leftArth.B, CONSTEXPR_TILIT) ||
// 			IsConstexpr(leftArth.B, CONSTEXPR_TFLIT) {
// 			leftArthB := leftArth.B.Val.(pr.ConstExpr)
// 			res, ok := rightConst.Add(&leftArthB)
// 			if !ok {
// 				return e, false
// 			}
// 			return MakeArth(
// 				MakeConstexpr(res),
// 				leftArth.A,
// 				ARTHEXPR_TADD,
// 			), true
// 		}
// 	} else if okA && IsArthexpr(evalB, ARTHEXPR_TADD) &&
// 		(IsConstexpr(evalA, CONSTEXPR_TILIT) ||
// 			IsConstexpr(evalA, CONSTEXPR_TFLIT)) {
// 		rightArth := evalB.Val.(pr.ArthExpr)
// 		leftConst := evalA.Val.(pr.ConstExpr)
// 		if IsConstexpr(rightArth.A, CONSTEXPR_TILIT) ||
// 			IsConstexpr(rightArth.A, CONSTEXPR_TFLIT) {
// 			rightArthA := rightArth.A.Val.(pr.ConstExpr)
// 			res, ok := leftConst.Add(&rightArthA)
// 			if !ok {
// 				return e, false
// 			}
// 			return MakeArth(
// 				MakeConstexpr(res),
// 				rightArth.B,
// 				ARTHEXPR_TADD,
// 			), true
// 		} else if IsConstexpr(rightArth.B, CONSTEXPR_TILIT) ||
// 			IsConstexpr(rightArth.B, CONSTEXPR_TFLIT) {
// 			rightArthB := rightArth.B.Val.(pr.ConstExpr)
// 			res, ok := leftConst.Add(&rightArthB)
// 			if !ok {
// 				return e, false
// 			}
// 			return MakeArth(
// 				rightArth.A,
// 				MakeConstexpr(res),
// 				ARTHEXPR_TADD,
// 			), true
// 		}
// 	}
// 	// if all else fails, just pass the best effort attempt with whatever could've been evaluated
// 	return MakeArth(evalA, evalB, eAsArth.Ty), false
// }

// Basically try to evalueate an expression at compile time.
// Does a best effor evaluation - tries to evaluate all expressions that involve constant expressions,
// otherwise returns them as is
func (ev *ExpressionEvaluator) TryEvaluateExpression(e *pr.Expr) (*pr.Expr, bool) {
	switch expr := e.Val.(type) {
	//if this is a constant expression, pass it on
	case pr.ConstExpr:
		return e, true
		//if this is an arthmetic expression, atttempt to evaluate it
	case pr.ArthExpr:
		return ev.performOperation(expr)
		// switch arth := expr.Val.(type) {
		// case pr.ArthExprAdd:
		// 	return ev.tryEvaluateBinaryExpression(arth.A, arth.B)
		// case pr.ArthExprSub:
		// 	return ev.tryEvaluateBinaryExpression(arth.A, arth.B)
		// case pr.ArthExprMul:
		// 	return ev.tryEvaluateBinaryExpression(arth.A, arth.B)
		// case pr.ArthExprDiv:
		// 	return ev.tryEvaluateBinaryExpression(arth.A, arth.B)
		// }
	case pr.DerefExpr:
		inner, ok := ev.TryEvaluateExpression(expr.Inner)
		return pr.MakeDeref(inner), ok
	}
	return e, false
}

// traverse the tree, find leaves that are not integer/float literals
// then try to move them up the tree so that addition chains of literals get coalesed
//
// basi-fucking-ly:
// such an expression (1+r0+r1+2) tree will be transformed:
//
//	 /\+
//	1 /\+
//	 r0/\+
//	  r1 2
//
// into this (r0+r1+1+2):
//
//	/\+
//
// r1 /\+
//
//	r0/\+
//	 1  2
//
// which can then be evaluated once more
// func PruneExpression(e, swap *Expr, pruneWith int) bool {
// 	switch pruneWith {
// 	case 0:
// 		if arthTy, ok := Aspr.ArthExprTy(e); ok {
// 			if _, isAssoc := associativeOperators[arthTy]; isAssoc {
// 				pruneWith = arthTy
// 			} else {
// 				return false
// 			}
// 		}
// 	default:
// 		if _, isAssoc := associativeOperators[pruneWith]; !isAssoc {
// 			return false
// 		}
// 	}
// 	switch {
// 	// if this is an associative operator expression
// 	case IsArthexpr(e, pruneWith):
// 		arthE := e.Val.(pr.ArthExpr)
// 		switch {
// 		// if it has one node as a literal node and the other as a different associative expr of the same op
// 		case (IsConstexpr(arthE.A, CONSTEXPR_TILIT) || IsConstexpr(arthE.A, CONSTEXPR_TFLIT)) &&
// 			IsArthexpr(arthE.B, pruneWith):
// 			//then recurse into it with the other literal as swap
// 			if swap != nil {
// 				return PruneExpression(arthE.B, swap, pruneWith)
// 			} else {
// 				return PruneExpression(arthE.B, arthE.A, pruneWith)
// 			}
// 		//same case but branches are flipped
// 		case (IsConstexpr(arthE.B, CONSTEXPR_TILIT) || IsConstexpr(arthE.B, CONSTEXPR_TFLIT)) &&
// 			IsArthexpr(arthE.A, pruneWith):
// 			if swap != nil {
// 				return PruneExpression(arthE.A, swap, pruneWith)
// 			} else {
// 				return PruneExpression(arthE.A, arthE.B, pruneWith)
// 			}
// 		// we've hit the bottom and the other guy is not a comp-time evaluable expression, and swap is not nil
// 		case (IsConstexpr(arthE.A, CONSTEXPR_TILIT) || IsConstexpr(arthE.A, CONSTEXPR_TFLIT)) &&
// 			(!IsConstexpr(arthE.B, CONSTEXPR_TILIT) && !IsConstexpr(arthE.B, CONSTEXPR_TFLIT)) &&
// 			swap != nil:
// 			*arthE.B, *swap = *swap, *arthE.B
// 			return true
// 		// the same but in reverse
// 		case (IsConstexpr(arthE.B, CONSTEXPR_TILIT) || IsConstexpr(arthE.B, CONSTEXPR_TFLIT)) &&
// 			(!IsConstexpr(arthE.A, CONSTEXPR_TILIT) && !IsConstexpr(arthE.A, CONSTEXPR_TFLIT)) &&
// 			swap != nil:
// 			*arthE.A, *swap = *swap, *arthE.A
// 			return true
// 		// if this is a nested expression of the same op then recurse into it
// 		case IsArthexpr(arthE.A, pruneWith):
// 			return PruneExpression(arthE.A, swap, pruneWith)
// 		case IsArthexpr(arthE.B, pruneWith):
// 			return PruneExpression(arthE.B, swap, pruneWith)
// 		}
// 	}
// 	return false
// }

const INVALID = -1

const (
	DEREF_T0RO = iota
	DEREF_T1RO
	DEREF_T2RO
)

type DerefData struct {
	Ty         int
	Reg1, Reg2 lx.RegisterData
	// plus or minus
	OffsetOp int
	Offset   int64
	// in case there are labels to resolve
	OffsetExpr *pr.Expr
}

func (ev *ExpressionEvaluator) processDerefNestedArth(
	arthExpr pr.ArthExpr, nestLvl int) (DerefData, error) {
	ret := DerefData{
		Reg1:       lx.GetInvalidRegister(),
		Reg2:       lx.GetInvalidRegister(),
		OffsetOp:   INVALID,
		Offset:     INVALID,
		OffsetExpr: nil,
	}
	switch {
	case pr.IsConstexprType[pr.RegExpr](arthExpr.A) &&
		IsConstexprType(arthExpr.B, CONSTEXPR_TILIT):

		ret.Ty = DEREF_T1RO
		ret.Reg1 = arthExpr.A.Val.(ConstExpr).UnpackAsRegisterData()
		ret.Offset = int64(arthExpr.B.Val.(ConstExpr).Val)
		ret.OffsetOp = arthExpr.GetVMOpType()
	case IsConstexprType(arthExpr.A, CONSTEXPR_TILIT) &&
		IsConstexprType(arthExpr.B, CONSTEXPR_TREG) &&
		(arthExpr.Ty == ARTHEXPR_TADD):

		ret.Ty = DEREF_T1RO
		ret.Reg1 = arthExpr.B.Val.(ConstExpr).UnpackAsRegisterData()
		ret.Offset = int64(arthExpr.A.Val.(ConstExpr).Val)
		ret.OffsetOp = vm.OP_TADD
	case IsConstexprType(arthExpr.A, CONSTEXPR_TREG) &&
		IsConstexprType(arthExpr.B, CONSTEXPR_TREG) &&
		(arthExpr.Ty == ARTHEXPR_TADD):

		ret.Ty = DEREF_T2RO
		ret.Reg1 = arthExpr.A.Val.(ConstExpr).UnpackAsRegisterData()
		ret.Reg2 = arthExpr.B.Val.(ConstExpr).UnpackAsRegisterData()
		ret.Offset = int64(0)
		ret.OffsetOp = vm.OP_TADD
	case IsConstexprType(arthExpr.A, CONSTEXPR_TREG) &&
		IsArthexprType(arthExpr.B, ARTHEXPR_TADD) &&
		nestLvl == 0:

		nestedD, err := ev.processDerefNestedArth(arthExpr.B.Val.(ArthExpr), nestLvl+1)
		if err != nil {
			return ret, err
		}
		if nestedD.OffsetOp != vm.OP_TADD && nestedD.OffsetOp != vm.OP_TSUB {
			em, _ := arthExpr.B.Emit()
			return ret, errors.FailedToParse("dereference expression",
				ev.currentStartToken.Line, ev.currentStartToken.Col,
				"Disallowed operation in expression, only addition or subtraction "+
					"is allowed for this expression\n In expression: `%s`",
				em,
			)
		}
		ret.Ty = DEREF_T2RO
		ret.Reg1 = arthExpr.A.Val.(ConstExpr).UnpackAsRegisterData()
		ret.Reg2 = nestedD.Reg1
		ret.Offset = nestedD.Offset
		ret.OffsetOp = nestedD.OffsetOp
	case IsConstexprType(arthExpr.B, CONSTEXPR_TREG) &&
		IsArthexprType(arthExpr.A, ARTHEXPR_TADD) &&
		nestLvl == 0 &&
		(arthExpr.Ty == ARTHEXPR_TADD):

		nestedD, err := ev.processDerefNestedArth(arthExpr.A.Val.(ArthExpr), nestLvl+1)
		if err != nil {
			return ret, err
		}
		ret.Ty = DEREF_T2RO
		ret.Reg1 = arthExpr.B.Val.(ConstExpr).UnpackAsRegisterData()
		ret.Reg2 = nestedD.Reg1
		ret.Offset = nestedD.Offset
		ret.OffsetOp = nestedD.OffsetOp
	case IsConstexprType(arthExpr.A, CONSTEXPR_TILIT) &&
		IsArthexprType(arthExpr.B, ARTHEXPR_TADD) &&
		nestLvl == 0 &&
		(arthExpr.Ty == ARTHEXPR_TADD):

		nestedD, err := ev.processDerefNestedArth(arthExpr.B.Val.(ArthExpr), nestLvl+1)
		if err != nil {
			return ret, err
		}
		if nestedD.OffsetOp != vm.OP_TADD || nestedD.Ty != DEREF_T2RO {
			em, _ := arthExpr.B.Emit()
			return ret, errors.FailedToParse("dereference expression",
				ev.currentStartToken.Line, ev.currentStartToken.Col,
				"Disallowed operation in expression, only addition "+
					"is allowed between registers in"+
					" this expression\n In expression: `%s`",
				em,
			)
		}
		ret.Ty = DEREF_T2RO
		ret.Reg1 = nestedD.Reg1
		ret.Reg2 = nestedD.Reg2
		ret.Offset = int64(arthExpr.A.Val.(ConstExpr).Val)
		ret.OffsetOp = nestedD.OffsetOp
	case IsConstexprType(arthExpr.B, CONSTEXPR_TILIT) &&
		IsArthexprType(arthExpr.A, ARTHEXPR_TADD) &&
		nestLvl == 0:

		nestedD, err := ev.processDerefNestedArth(arthExpr.A.Val.(ArthExpr), nestLvl+1)
		if err != nil {
			return ret, err
		}
		if nestedD.OffsetOp != vm.OP_TADD || nestedD.Ty != DEREF_T2RO {
			em, _ := arthExpr.B.Emit()
			return ret, errors.FailedToParse("dereference expression",
				ev.currentStartToken.Line, ev.currentStartToken.Col,
				"Disallowed operation in expression, only addition "+
					"is allowed between registers in"+
					" this expression\n In expression: `%s`",
				em,
			)
		}
		ret.Ty = DEREF_T2RO
		ret.Reg1 = nestedD.Reg1
		ret.Reg2 = nestedD.Reg2
		ret.Offset = int64(arthExpr.B.Val.(ConstExpr).Val)
		ret.OffsetOp = nestedD.OffsetOp
	default:
		em, _ := arthExpr.Emit()
		return ret, errors.FailedToParse(
			"dereference expression",
			ev.currentStartToken.Line, ev.currentStartToken.Col,
			"Invalid dereference expression `%s`",
			em,
		)
		// TODO: label dereference support
	}
	return ret, nil
}

func (ev *ExpressionEvaluator) processDeref(inner *pr.Expr) (DerefData, error) {
	ret := DerefData{
		Reg1:       lx.GetInvalidRegister(),
		Reg2:       lx.GetInvalidRegister(),
		OffsetOp:   INVALID,
		Offset:     INVALID,
		OffsetExpr: nil,
	}
	switch expr := inner.Val.(type) {
	case pr.ConstExpr:
		switch v := expr.Val.(type) {
		case pr.ConstExprILit:
			ret.Ty = DEREF_T0RO
			ret.Offset = int64(v.Integer)
		default:
			em, _ := inner.Emit()
			return ret, fmt.Errorf(
				"Invalid single parameter dereference expression. "+
					"In expression : `%s` at (%v:%v)",
				em, inner.Line, inner.Col,
			)
			// return ret, errors.FailedToParse("dereference expression",
			// 	p.currentStartToken.Line, p.currentStartToken.Col,
			// 	"Invalid single parameter dereference expression."+
			// 		"In Expression `%s`",
			// 	em,
			// )
		}
	case pr.RegExpr:
		ret.Ty = DEREF_T1RO
		regD := expr.Reg
		ret.Reg1 = regD
		ret.Offset = int64(0)
		ret.OffsetOp = vm.OP_TADD
	case pr.ArthExpr:
		return ev.processDerefNestedArth(expr, 0)
	default:
		// em, _ := inner.Emit()
		// return ret, errors.FailedToParse("dereference expression",
		// 	p.currentStartToken.Line, p.currentStartToken.Col,
		// 	"Invalid dereference expression."+
		// 		"In Expression `%s`",
		// 	em,
		// )
		em, _ := inner.Emit()
		return ret, fmt.Errorf(
			"Invalid dereference expression. "+
				"In expression : `%s` at (%v:%v)",
			em, inner.Line, inner.Col,
		)
	}
	return ret, nil
}
