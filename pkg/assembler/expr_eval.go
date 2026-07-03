package assembler

import (
	"fmt"
	"math"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
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
func TryConstEvalExprT[CE pr.CExpr](
	ev *ExpressionEvaluator, e *pr.Expr) (any, bool) {
	if expr, ok := ev.TryEvaluateExpression(e); ok && expr.IsConstexpr() {
		ce, ok := expr.Val.(pr.ConstExpr).Val.(CE), true
		return ce, ok
	}
	return nil, false
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

// TODO: make this work with the current paradigm
// maybe first do the fat tree stuff?
// func processDerefNestedArth(
// 	arthExpr pr.ArthExpr, nestLvl int) (DerefData, error) {
// 	ret := DerefData{
// 		Reg1:       lx.GetInvalidRegister(),
// 		Reg2:       lx.GetInvalidRegister(),
// 		OffsetOp:   INVALID,
// 		Offset:     INVALID,
// 		OffsetExpr: nil,
// 	}
// 	extract1RO := func(a, b *pr.Expr, op int) error {
// 		var reg, off *pr.Expr
// 		if a.IsRegexpr() && pr.IsConstexprType[pr.ConstExprILit](b) {
// 			reg = a
// 			off = b
// 		} else if b.IsRegexpr() && pr.IsConstexprType[pr.ConstExprILit](a) &&
// 			(op == vm.OP_TADD || op == vm.OP_TMUL) {
// 			reg = b
// 			off = a
// 		} else {
// 			return errors.InvalidDerefExpr(
// 				a.Line, a.Col,
// 				"In expression `%s`",
// 			)
// 		}
// 		ret.Ty = DEREF_T1RO
// 		ret.Reg1 = reg.Val.(pr.RegExpr).Reg
// 		ret.Offset = off.Val.(pr.ConstExpr).Val.(pr.ConstExprILit).Signed()
// 		ret.OffsetOp = op
// 		return nil
// 	}
// 	extract2RO := func(a, b *pr.Expr, op int) error {
// 		// r0 <op> r0
// 		if a.IsRegexpr() && b.IsRegexpr() {
// 			ret.Ty = DEREF_T2RO
// 			ret.Reg1 = a.Val.(pr.RegExpr).Reg
// 			ret.Reg2 = b.Val.(pr.RegExpr).Reg
// 			ret.Offset = int64(0)
// 			ret.OffsetOp = op
// 			// r0 <op> ( + )
// 		} else if a.IsRegexpr() && pr.IsArthexprType[pr.ArthExprAdd](b) &&
// 			nestLvl == 0 {
// 			if nestedD, err := processDerefNestedArth(
// 				b.Val.(pr.ArthExpr), nestLvl+1); err != nil {
// 				return err
// 			} else if nestedD.OffsetOp != vm.OP_TADD && nestedD.OffsetOp != vm.OP_TSUB {
// 				return errors.InvalidDerefExpr(
// 					b.Line, b.Col,
// 					"In expression `%s`",
// 				)
// 			} else {
// 				ret.Ty = DEREF_T2RO
// 				ret.Reg1 = a.Val.(pr.RegExpr).Reg
// 				ret.Reg2 = nestedD.Reg1
// 				ret.Offset = nestedD.Offset
// 				ret.OffsetOp = nestedD.OffsetOp
// 			}
// 			// ( + ) <op> r0
// 		} else if b.IsRegexpr() && pr.IsArthexprType[pr.ArthExprAdd](a) &&
// 			nestLvl == 0 &&
// 			op == vm.OP_TADD {
// 			if nestedD, err := processDerefNestedArth(
// 				a.Val.(pr.ArthExpr), nestLvl+1); err != nil {
// 				return err
// 			} else {
// 				ret.Ty = DEREF_T2RO
// 				ret.Reg1 = b.Val.(pr.RegExpr).Reg
// 				ret.Reg2 = nestedD.Reg1
// 				ret.Offset = nestedD.Offset
// 				ret.OffsetOp = nestedD.OffsetOp
// 			}
// 		} else {
// 			return errors.InvalidDerefExpr(
// 				a.Line, a.Col,
// 				"In expression `%s`",
// 			)
// 		}
// 		return nil
// 	}
// 	switch arth := arthExpr.Val.(type) {
// 	case pr.ArthExprAdd:
// 		if err := extract1RO(arth.A, arth.B, vm.OP_TADD); err != nil {
// 			return ret, err
// 		} else if err := extract2RO(arth.A, arth.B, vm.OP_TADD); err != nil {
// 			return ret, err
// 		}
// 	case pr.ArthExprSub:
// 		if err := extract1RO(arth.A, arth.B, vm.OP_TSUB); err != nil {
// 			return ret, err
// 		}
// 	case pr.ArthExprDiv:
// 		if err := extract1RO(arth.A, arth.B, vm.OP_TDIV); err != nil {
// 			return ret, err
// 		}
// 	case pr.ArthExprMul:
// 		if err := extract1RO(arth.A, arth.B, vm.OP_TMUL); err != nil {
// 			return ret, err
// 		} else if err := extract2RO(arth.A, arth.B, vm.OP_TMUL); err != nil {
// 			return ret, err
// 		}
//
// 	}
// 	switch {
// 	case IsConstexprType(arthExpr.B, CONSTEXPR_TREG) &&
// 		IsArthexprType(arthExpr.A, ARTHEXPR_TADD) &&
// 		nestLvl == 0 &&
// 		(arthExpr.Ty == ARTHEXPR_TADD):
//
// 		nestedD, err := ev.processDerefNestedArth(arthExpr.A.Val.(ArthExpr), nestLvl+1)
// 		if err != nil {
// 			return ret, err
// 		}
// 		ret.Ty = DEREF_T2RO
// 		ret.Reg1 = arthExpr.B.Val.(ConstExpr).UnpackAsRegisterData()
// 		ret.Reg2 = nestedD.Reg1
// 		ret.Offset = nestedD.Offset
// 		ret.OffsetOp = nestedD.OffsetOp
// 	case IsConstexprType(arthExpr.A, CONSTEXPR_TILIT) &&
// 		IsArthexprType(arthExpr.B, ARTHEXPR_TADD) &&
// 		nestLvl == 0 &&
// 		(arthExpr.Ty == ARTHEXPR_TADD):
//
// 		nestedD, err := ev.processDerefNestedArth(arthExpr.B.Val.(ArthExpr), nestLvl+1)
// 		if err != nil {
// 			return ret, err
// 		}
// 		if nestedD.OffsetOp != vm.OP_TADD || nestedD.Ty != DEREF_T2RO {
// 			em, _ := arthExpr.B.Emit()
// 			return ret, errors.FailedToParse("dereference expression",
// 				ev.currentStartToken.Line, ev.currentStartToken.Col,
// 				"Disallowed operation in expression, only addition "+
// 					"is allowed between registers in"+
// 					" this expression\n In expression: `%s`",
// 				em,
// 			)
// 		}
// 		ret.Ty = DEREF_T2RO
// 		ret.Reg1 = nestedD.Reg1
// 		ret.Reg2 = nestedD.Reg2
// 		ret.Offset = int64(arthExpr.A.Val.(ConstExpr).Val)
// 		ret.OffsetOp = nestedD.OffsetOp
// 	case IsConstexprType(arthExpr.B, CONSTEXPR_TILIT) &&
// 		IsArthexprType(arthExpr.A, ARTHEXPR_TADD) &&
// 		nestLvl == 0:
//
// 		nestedD, err := ev.processDerefNestedArth(arthExpr.A.Val.(ArthExpr), nestLvl+1)
// 		if err != nil {
// 			return ret, err
// 		}
// 		if nestedD.OffsetOp != vm.OP_TADD || nestedD.Ty != DEREF_T2RO {
// 			em, _ := arthExpr.B.Emit()
// 			return ret, errors.FailedToParse("dereference expression",
// 				ev.currentStartToken.Line, ev.currentStartToken.Col,
// 				"Disallowed operation in expression, only addition "+
// 					"is allowed between registers in"+
// 					" this expression\n In expression: `%s`",
// 				em,
// 			)
// 		}
// 		ret.Ty = DEREF_T2RO
// 		ret.Reg1 = nestedD.Reg1
// 		ret.Reg2 = nestedD.Reg2
// 		ret.Offset = int64(arthExpr.B.Val.(ConstExpr).Val)
// 		ret.OffsetOp = nestedD.OffsetOp
// 	default:
// 		em, _ := arthExpr.Emit()
// 		return ret, errors.FailedToParse(
// 			"dereference expression",
// 			ev.currentStartToken.Line, ev.currentStartToken.Col,
// 			"Invalid dereference expression `%s`",
// 			em,
// 		)
// 		// TODO: label dereference support
// 	}
// 	return ret, nil
// }

// probably does not need a reciever argument
func ProcessDeref(inner *pr.Expr) (DerefData, error) {
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
		// return processDerefNestedArth(expr, 0)
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

const (
	// expect a single register or a single immediate value
	EXPECT_R_OR_IMM = 1
	// expect a register, operator and then a register or an immediate value
	EXPECT_R_OP_R_OR_IMM = 3
	// expect two registers and an immediate value
	EXPECT_2R1IMM = 5
)

func ProcessDeref2(dexpr pr.DerefExpr) (DerefData, error) {
	ret := DerefData{}
	var flattened []lx.Token
	inner := dexpr.Inner
	if out, err := flattenDeref(inner); err != nil {
		return ret, err
	} else {
		flattened = out
	}
	switch len(flattened) {
	case EXPECT_R_OR_IMM:
		t0 := flattened[0]
		switch t0.Ty {
		case lx.TOKEN_TREG:
			ret = DerefData{
				Ty:       DEREF_T1RO,
				Reg1:     t0.Val.(lx.RegisterData),
				Offset:   0,
				OffsetOp: vm.OP_TADD,
			}
		case lx.TOKEN_TINTEGER_LIT:
			ret = DerefData{
				Ty:     DEREF_T0RO,
				Offset: int64(t0.Val.(uint64)),
			}
		default:
			return ret, errors.InvalidDerefExpr(
				t0.Line, t0.Col,
				"Disallowed derefernce parameter",
			)
		}
	case EXPECT_R_OP_R_OR_IMM:
		t0 := flattened[0]
		op := flattened[1]
		t2 := flattened[2]
		// swap them so the immediate value will always be on the right side
		if t2.Ty == lx.TOKEN_TREG &&
			(op.Ty == lx.TOKEN_TASTERISK || op.Ty == lx.TOKEN_TPLUS) {
			t0, t2 = t2, t0
		}
		if t0.Ty != lx.TOKEN_TREG {
			return ret, errors.InvalidDerefExpr(
				t0.Line, t0.Col,
				"Disallowed derefernce parameter",
			)
		}
		ret.Reg1 = t0.Val.(lx.RegisterData)
		switch op.Ty {
		case lx.TOKEN_TPLUS:
			ret.OffsetOp = vm.OP_TADD
		case lx.TOKEN_TMINUS:
			ret.OffsetOp = vm.OP_TSUB
		case lx.TOKEN_TASTERISK:
			ret.OffsetOp = vm.OP_TMUL
		case lx.TOKEN_TSLASH:
			ret.OffsetOp = vm.OP_TDIV
		default:
			return ret, errors.InvalidDerefExpr(
				op.Line, op.Col,
				"Bad expression",
			)
		}
		switch t2.Ty {
		case lx.TOKEN_TINTEGER_LIT:
			ret.Ty = DEREF_T1RO
			ret.Offset = int64(t2.Val.(uint64))
		case lx.TOKEN_TREG:
			switch ret.OffsetOp {
			case vm.OP_TADD:
				fallthrough
			case vm.OP_TSUB:
				break
			default:
				return ret, errors.InvalidDerefExpr(
					op.Line, op.Col,
					"Disallowed operator",
				)
			}
			ret.Ty = DEREF_T2RO
			ret.Offset = 0
			ret.Reg2 = t2.Val.(lx.RegisterData)
		default:
			return ret, errors.InvalidDerefExpr(
				t2.Line, t2.Col,
				"Disallowed derefernce parameter",
			)
		}
	case EXPECT_2R1IMM:
	}

	return ret, nil
}

func flattenDeref(expr *pr.Expr) ([]lx.Token, error) {
	flattened := make([]lx.Token, 0)
	if out, err := flattenImpl(expr, flattened, 0); err != nil {
		return nil, err
	} else {
		flattened = out
	}
	return flattened, nil
}

func flattenImpl(expr *pr.Expr, out []lx.Token, depth int) ([]lx.Token, error) {
	tok := lx.Token{
		Col:  expr.Col,
		Line: expr.Line,
	}
	switch v := expr.Val.(type) {
	case pr.ConstExprILit:
		tok.Ty = lx.TOKEN_TINTEGER_LIT
		tok.Val = v.Integer
		out = append(out, tok)
	case pr.RegExpr:
		tok.Ty = lx.TOKEN_TREG
		tok.Val = v.Reg
		out = append(out, tok)
	case pr.ArthExpr:
		if depth > 0 {
			return out, errors.InvalidDerefExpr(
				expr.Line, expr.Col,
				"Expression too complex",
			)
		}
		switch arth := v.Val.(type) {
		case pr.ArthExprAdd:
			if fa, err := flattenImpl(arth.A, out, depth+1); err != nil {
				return out, err
			} else {
				out = fa
			}
			tok.Ty = lx.TOKEN_TPLUS
			out = append(out, tok)
			if fb, err := flattenImpl(arth.B, out, depth+1); err != nil {
				return out, err
			} else {
				out = fb
			}
		case pr.ArthExprSub:
			if fa, err := flattenImpl(arth.A, out, depth+1); err != nil {
				return out, err
			} else {
				out = fa
			}
			tok.Ty = lx.TOKEN_TMINUS
			out = append(out, tok)
			if fb, err := flattenImpl(arth.B, out, depth+1); err != nil {
				return out, err
			} else {
				out = fb
			}
		case pr.ArthExprMul:
			if fa, err := flattenImpl(arth.A, out, depth+1); err != nil {
				return out, err
			} else {
				out = fa
			}
			tok.Ty = lx.TOKEN_TASTERISK
			out = append(out, tok)
			if fb, err := flattenImpl(arth.B, out, depth+1); err != nil {
				return out, err
			} else {
				out = fb
			}
		case pr.ArthExprDiv:
			if fa, err := flattenImpl(arth.A, out, depth+1); err != nil {
				return out, err
			} else {
				out = fa
			}
			tok.Ty = lx.TOKEN_TSLASH
			out = append(out, tok)
			if fb, err := flattenImpl(arth.B, out, depth+1); err != nil {
				return out, err
			} else {
				out = fb
			}
		default:
			return out, errors.InvalidDerefExpr(
				expr.Line, expr.Col,
				"Unsupported expression type",
			)
		}
	default:
		return out, errors.InvalidDerefExpr(
			expr.Line, expr.Col,
			"Invalid parameter in expression",
		)
	}
	return out, nil
}
