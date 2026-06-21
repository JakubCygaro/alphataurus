package assembler

import (

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func (p *Parser) parseLogical(logTy int) error {
	var op1 Token
	if expr, err := p.parseExpression(0); err != nil {
		return err
	} else if eval, _ := TryEvaluateExpression(expr); eval.Ty != CONSTEXPR_TREG {
		em, _ := expr.Emit()
		return errors.FailedToParse(p.currentIdent,
			p.currentStartToken.Line, p.currentStartToken.Col,
			"First operand to instruction must be a valid register.\n"+
				"In expression `%s`",
			em,
		)
	} else {
		op1.Val = expr.Val.(ConstExpr).UnpackAsRegisterData()
	}
	if !vm.IsLogRAllowed(byte(op1.Val.(RegisterData).Reg)) {
		return errors.FailedToParse(p.currentIdent,
			p.currentStartToken.Line, p.currentStartToken.Col,
			"Disallowed first operand register `%s`"+
				op1.ForceValAsString(),
		)
	}

	if logTy == LOG_TNOT {
		p.currentInst = Instruction{
			Ty: INST_TNOT,
			Data: InstLogicalData{
				First: op1.Val.(RegisterData),
			},
		}
		return nil
	}

	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	comma := p.lexer.CurrentToken()

	if comma.Ty != TOKEN_TCOMMA {
		return errors.FailedToParse(p.currentIdent,
			comma.Line, comma.Col,
			"Instruction missing a comma, got `%s`"+
				comma.ForceValAsString(),
		)
	}
	var op2 ConstExpr
	if expr, err := p.parseExpression(0); err != nil {
		return err
	} else {
		eval, ok := TryConstEvaluateExpression(expr)
		if eval.Ty != CONSTEXPR_TREG && !ok {
			em, _ := expr.Emit()
			return errors.FailedToParse(p.currentIdent,
				expr.Line, expr.Col,
				"Second operand to instruction has to be a "+
					"valid register or a compile time expression\n"+
					"In expression `%s`",
				em,
			)
		}
		op2 = eval
	}
	switch op2.Ty {
	case CONSTEXPR_TREG:
		if !vm.IsMovFromRAllowed(byte(op2.Val)) {
			return errors.FailedToParse(p.currentIdent,
				p.currentStartToken.Line, p.currentStartToken.Col,
				"Disallowed second operand register, got `%s`"+
					op2.UnpackAsRegisterData().String(),
			)
		}
		var ty InstTy
		switch logTy {
		case LOG_TOR:
			ty = INST_TORRR
		case LOG_TAND:
			ty = INST_TANDRR
		case LOG_TXOR:
			ty = INST_TXORRR
		case LOG_TLSH:
			ty = INST_TLSHRR
		case LOG_TRSH:
			ty = INST_TRSHRR
		}
		p.currentInst = Instruction{
			Ty: ty,
			Data: InstLogicalData{
				First:  op1.Val.(RegisterData),
				Second: op2.UnpackAsRegisterData(),
			},
		}
	case CONSTEXPR_TILIT:
		var ty InstTy
		switch logTy {
		case LOG_TOR:
			ty = INST_TORIR
		case LOG_TAND:
			ty = INST_TANDIR
		case LOG_TXOR:
			ty = INST_TXORIR
		case LOG_TLSH:
			ty = INST_TLSHIR
		case LOG_TRSH:
			ty = INST_TRSHIR
		}
		p.currentInst = Instruction{
			Ty: ty,
			Data: InstLogicalData{
				First: op1.Val.(RegisterData),
				Imm:   op2.Val,
			},
		}
	default:
		em, _ := op2.Emit()
		return errors.FailedToParse(p.currentIdent,
			p.currentStartToken.Line, p.currentStartToken.Col,
			"Second operand to instruction has to be a valid "+
			"register or a compile time expression."+
			"In expression: `%s`",
			em,
		)
	}
	return nil
}
