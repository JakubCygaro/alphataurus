package assembler

import (
	"math"
)

const (
	EXPR_TCONST = iota
	EXPR_TARTH
	EXPR_TDEREF
)
const (
	ARTHEXPR_TADD = TOKEN_TPLUS
	ARTHEXPR_TSUB = TOKEN_TMINUS
	ARTHEXPR_TMUL = TOKEN_TASTERISK
	ARTHEXPR_TDIV = TOKEN_TSLASH
)
const (
	INVALID = -1
	// r0
	// sp
	CONSTEXPR_TREG = 0
	// 1313
	CONSTEXPR_TILIT = iota
	// 1.23
	CONSTEXPR_TFLIT
)

type Expr struct {
	Ty  int
	Val any
}

type ConstExpr struct {
	Ty  int
	Val uint64
}

func opTy(a, b*ConstExpr) int {
	if a.Ty == CONSTEXPR_TFLIT || a.Ty == CONSTEXPR_TREG{
		return a.Ty
	} else {
		return b.Ty
	}
}

func (a*ConstExpr) Add(b*ConstExpr) (ConstExpr, bool) {
	ty := opTy(a, b)
	if ty == CONSTEXPR_TREG {
		return ConstExpr{}, false
	}
	switch ty {
	case CONSTEXPR_TFLIT:
		aV, bV := math.Float64frombits(a.Val), math.Float64frombits(b.Val)
		return ConstExpr{
			Ty: ty,
			Val: math.Float64bits(aV + bV),
		}, true
	case CONSTEXPR_TILIT:
		aV, bV := a.Val, b.Val
		return ConstExpr{
			Ty: ty,
			Val: aV + bV,
		}, true
	}
	return ConstExpr{}, false
}
func (a*ConstExpr) Sub(b*ConstExpr) (ConstExpr, bool) {
	ty := opTy(a, b)
	if ty == CONSTEXPR_TREG {
		return ConstExpr{}, false
	}
	switch ty {
	case CONSTEXPR_TFLIT:
		aV, bV := math.Float64frombits(a.Val), math.Float64frombits(b.Val)
		return ConstExpr{
			Ty: ty,
			Val: math.Float64bits(aV - bV),
		}, true
	case CONSTEXPR_TILIT:
		aV, bV := a.Val, b.Val
		return ConstExpr{
			Ty: ty,
			Val: aV - bV,
		}, true
	}
	return ConstExpr{}, false
}
func (a*ConstExpr) Mul(b*ConstExpr) (ConstExpr, bool) {
	ty := opTy(a, b)
	if ty == CONSTEXPR_TREG {
		return ConstExpr{}, false
	}
	switch ty {
	case CONSTEXPR_TFLIT:
		aV, bV := math.Float64frombits(a.Val), math.Float64frombits(b.Val)
		return ConstExpr{
			Ty: ty,
			Val: math.Float64bits(aV * bV),
		}, true
	case CONSTEXPR_TILIT:
		aV, bV := a.Val, b.Val
		return ConstExpr{
			Ty: ty,
			Val: aV * bV,
		}, true
	}
	return ConstExpr{}, false
}
func (a*ConstExpr) Div(b*ConstExpr) (ConstExpr, bool) {
	ty := opTy(a, b)
	if ty == CONSTEXPR_TREG {
		return ConstExpr{}, false
	}
	switch ty {
	case CONSTEXPR_TFLIT:
		aV, bV := math.Float64frombits(a.Val), math.Float64frombits(b.Val)
		return ConstExpr{
			Ty: ty,
			Val: math.Float64bits(aV / bV),
		}, true
	case CONSTEXPR_TILIT:
		aV, bV := a.Val, b.Val
		return ConstExpr{
			Ty: ty,
			Val: aV / bV,
		}, true
	}
	return ConstExpr{}, false
}

type ArthExpr struct {
	Ty   int
	A, B Expr
}

// [r0]
// [r1+1]
// [bp+1]
// [bp+r0]
type DerefExpr struct {
	Inner ArthExpr
}
// basically try to evalueate an expression at compile time
// fails if the expression contains any registers or a dereference
func TryEvaluateExpression(e *Expr) (ConstExpr, bool) {
	ret := ConstExpr { Ty: INVALID }
	switch e.Ty {
	case EXPR_TCONST:
		eAsConst := e.Val.(ConstExpr)
		if eAsConst.Ty == CONSTEXPR_TREG {
			return eAsConst, false
		}
		ret = eAsConst
	case EXPR_TARTH:
		eAsArth := e.Val.(ArthExpr)
		if evalA, ok := TryEvaluateExpression(&eAsArth.A); !ok {
			return ret, false
		} else if evalB, ok := TryEvaluateExpression(&eAsArth.B); !ok {
			return ret, false
		} else {
			switch eAsArth.Ty {
				case ARTHEXPR_TADD:
					return evalA.Add(&evalB)
				case ARTHEXPR_TSUB:
					return evalA.Sub(&evalB)
				case ARTHEXPR_TMUL:
					return evalA.Mul(&evalB)
				case ARTHEXPR_TDIV:
					return evalA.Div(&evalB)
				default:
					return ret, false
			}
		}
	default:
		return ret, false
	}
	return ret, true
}
