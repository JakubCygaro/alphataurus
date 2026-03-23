package assembler

import (
	"fmt"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func (p *Parser) parseLogical(logTy int) error {
	var op1 Token
	if expr, err := p.parseExpression(0); err != nil {
		return err
	} else if eval, _ := TryEvaluateExpression(expr); eval.Ty != CONSTEXPR_TREG {
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"First operand to instruction must be a valid register",
			p.lexer.line, p.lexer.col)
	} else {
		op1.Val = int(expr.Val.(ConstExpr).Val)
	}
	if !vm.IsGpReg(byte(op1.Val.(int))){
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"Disallowed first operand register",
			p.lexer.line, p.lexer.col)
	}

	if logTy == LOG_TNOT {
		p.currentInst = Instruction{
			Ty: INST_TNOT,
			Data: InstLogicalData{
				First:  int(op1.Val.(int)),
			},
		}
		return nil
	}

	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	comma := p.lexer.CurrentToken()

	if comma.Ty != TOKEN_TCOMMA {
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"Instruction missing a comma",
			p.lexer.line, p.lexer.col)
	}
	var op2 ConstExpr
	if expr, err := p.parseExpression(0); err != nil {
		return err
	} else {
		eval, ok := TryConstEvaluateExpression(expr)
		if eval.Ty != CONSTEXPR_TREG && !ok {
			return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
				"Second operand to instruction has to be a valid register or a compile time expression",
				p.lexer.line, p.lexer.col)
		}
		op2 = eval
	}
	switch op2.Ty {
	case CONSTEXPR_TREG:
		if !vm.IsGpReg(byte(op2.Val)){
			return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"Disallowed second operand register",
			p.lexer.line, p.lexer.col)
		}
		var ty int
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
				First: op1.Val.(int),
				Second:  int(op2.Val),
			},
		}
	case CONSTEXPR_TILIT:
		var ty int
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
				First: op1.Val.(int),
				Imm:  op2.Val,
			},
		}
	default:
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"Second operand to instruction has to be a valid register or a compile time expression",
			p.lexer.line, p.lexer.col)
	}
	return nil
}
