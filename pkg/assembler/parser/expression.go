package assembler

import (
	"fmt"
	"math"
	"strings"

	"github.com/JakubCygaro/alphataurus/internal/pkg/opt"
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
)

const (
	EXPR_TCONST = iota
	EXPR_TARTH
	EXPR_TDEREF
)
const (
	ARTHEXPR_TADD = lx.TOKEN_TPLUS
	ARTHEXPR_TSUB = lx.TOKEN_TMINUS
	ARTHEXPR_TMUL = lx.TOKEN_TASTERISK
	ARTHEXPR_TDIV = lx.TOKEN_TSLASH
)
const (
	INVALID = -1
	PENDING = -1
	// r0
	// sp
	// CONSTEXPR_TREG stores its register data in the bits of the uint64 Val field
	//
	// As such it needs to be extracted to be usable
	CONSTEXPR_TREG = 0
	// 1313
	CONSTEXPR_TILIT = iota
	// 1.23
	CONSTEXPR_TFLIT
	// abcd
	CONSTEXPR_TIDENT
)

type void struct{}

type Expr struct {
	Val       any
	Line, Col int
}
type MemExpr struct {
	Base   lx.RegisterData
	Index  opt.Opt[lx.RegisterData]
	ScaleF *Expr
	Disp   *Expr
	RegOp  int
	DispOp int
}

type ConstExpr struct {
	Val any
}
type RegExpr struct {
	Reg lx.RegisterData
}

type Scale byte

const (
	SCALE_0 Scale = iota
	SCALE_2
	SCALE_4
	SCALE_8
)

type ConstExprILit struct {
	Integer uint64
}

func (c ConstExprILit) Signed() int64 {
	return int64(c.Integer)
}

type ConstExprFLit struct {
	Float uint64
}

func (c ConstExprFLit) ToFloat() float64 {
	return math.Float64frombits(c.Float)
}

type ExprIden struct {
	Ident string
}
type NegExpr struct {
	Inner *Expr
}

type ArthExpr struct {
	Val any
}
type ArthExprAdd struct {
	A, B *Expr
}
type ArthExprSub struct {
	A, B *Expr
}
type ArthExprMul struct {
	A, B *Expr
}
type ArthExprDiv struct {
	A, B *Expr
}

// [r0]
// [r1+1]
// [bp+1]
// [bp+r0]
type DerefExpr struct {
	Inner *Expr
}

// op type as used by the vm
// func (e ArthExpr) GetVMOpType() int {
// 	switch e.Val.(type) {
// 	case ArthExprAdd:
// 		return vm.OP_TADD
// 	case ArthExprSub:
// 		return vm.OP_TSUB
// 	case ArthExprMul:
// 		return vm.OP_TMUL
// 	case ArthExprDiv:
// 		return vm.OP_TDIV
// 	default:
// 		return INVALID
// 	}
// }

func (e *Expr) IsConstexpr() bool {
	_, ok := e.Val.(ConstExpr)
	return ok
}
func (e *Expr) IsArthexpr() bool {
	_, ok := e.Val.(ArthExpr)
	return ok
}
func (e *Expr) IsRegexpr() bool {
	_, ok := e.Val.(RegExpr)
	return ok
}
func (e *Expr) IsMemexpr() bool {
	_, ok := e.Val.(MemExpr)
	return ok
}
func IsExpr[E any](e *Expr) bool {
	_, ok := e.Val.(E)
	return ok
}

type CExpr interface {
	ConstExprILit | ConstExprFLit | ExprIden
}

func IsConstexprType[CT CExpr](e *Expr) (CT, bool) {
	var def CT
	if c, ok := e.Val.(ConstExpr); !ok {
		return def, false
	} else {
		r, ok := c.Val.(CT)
		return r, ok
	}
}

type AExpr interface {
	ArthExprAdd | ArthExprSub | ArthExprMul | ArthExprDiv
}

func IsArthexprType[AT AExpr](e *Expr) bool {
	if a, ok := e.Val.(ArthExpr); !ok {
		return false
	} else {
		_, ok := a.Val.(AT)
		return ok
	}
}
func AsArthExprTy[
	AT ArthExprAdd | ArthExprSub | ArthExprMul | ArthExprDiv](e *Expr) (AT, bool) {
	if a, ok := e.Val.(ArthExpr); !ok {
		return AT{}, false
	} else {
		at, ok := a.Val.(AT)
		return at, ok
	}
}

