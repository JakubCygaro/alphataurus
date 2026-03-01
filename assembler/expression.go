package assembler

const (
	EXPR_TCONST = iota
	EXPR_TARTH
	EXPR_TDEREF
)
const (
	ARTHEXPR_TADD = iota
	ARTHEXPR_TSUB
	ARTHEXPR_TMUL
	ARTHEXPR_TDIV
)
const (
	// r0
	// sp
	CONSTEXPR_TREG = iota
	// 1.23
	// 1313
	CONSTEXPR_TLIT
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
	A, B  *ArthExpr
	Const *ConstExpr
}

// [r0]
// [r1+1]
// [bp+1]
// [bp+r0]
type DerefExpr struct {
	Inner ArthExpr
}
