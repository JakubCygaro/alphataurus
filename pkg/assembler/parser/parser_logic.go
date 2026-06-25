package assembler

import (

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func (p *Parser) parseLogical(logTy int) error {
	var op1 lx.Token
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
	if !vm.IsLogRAllowed(byte(op1.Val.(lx.RegisterData).Reg)) {
		return errors.FailedToParse(p.currentIdent,
			p.currentStartToken.Line, p.currentStartToken.Col,
			"Disallowed first operand register `%s`",
			op1.ForceValAsString(),
		)
	}

	if logTy == LOG_TNOT {
		p.currentInst = Instruction{
			Data: InstNot{
				First: op1.Val.(lx.RegisterData),
			},
		}
		return nil
	}

	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	comma := p.lexer.CurrentToken()

	if comma.Ty != lx.TOKEN_TCOMMA {
		return errors.FailedToParse(p.currentIdent,
			comma.Line, comma.Col,
			"Instruction missing a comma, got `%s`",
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
				"Disallowed second operand register, got `%s`",
				op2.UnpackAsRegisterData().String(),
			)
		}
		var ty LogTy
		switch logTy {
		case LOG_TOR:
			ty = OR
		case LOG_TAND:
			ty = AND
		case LOG_TXOR:
			ty = XOR
		case LOG_TLSH:
			ty = LSH
		case LOG_TRSH:
			ty = RSH
		}
		p.currentInst = Instruction{
			Data: InstLogicalRR{
				First:  op1.Val.(lx.RegisterData),
				Second: op2.UnpackAsRegisterData(),
				LogTy: ty,
			},
		}
	case CONSTEXPR_TILIT:
		var ty LogTy
		switch logTy {
		case LOG_TOR:
			ty = OR
		case LOG_TAND:
			ty = AND
		case LOG_TXOR:
			ty = XOR
		case LOG_TLSH:
			ty = LSH
		case LOG_TRSH:
			ty = RSH
		}
		p.currentInst = Instruction{
			Data: InstLogicalIR{
				First: op1.Val.(lx.RegisterData),
				Imm:   op2.Val,
				LogTy: ty,
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
