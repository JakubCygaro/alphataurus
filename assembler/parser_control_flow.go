package assembler

import (
	"fmt"

	"github.com/JakubCygaro/alphataurus/assembler/errors"
	"github.com/JakubCygaro/alphataurus/internal/vm"
)

func (p *Parser) parseCmp() error {

	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	op1 := p.lexer.CurrentToken()
	ty := vm.TY_INT64
	switch op1.Ty {
	case TOKEN_TFLOAT:
		ty = vm.TY_FLOAT64
	default:
		p.lexer.UnreadToken()
	}
	if expr, err := p.parseExpression(0); err != nil {
		return err
	} else if eval, _ := TryConstEvaluateExpression(expr); eval.Ty == CONSTEXPR_TREG {
		op1.Ty = TOKEN_TREG
		op1.val = int(eval.Val)
	} else {
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"First operand to instruction must be a valid register",
			p.lexer.line, p.lexer.col)
	}
	if op1.val.(int) > vm.GP_REG_MAX {
		return fmt.Errorf("Disallowed minuend register %s", p.lexer.CurrentPosition())
	}
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	if p.lexer.CurrentToken().Ty != TOKEN_TCOMMA {
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"Instruction missing a comma",
			p.lexer.line, p.lexer.col)
	}
	// if err := p.lexer.ReadNextToken(); err != nil {
	// 	return err
	// }
	// op2 := p.lexer.CurrentToken()
	op2 := Token{Ty: INVALID}
	if expr, err := p.parseExpression(0); err != nil {
		return err
	} else if eval, ok := TryConstEvaluateExpression(expr); eval.Ty == CONSTEXPR_TREG {
		op2.Ty = TOKEN_TREG
		op2.val = int(eval.Val)
	} else if !ok {
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"Second operand to instruction must be a valid register or a constant expression",
			p.lexer.line, p.lexer.col)
	} else if eval.Ty == CONSTEXPR_TILIT {
		op2.Ty = TOKEN_TINTEGER_LIT
		op2.val = eval.Val
	} else if eval.Ty == CONSTEXPR_TFLIT {
		op2.Ty = TOKEN_TFLOAT_LIT
		op2.val = eval.Val
	} else {
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"Bad expression",
			p.lexer.line, p.lexer.col)
	}
	switch op2.Ty {
	case TOKEN_TREG:
		if op2.val.(int) > vm.GP_REG_MAX {
			return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
				"Disallowed subtrahend register",
				p.lexer.line, p.lexer.col)
		}
		p.currentInst = Instruction{
			Ty: INST_TCMPRR,
			Data: InstCmpData{
				Ty:  ty,
				Min: op1.val.(int),
				Sub: op2.val.(int),
			},
		}
	case TOKEN_TINTEGER_LIT:
		p.currentInst = Instruction{
			Ty: INST_TCMPIR,
			Data: InstCmpData{
				Ty:  ty,
				Min: op1.val.(int),
				Imm: op2.val.(uint64),
			},
		}
	case TOKEN_TFLOAT_LIT:
		p.currentInst = Instruction{
			Ty: INST_TCMPIR,
			Data: InstCmpData{
				Ty:  ty,
				Min: op1.val.(int),
				Imm: op2.val.(uint64),
			},
		}
	}
	return nil
}
func (p *Parser) parseJmp(ty int) error {
	inst := Instruction{
		Ty: ty,
	}
	addr := Token{Ty: INVALID}
	if expr, err := p.parseExpression(0); err != nil {
		return err
	} else if eval, ok := TryConstEvaluateExpression(expr); eval.Ty == CONSTEXPR_TIDENT {
		addr.Ty = TOKEN_TIDENT
		addr.val = eval.Ident
	} else if !ok {
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"Second operand to instruction must be a valid label or a an address",
			p.lexer.line, p.lexer.col)
	} else if eval.Ty == CONSTEXPR_TILIT {
		addr.Ty = TOKEN_TINTEGER_LIT
		addr.val = eval.Val
	} else {
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"Bad expression",
			p.lexer.line, p.lexer.col)
	}
	switch addr.Ty {
	case TOKEN_TINTEGER_LIT:
		inst.Data = InstJmpData{
			Address: addr.val.(uint64),
		}
	case TOKEN_TIDENT:
		inst.Data = InstJmpData{
			Address: addr.val.(string),
		}
	default:
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"instruction requires a valid address or label as a parameter",
			p.lexer.line, p.lexer.col)
	}
	p.currentInst = inst
	return nil
}
func (p *Parser) parseCall() error {
	ok := false
	var pruned ConstExpr
	if param, err := p.parseExpression(0); err != nil {
		return err
	} else if pruned, ok = TryConstEvaluatePruneExpression(param); !ok {
		return errors.FailedToParse("call instruction",
			"Parameter of call instruction must be an address literal or label", p.lexer.line, p.lexer.col)
	}
	switch pruned.Ty {
	case CONSTEXPR_TILIT:
		p.currentInst = Instruction{
			Ty: INST_TCALL,
			Data: InstCallData{
				Addr: pruned.Val,
				Expr: nil,
			},
		}
	case CONSTEXPR_TIDENT:
		p.currentInst = Instruction{
			Ty: INST_TCALL,
			Data: InstCallData{
				Addr:  pruned.Val,
				Ident: pruned.Ident,
				Expr:  nil,
			},
		}
	default:
		return errors.FailedToParse("call instruction",
			"Parameter of call instruction must be an address literal or label", p.lexer.line, p.lexer.col)
	}
	return nil
}
