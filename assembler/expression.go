package assembler

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
	// r0
	// sp
	CONSTEXPR_TREG = iota
	// 1313
	CONSTEXPR_TILIT
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

type ArthExpr struct {
	Ty    int
	A, B  *Expr
}

// [r0]
// [r1+1]
// [bp+1]
// [bp+r0]
type DerefExpr struct {
	Inner ArthExpr
}
