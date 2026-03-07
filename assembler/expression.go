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
	// abcd
	CONSTEXPR_TIDENT
)

type void struct{}

var associativeOperators = map[int]void{
	ARTHEXPR_TADD: void{},
	ARTHEXPR_TMUL: void{},
}

type Expr struct {
	Ty  int
	Val any
}

type ConstExpr struct {
	Ty    int
	Val   uint64
	Ident string
}

// op type as used by the vm
func (e ArthExpr) GetVMOpType() int {
	switch e.Ty {
	case ARTHEXPR_TADD:
		return vm.OP_TADD
	case ARTHEXPR_TSUB:
		return vm.OP_TSUB
	case ARTHEXPR_TMUL:
		return vm.OP_TMUL
	case ARTHEXPR_TDIV:
		return vm.OP_TDIV
	default:
		return INVALID
	}
}

func IsConstexpr(e *Expr, ty int) bool {
	if e.Ty != EXPR_TCONST {
		return false
	}
	return e.Val.(ConstExpr).Ty == ty
}
func IsArthexpr(e *Expr, ty int) bool {
	if e.Ty != EXPR_TARTH {
		return false
	}
	return e.Val.(ArthExpr).Ty == ty
}
func AsArthExprTy(e *Expr) (int, bool) {
	if asArth, ok := e.Val.(ArthExpr); ok && e.Ty == EXPR_TARTH {
		return asArth.Ty, true
	} else {
		return INVALID, false
	}
}

