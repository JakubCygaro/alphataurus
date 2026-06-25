package assembler

import (
	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

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
	OffsetExpr *Expr
}

func (p *Parser) processDerefNestedArth(arthExpr ArthExpr, nestLvl int) (DerefData, error) {
	ret := DerefData{
		Reg1:       lx.GetInvalidRegister(),
		Reg2:       lx.GetInvalidRegister(),
		OffsetOp:   INVALID,
		Offset:     INVALID,
		OffsetExpr: nil,
	}
	switch {
	case IsConstexpr(arthExpr.A, CONSTEXPR_TREG) &&
		IsConstexpr(arthExpr.B, CONSTEXPR_TILIT):

		ret.Ty = DEREF_T1RO
		ret.Reg1 = arthExpr.A.Val.(ConstExpr).UnpackAsRegisterData()
		ret.Offset = int64(arthExpr.B.Val.(ConstExpr).Val)
		ret.OffsetOp = arthExpr.GetVMOpType()
	case IsConstexpr(arthExpr.A, CONSTEXPR_TILIT) &&
		IsConstexpr(arthExpr.B, CONSTEXPR_TREG) &&
		(arthExpr.Ty == ARTHEXPR_TADD):

		ret.Ty = DEREF_T1RO
		ret.Reg1 = arthExpr.B.Val.(ConstExpr).UnpackAsRegisterData()
		ret.Offset = int64(arthExpr.A.Val.(ConstExpr).Val)
		ret.OffsetOp = vm.OP_TADD
	case IsConstexpr(arthExpr.A, CONSTEXPR_TREG) &&
		IsConstexpr(arthExpr.B, CONSTEXPR_TREG) &&
		(arthExpr.Ty == ARTHEXPR_TADD):

		ret.Ty = DEREF_T2RO
		ret.Reg1 = arthExpr.A.Val.(ConstExpr).UnpackAsRegisterData()
		ret.Reg2 = arthExpr.B.Val.(ConstExpr).UnpackAsRegisterData()
		ret.Offset = int64(0)
		ret.OffsetOp = vm.OP_TADD
	case IsConstexpr(arthExpr.A, CONSTEXPR_TREG) &&
		IsArthexpr(arthExpr.B, ARTHEXPR_TADD) &&
		nestLvl == 0:

		nestedD, err := p.processDerefNestedArth(arthExpr.B.Val.(ArthExpr), nestLvl+1)
		if err != nil {
			return ret, err
		}
		if nestedD.OffsetOp != vm.OP_TADD && nestedD.OffsetOp != vm.OP_TSUB {
			em, _ := arthExpr.B.Emit()
			return ret, errors.FailedToParse("dereference expression",
				p.currentStartToken.Line, p.currentStartToken.Col,
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
	case IsConstexpr(arthExpr.B, CONSTEXPR_TREG) &&
		IsArthexpr(arthExpr.A, ARTHEXPR_TADD) &&
		nestLvl == 0 &&
		(arthExpr.Ty == ARTHEXPR_TADD):

		nestedD, err := p.processDerefNestedArth(arthExpr.A.Val.(ArthExpr), nestLvl+1)
		if err != nil {
			return ret, err
		}
		ret.Ty = DEREF_T2RO
		ret.Reg1 = arthExpr.B.Val.(ConstExpr).UnpackAsRegisterData()
		ret.Reg2 = nestedD.Reg1
		ret.Offset = nestedD.Offset
		ret.OffsetOp = nestedD.OffsetOp
	case IsConstexpr(arthExpr.A, CONSTEXPR_TILIT) &&
		IsArthexpr(arthExpr.B, ARTHEXPR_TADD) &&
		nestLvl == 0 &&
		(arthExpr.Ty == ARTHEXPR_TADD):

		nestedD, err := p.processDerefNestedArth(arthExpr.B.Val.(ArthExpr), nestLvl+1)
		if err != nil {
			return ret, err
		}
		if nestedD.OffsetOp != vm.OP_TADD || nestedD.Ty != DEREF_T2RO {
			em, _ := arthExpr.B.Emit()
			return ret, errors.FailedToParse("dereference expression",
				p.currentStartToken.Line, p.currentStartToken.Col,
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
	case IsConstexpr(arthExpr.B, CONSTEXPR_TILIT) &&
		IsArthexpr(arthExpr.A, ARTHEXPR_TADD) &&
		nestLvl == 0:

		nestedD, err := p.processDerefNestedArth(arthExpr.A.Val.(ArthExpr), nestLvl+1)
		if err != nil {
			return ret, err
		}
		if nestedD.OffsetOp != vm.OP_TADD || nestedD.Ty != DEREF_T2RO {
			em, _ := arthExpr.B.Emit()
			return ret, errors.FailedToParse("dereference expression",
				p.currentStartToken.Line, p.currentStartToken.Col,
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
			p.currentStartToken.Line, p.currentStartToken.Col,
			"Invalid dereference expression `%s`",
			em,
		)
		// TODO: label dereference support
	}
	return ret, nil
}

func (p *Parser) processDeref(inner *Expr) (DerefData, error) {
	ret := DerefData{
		Reg1:       lx.GetInvalidRegister(),
		Reg2:       lx.GetInvalidRegister(),
		OffsetOp:   INVALID,
		Offset:     INVALID,
		OffsetExpr: nil,
	}
	switch inner.Ty {
	case EXPR_TCONST:
		innerConst := inner.Val.(ConstExpr)
		switch innerConst.Ty {
		case CONSTEXPR_TILIT:

			ret.Ty = DEREF_T0RO
			ret.Offset = int64(innerConst.Val)
		case CONSTEXPR_TREG:

			ret.Ty = DEREF_T1RO
			regD := innerConst.UnpackAsRegisterData()
			ret.Reg1 = regD
			ret.Offset = int64(0)
			ret.OffsetOp = vm.OP_TADD
		// TODO: label dereference support
		default:
			em, _ := inner.Emit()
			return ret, errors.FailedToParse("dereference expression",
				p.currentStartToken.Line, p.currentStartToken.Col,
				"Invalid single parameter dereference expression."+
				"In Expression `%s`",
				em,
			)
		}
	case EXPR_TARTH:
		arthExpr := inner.Val.(ArthExpr)
		return p.processDerefNestedArth(arthExpr, 0)
	default:
		em, _ := inner.Emit()
		return ret, errors.FailedToParse("dereference expression",
			p.currentStartToken.Line, p.currentStartToken.Col,
			"Invalid dereference expression."+
			"In Expression `%s`",
			em,
		)
	}
	return ret, nil
}