func MakeConstexpr(c ConstExpr) *Expr {
	return &Expr{
		Val: c,
	}
}
func MakeConstexprU64(v uint64) *Expr {
	return &Expr{
		Val: ConstExpr{
			Val: ConstExprILit{
				Integer: v,
			},
		},
	}
}
func MakeConstexprI64(v int64) *Expr {
	return &Expr{
		Val: ConstExpr{
			Val: ConstExprILit{
				Integer: uint64(v),
			},
		},
	}
}
func MakeConstexprAny(v any) (*Expr, error) {
	var cexpr any
	switch v := v.(type) {
	case int64:
		cexpr = ConstExprILit{Integer: uint64(v)}
	case uint64:
		cexpr = ConstExprILit{Integer: v}
	case float64:
		cexpr = ConstExprFLit{Float: math.Float64bits(v)}
	case string:
		cexpr = ExprIden{Ident: v}
	default:
		return nil, fmt.
			Errorf("Unsupported constexpr value type `%T`.", v)
	}
	return &Expr{
		Val: ConstExpr{
			Val: cexpr,
		},
	}, nil
}
func MakeConstexprF64Bits(bits uint64) *Expr {
	return &Expr{
		Val: ConstExpr{
			Val: ConstExprFLit{
				Float: uint64(bits),
			},
		},
	}
}
func MakeConstexprF64(v float64) *Expr {
	return &Expr{
		Val: ConstExpr{
			Val: ConstExprFLit{
				Float: math.Float64bits(v),
			},
		},
	}
}
func MakeConstexprIdent(v string) *Expr {
	return &Expr{
		Val: ConstExpr{
			Val: ExprIden{
				Ident: v,
			},
		},
	}
}
func MakeRegexpr(r, size byte) *Expr {
	return &Expr{
		Val: RegExpr{
			Reg: lx.RegisterData{Reg: int(r), Size: size},
		},
	}
}
func MakeArth[
	AT ArthExprAdd | ArthExprSub | ArthExprDiv | ArthExprMul](aex AT) *Expr {
	return &Expr{
		Val: ArthExpr{
			Val: aex,
		},
	}
}
func MakeDeref(inner *Expr) *Expr {
	return &Expr{
		Val: DerefExpr{
			Inner: inner,
		},
	}
}
func MakeNeg(inner *Expr) *Expr {
	return &Expr{
		Val: NegExpr{
			Inner: inner,
		},
	}
}
func (e *ConstExpr) AsFloat() float64 {
	switch c := e.Val.(type) {
	case ConstExprFLit:
		return math.Float64frombits(c.Float)
	case ConstExprILit:
		return float64(c.Integer)
	default:
		return math.NaN()
	}
}

// CONSTEXPR_TREG stores its register data in the bits of the uint64 Val field
//
// As such it needs to be extracted to be usable
func (c ConstExpr) UnpackAsRegisterData() (lx.RegisterData, bool) {
	if r, ok := c.Val.(RegExpr); ok {
		return r.Reg, true
	} else {
		return lx.RegisterData{}, false
	}
}
func UnpackRegisterData(packed uint64) lx.RegisterData {
	return lx.RegisterData{
		Reg:  int(byte(packed)),
		Size: byte(packed >> 8),
	}
}
func (e *ArthExpr) Emit() (string, error) {
	emitEach := func(ae, be *Expr) (a, b string, err error) {
		if a, err := ae.Emit(); err != nil {
			return "", "", err
		} else if b, err := be.Emit(); err != nil {
			return "", "", err
		} else {
			return a, b, nil
		}
	}
	switch a := e.Val.(type) {
	case ArthExprAdd:
		if a, b, err := emitEach(a.A, a.B); err != nil {
			return "", err
		} else {
			return fmt.Sprintf("%s + %s", a, b), nil
		}
	case ArthExprSub:
		if a, b, err := emitEach(a.A, a.B); err != nil {
			return "", err
		} else {
			return fmt.Sprintf("%s - %s", a, b), nil
		}
	case ArthExprMul:
		if a, b, err := emitEach(a.A, a.B); err != nil {
			return "", err
		} else {
			return fmt.Sprintf("%s * %s", a, b), nil
		}
	case ArthExprDiv:
		if a, b, err := emitEach(a.A, a.B); err != nil {
			return "", err
		} else {
			return fmt.Sprintf("%s / %s", a, b), nil
		}
	default:
		return "?", fmt.Errorf("Invalid arthmetic expression type")
	}
}
func (e *ConstExpr) Emit() string {
	switch c := e.Val.(type) {
	case ConstExprILit:
		return fmt.Sprintf("%v", c.Integer)
	case ConstExprFLit:
		return fmt.Sprintf("%v", math.Float64frombits(c.Float))
	case ExprIden:
		return c.Ident
	}
	return ""
}
func (e *Expr) Emit() (string, error) {
	switch v := e.Val.(type) {
	case NegExpr:
		emit, err := v.Inner.Emit()
		return fmt.Sprintf("- %s", emit), err
	case RegExpr:
		return v.Reg.String(), nil
	case ConstExpr:
		return v.Emit(), nil
	case DerefExpr:
		emit, err := v.Inner.Emit()
		return fmt.Sprintf("[%s]", emit), err
	case ArthExpr:
		if inner, err := v.Emit(); err != nil {
			return "", err
		} else {
			return fmt.Sprintf("(%s)", inner), nil
		}
	case MemExpr:
		em := [5]string{}
		em[0] = v.Base.String()
		switch v.RegOp {
		case lx.TOKEN_TPLUS:
			em[1] = "+"
		case lx.TOKEN_TMINUS:
			em[1] = "-"
		default:
			em[1] = "?"
		}
		if i, ok := v.Index.TryGet(); ok {
			em[2] = i.String()
		}
		switch v.DispOp {
		case lx.TOKEN_TPLUS:
			em[3] = "+"
		case lx.TOKEN_TMINUS:
			em[3] = "-"
		default:
			em[3] = "?"
		}
		if v.Disp == nil {
			em[3] = ""
		} else if oem, err := v.Disp.Emit(); err != nil {
			return "", err
		} else {
			em[4] = oem
		}
		return strings.
				Join(
					em[:],
					" ",
				),
			nil

	default:
		return "", fmt.Errorf("<UNEMITABLE EXPRESSION>")
	}
}
