package assembler

import (
	"fmt"
	"math"

	"github.com/JakubCygaro/alphataurus/internal/vm"
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

func MakeConstexpr(c ConstExpr) Expr {
	return Expr{
		Ty:  EXPR_TCONST,
		Val: c,
	}
}
func MakeConstexprU64(v uint64) Expr {
	return Expr{
		Ty: EXPR_TCONST,
		Val: ConstExpr{
			Ty:  CONSTEXPR_TILIT,
			Val: v,
		},
	}
}
func MakeConstexprI64(v int64) Expr {
	return Expr{
		Ty: EXPR_TCONST,
		Val: ConstExpr{
			Ty:  CONSTEXPR_TILIT,
			Val: uint64(v),
		},
	}
}
func MakeConstexprF64(v float64) Expr {
	return Expr{
		Ty: EXPR_TCONST,
		Val: ConstExpr{
			Ty:  CONSTEXPR_TFLIT,
			Val: math.Float64bits(v),
		},
	}
}
func MakeConstexprR(r int) Expr {
	return Expr{
		Ty: EXPR_TCONST,
		Val: ConstExpr{
			Ty:  CONSTEXPR_TREG,
			Val: uint64(r),
		},
	}
}
func MakeArth(a, b Expr, ty int) Expr {
	return Expr{
		Ty: EXPR_TARTH,
		Val: ArthExpr{
			Ty: ty,
			A:  a,
			B:  b,
		},
	}
}
func MakeDeref(inner Expr) Expr {
	return Expr{
		Ty: EXPR_TDEREF,
		Val: DerefExpr{
			Inner: inner,
		},
	}
}

func (e *ConstExpr) AsFloat() float64 {
	switch e.Ty {
	case CONSTEXPR_TFLIT:
		return math.Float64frombits(e.Val)
	case CONSTEXPR_TILIT:
		return float64(e.Val)
	default:
		return math.NaN()
	}
}
func opTy(a, b *ConstExpr) int {
	if a.Ty == CONSTEXPR_TFLIT || a.Ty == CONSTEXPR_TREG {
		return a.Ty
	} else {
		return b.Ty
	}
}

func (a *ConstExpr) Add(b *ConstExpr) (ConstExpr, bool) {
	ty := opTy(a, b)
	if ty == CONSTEXPR_TREG {
		return ConstExpr{}, false
	}
	switch ty {
	case CONSTEXPR_TFLIT:
		aV, bV := a.AsFloat(), b.AsFloat()
		return ConstExpr{
			Ty:  ty,
			Val: math.Float64bits(aV + bV),
		}, true
	case CONSTEXPR_TILIT:
		aV, bV := a.Val, b.Val
		return ConstExpr{
			Ty:  ty,
			Val: aV + bV,
		}, true
	}
	return ConstExpr{}, false
}
func (a *ConstExpr) Sub(b *ConstExpr) (ConstExpr, bool) {
	ty := opTy(a, b)
	if ty == CONSTEXPR_TREG {
		return ConstExpr{}, false
	}
	switch ty {
	case CONSTEXPR_TFLIT:
		aV, bV := a.AsFloat(), b.AsFloat()
		return ConstExpr{
			Ty:  ty,
			Val: math.Float64bits(aV - bV),
		}, true
	case CONSTEXPR_TILIT:
		aV, bV := a.Val, b.Val
		return ConstExpr{
			Ty:  ty,
			Val: aV - bV,
		}, true
	}
	return ConstExpr{}, false
}
func (a *ConstExpr) Mul(b *ConstExpr) (ConstExpr, bool) {
	ty := opTy(a, b)
	if ty == CONSTEXPR_TREG {
		return ConstExpr{}, false
	}
	switch ty {
	case CONSTEXPR_TFLIT:
		aV, bV := a.AsFloat(), b.AsFloat()
		return ConstExpr{
			Ty:  ty,
			Val: math.Float64bits(aV * bV),
		}, true
	case CONSTEXPR_TILIT:
		aV, bV := a.Val, b.Val
		return ConstExpr{
			Ty:  ty,
			Val: aV * bV,
		}, true
	}
	return ConstExpr{}, false
}
func (a *ConstExpr) Div(b *ConstExpr) (ConstExpr, bool) {
	ty := opTy(a, b)
	if ty == CONSTEXPR_TREG {
		return ConstExpr{}, false
	}
	switch ty {
	case CONSTEXPR_TFLIT:
		aV, bV := a.AsFloat(), b.AsFloat()
		return ConstExpr{
			Ty:  ty,
			Val: math.Float64bits(aV / bV),
		}, true
	case CONSTEXPR_TILIT:
		aV, bV := a.Val, b.Val
		return ConstExpr{
			Ty:  ty,
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
	Inner Expr
}

// Does what TryEvaluateExpression does, but at the end verifies that the expression is constant
func TryConstEvaluateExpression(e *Expr) (ConstExpr, bool) {
	if expr, ok := TryEvaluateExpression(e); ok && expr.Ty == EXPR_TCONST {
		return expr.Val.(ConstExpr), true
	}
	return ConstExpr{Ty: INVALID}, false
}

// Basically try to evalueate an expression at compile time.
// Does a best effor evaluation - tries to evaluate all expressions that involve constant expressions,
// otherwise returns them as is
func TryEvaluateExpression(e *Expr) (Expr, bool) {
	switch e.Ty {
	//if this is a constant expression, pass it on
	case EXPR_TCONST:
		// eAsConst := e.Val.(ConstExpr)
		// if eAsConst.Ty == CONSTEXPR_TREG {
		// 	return eAsConst, false
		// }
		return *e, true
		//if this is an arthmetic expression, atttempt to evaluate it
	case EXPR_TARTH:
		eAsArth := e.Val.(ArthExpr)
		evalA, okA := TryEvaluateExpression(&eAsArth.A)
		evalB, okB := TryEvaluateExpression(&eAsArth.B)
		// in case both have been succesfully evaluated and both are const,
		// try to evaluate the arthmetic expression into a constant expression
		bothOk := okA && okB
		bothConst := evalA.Ty == EXPR_TCONST && evalB.Ty == EXPR_TCONST
		if bothOk && bothConst {
			valA := evalA.Val.(ConstExpr)
			valB := evalB.Val.(ConstExpr)
			ok := true
			var res ConstExpr
			switch eAsArth.Ty {
			case ARTHEXPR_TADD:
				res, ok = valA.Add(&valB)
			case ARTHEXPR_TSUB:
				res, ok = valA.Sub(&valB)
			case ARTHEXPR_TMUL:
				res, ok = valA.Mul(&valB)
			case ARTHEXPR_TDIV:
				res, ok = valA.Div(&valB)
			default:
				return *e, false
			}
			if ok {
				return MakeConstexpr(res), true
			}
		}
		// if all else fails, just pass the best effort attempt with whatever could've benn evaluated
		return MakeArth(evalA, evalB, eAsArth.Ty), false
	case EXPR_TDEREF:
		inner := e.Val.(DerefExpr).Inner
		inner, ok := TryEvaluateExpression(&inner)
		return MakeDeref(inner), ok
	default:
		return *e, false
	}
	return *e, false
}
func (e ArthExpr) Emit() (string, error) {
	if a, err := e.A.Emit(); err != nil {
		return "", err
	} else if b, err := e.B.Emit(); err != nil {
		return "", err
	} else {
		switch e.Ty {
		case ARTHEXPR_TADD:
			return fmt.Sprintf("%s + %s", a, b), nil
		case ARTHEXPR_TSUB:
			return fmt.Sprintf("%s - %s", a, b), nil
		case ARTHEXPR_TMUL:
			return fmt.Sprintf("%s * %s", a, b), nil
		case ARTHEXPR_TDIV:
			return fmt.Sprintf("%s / %s", a, b), nil
		default:
			return "", fmt.Errorf("Invalid arthmetic expression type")
		}
	}
}
func (e ConstExpr) Emit() (string, error) {
	switch e.Ty {
	case CONSTEXPR_TREG:
		if e.Val <= vm.GP_REG_MAX {
			return fmt.Sprintf("r%d", e.Val), nil
		} else if e.Val == vm.SP_IDX {
			return "sp", nil
		} else if e.Val == vm.BP_IDX {
			return "bp", nil
		} else if e.Val == vm.IP_IDX {
			return "ip", nil
		} else {
			return "", fmt.Errorf("Invalid register type %v", e.Val)
		}
	case CONSTEXPR_TILIT:
		return fmt.Sprintf("%v", e.Val), nil
	case CONSTEXPR_TFLIT:
		return fmt.Sprintf("%v", math.Float64frombits(e.Val)), nil
	default:
		return "", fmt.Errorf("Invalid constant expression type %v", e.Val)
	}
}
func (e *Expr) Emit() (string, error) {
	switch e.Ty {
	case EXPR_TCONST:
		return e.Val.(ConstExpr).Emit()
	case EXPR_TDEREF:
		deref, ok := e.Val.(DerefExpr)
		if !ok {
			return "", nil
		}
		emit, err := deref.Inner.Emit()
		return fmt.Sprintf("[%s]", emit), err
	case EXPR_TARTH:
		inner, err := e.Val.(ArthExpr).Emit()
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("(%s)", inner), nil
	default:
		return "", fmt.Errorf("Invalid expression type")
	}
}