func MakeConstexpr(c ConstExpr) *Expr {
	return &Expr{
		Ty:  EXPR_TCONST,
		Val: c,
	}
}
func MakeConstexprU64(v uint64) *Expr {
	return &Expr{
		Ty: EXPR_TCONST,
		Val: ConstExpr{
			Ty:  CONSTEXPR_TILIT,
			Val: v,
		},
	}
}
func MakeConstexprI64(v int64) *Expr {
	return &Expr{
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
func MakeConstexprIdent(v string) *Expr {
	return &Expr{
		Ty: EXPR_TCONST,
		Val: ConstExpr{
			Ty:    CONSTEXPR_TIDENT,
			Ident: v,
		},
	}
}
func MakeConstexprR(r int) *Expr {
	return &Expr{
		Ty: EXPR_TCONST,
		Val: ConstExpr{
			Ty:  CONSTEXPR_TREG,
			Val: uint64(r),
		},
	}
}
func MakeArth(a, b *Expr, ty int) *Expr {
	return &Expr{
		Ty: EXPR_TARTH,
		Val: ArthExpr{
			Ty: ty,
			A:  a,
			B:  b,
		},
	}
}
func MakeDeref(inner *Expr) *Expr {
	return &Expr{
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
	if a.Ty == CONSTEXPR_TFLIT || a.Ty == CONSTEXPR_TREG || a.Ty == CONSTEXPR_TIDENT {
		return a.Ty
	} else {
		return b.Ty
	}
}
func PerformOperation(a, b *ConstExpr, op int) (ConstExpr, bool) {
	switch op {
	case ARTHEXPR_TADD:
		return a.Add(b)
	case ARTHEXPR_TSUB:
		return a.Sub(b)
	case ARTHEXPR_TMUL:
		return a.Mul(b)
	case ARTHEXPR_TDIV:
		return a.Div(b)
	default:
		return ConstExpr{}, false
	}
}

func (a *ConstExpr) Add(b *ConstExpr) (ConstExpr, bool) {
	ty := opTy(a, b)
	if ty == CONSTEXPR_TREG || ty == CONSTEXPR_TIDENT {
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
	if ty == CONSTEXPR_TREG || ty == CONSTEXPR_TIDENT {
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
	if ty == CONSTEXPR_TREG || ty == CONSTEXPR_TIDENT {
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
	if ty == CONSTEXPR_TREG || ty == CONSTEXPR_TIDENT {
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
	A, B *Expr
}

// [r0]
// [r1+1]
// [bp+1]
// [bp+r0]
type DerefExpr struct {
	Inner *Expr
}

func TryEvaluatePruneExpression(e *Expr) (*Expr, bool) {
	ret1 := e
	var ret2 bool
	for pruned := true; pruned; {
		ret1, ret2 = TryEvaluateExpression(ret1)
		pruned = PruneExpression(ret1, nil, 0)
	}
	return ret1, ret2
}

func TryConstEvaluatePruneExpression(e *Expr) (ConstExpr, bool) {
	if expr, ok := TryEvaluatePruneExpression(e); ok && expr.Ty == EXPR_TCONST {
		return expr.Val.(ConstExpr), true
	}
	return ConstExpr{Ty: INVALID}, false
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
func TryEvaluateExpression(e *Expr) (*Expr, bool) {
	switch e.Ty {
	//if this is a constant expression, pass it on
	case EXPR_TCONST:
		// eAsConst := e.Val.(ConstExpr)
		// if eAsConst.Ty == CONSTEXPR_TREG {
		// 	return eAsConst, false
		// }
		return e, true
		//if this is an arthmetic expression, atttempt to evaluate it
	case EXPR_TARTH:
		eAsArth := e.Val.(ArthExpr)
		evalA, okA := TryEvaluateExpression(eAsArth.A)
		evalB, okB := TryEvaluateExpression(eAsArth.B)
		// in case both have been succesfully evaluated and both are const,
		// try to evaluate the arthmetic expression into a constant expression
		bothOk := okA && okB
		bothConst := evalA.Ty == EXPR_TCONST && evalB.Ty == EXPR_TCONST
		if bothOk && bothConst {
			valA := evalA.Val.(ConstExpr)
			valB := evalB.Val.(ConstExpr)
			res, ok := PerformOperation(&valA, &valB, eAsArth.Ty)
			if ok {
				return MakeConstexpr(res), true
			} else {
				return e, false
			}
			// in case only the left one is const and the right is an Arth expr
		} else if okB && IsArthexpr(evalA, ARTHEXPR_TADD) &&
			(IsConstexpr(evalB, CONSTEXPR_TILIT) || IsConstexpr(evalB, CONSTEXPR_TFLIT)) {
			leftArth := evalA.Val.(ArthExpr)
			rightConst := evalB.Val.(ConstExpr)
			if IsConstexpr(leftArth.A, CONSTEXPR_TILIT) || IsConstexpr(leftArth.A, CONSTEXPR_TFLIT) {
				leftArthA := leftArth.A.Val.(ConstExpr)
				res, ok := rightConst.Add(&leftArthA)
				if !ok {
					return e, false
				}
				return MakeArth(
					leftArth.B,
					MakeConstexpr(res),
					ARTHEXPR_TADD,
				), true
			} else if IsConstexpr(leftArth.B, CONSTEXPR_TILIT) || IsConstexpr(leftArth.B, CONSTEXPR_TFLIT) {
				leftArthB := leftArth.B.Val.(ConstExpr)
				res, ok := rightConst.Add(&leftArthB)
				if !ok {
					return e, false
				}
				return MakeArth(
					MakeConstexpr(res),
					leftArth.A,
					ARTHEXPR_TADD,
				), true
			}
		} else if okA && IsArthexpr(evalB, ARTHEXPR_TADD) &&
			(IsConstexpr(evalA, CONSTEXPR_TILIT) || IsConstexpr(evalA, CONSTEXPR_TFLIT)) {
			rightArth := evalB.Val.(ArthExpr)
			leftConst := evalA.Val.(ConstExpr)
			if IsConstexpr(rightArth.A, CONSTEXPR_TILIT) || IsConstexpr(rightArth.A, CONSTEXPR_TFLIT) {
				rightArthA := rightArth.A.Val.(ConstExpr)
				res, ok := leftConst.Add(&rightArthA)
				if !ok {
					return e, false
				}
				return MakeArth(
					MakeConstexpr(res),
					rightArth.B,
					ARTHEXPR_TADD,
				), true
			} else if IsConstexpr(rightArth.B, CONSTEXPR_TILIT) || IsConstexpr(rightArth.B, CONSTEXPR_TFLIT) {
				rightArthB := rightArth.B.Val.(ConstExpr)
				res, ok := leftConst.Add(&rightArthB)
				if !ok {
					return e, false
				}
				return MakeArth(
					rightArth.A,
					MakeConstexpr(res),
					ARTHEXPR_TADD,
				), true
			}
		}
		// if all else fails, just pass the best effort attempt with whatever could've been evaluated
		return MakeArth(evalA, evalB, eAsArth.Ty), false
	case EXPR_TDEREF:
		inner := e.Val.(DerefExpr).Inner
		inner, ok := TryEvaluateExpression(inner)
		return MakeDeref(inner), ok
	default:
		return e, false
	}
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
func PruneExpression(e, swap *Expr, pruneWith int) bool {
	switch pruneWith {
	case 0:
		if arthTy, ok := AsArthExprTy(e); ok {
			if _, isAssoc := associativeOperators[arthTy]; isAssoc {
				pruneWith = arthTy
			} else {
				return false
			}
		}
	default:
		if _, isAssoc := associativeOperators[pruneWith]; !isAssoc {
			return false
		}
	}
	switch {
	// if this is an associative operator expression
	case IsArthexpr(e, pruneWith):
		arthE := e.Val.(ArthExpr)
		switch {
		// if it has one node as a literal node and the other as a different associative expr of the same op
		case (IsConstexpr(arthE.A, CONSTEXPR_TILIT) || IsConstexpr(arthE.A, CONSTEXPR_TFLIT)) &&
			IsArthexpr(arthE.B, pruneWith):
			//then recurse into it with the other literal as swap
			if swap != nil {
				return PruneExpression(arthE.B, swap, pruneWith)
			} else {
				return PruneExpression(arthE.B, arthE.A, pruneWith)
			}
		//same case but branches are flipped
		case (IsConstexpr(arthE.B, CONSTEXPR_TILIT) || IsConstexpr(arthE.B, CONSTEXPR_TFLIT)) &&
			IsArthexpr(arthE.A, pruneWith):
			if swap != nil {
				return PruneExpression(arthE.A, swap, pruneWith)
			} else {
				return PruneExpression(arthE.A, arthE.B, pruneWith)
			}
		// we've hit the bottom and the other guy is not a comp-time evaluable expression, and swap is not nil
		case (IsConstexpr(arthE.A, CONSTEXPR_TILIT) || IsConstexpr(arthE.A, CONSTEXPR_TFLIT)) &&
			(!IsConstexpr(arthE.B, CONSTEXPR_TILIT) && !IsConstexpr(arthE.B, CONSTEXPR_TFLIT)) &&
			swap != nil:
			*arthE.B, *swap = *swap, *arthE.B
			return true
		// the same but in reverse
		case (IsConstexpr(arthE.B, CONSTEXPR_TILIT) || IsConstexpr(arthE.B, CONSTEXPR_TFLIT)) &&
			(!IsConstexpr(arthE.A, CONSTEXPR_TILIT) && !IsConstexpr(arthE.A, CONSTEXPR_TFLIT)) &&
			swap != nil:
			*arthE.A, *swap = *swap, *arthE.A
			return true
		// if this is a nested expression of the same op then recurse into it
		case IsArthexpr(arthE.A, pruneWith):
			return PruneExpression(arthE.A, swap, pruneWith)
		case IsArthexpr(arthE.B, pruneWith):
			return PruneExpression(arthE.B, swap, pruneWith)
		}
	}
	return false
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
