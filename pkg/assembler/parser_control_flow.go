package assembler

import (
	"fmt"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func (p *Parser) parseCmp() error {

	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	op1 := p.lexer.CurrentToken()
	ty := vm.TY_UINT
	switch op1.Ty {
	case TOKEN_TFLOAT:
		ty = vm.TY_FLOAT
	case TOKEN_TSIGNED:
		ty = vm.TY_SINT
	case TOKEN_TUNSIGNED:
		ty = vm.TY_UINT
	default:
		p.lexer.UnreadToken()
	}
	if expr, err := p.parseExpression(0); err != nil {
		return err
	} else if eval, _ := TryConstEvaluateExpression(expr); eval.Ty == CONSTEXPR_TREG {
		op1.Ty = TOKEN_TREG
		op1.Val = eval.UnpackAsRegisterData()
	} else {
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"First operand to instruction must be a valid register",
			p.lexer.line, p.lexer.col)
	}
	if op1.Val.(RegisterData).Reg > vm.MAX_REG_IDX {
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"Disallowed minuend register",
			p.lexer.line, p.lexer.col)
	}
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	if p.lexer.CurrentToken().Ty != TOKEN_TCOMMA {
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"Instruction missing a comma",
			p.lexer.line, p.lexer.col)
	}

	op2 := Token{Ty: INVALID}
	if expr, err := p.parseExpression(0); err != nil {
		return err
	} else if eval, ok := TryConstEvaluateExpression(expr); eval.Ty == CONSTEXPR_TREG {
		op2.Ty = TOKEN_TREG
		op2.Val = eval.UnpackAsRegisterData()
	} else if !ok {
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"Second operand to instruction must be a valid register or a constant expression",
			p.lexer.line, p.lexer.col)
	} else if eval.Ty == CONSTEXPR_TILIT {
		op2.Ty = TOKEN_TINTEGER_LIT
		op2.Val = eval.Val
	} else if eval.Ty == CONSTEXPR_TFLIT {
		op2.Ty = TOKEN_TFLOAT_LIT
		op2.Val = eval.Val
	} else {
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"Bad expression",
			p.lexer.line, p.lexer.col)
	}
	switch op2.Ty {
	case TOKEN_TREG:
		if op2.Val.(RegisterData).Reg > vm.GP_REG_MAX {
			return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
				"Disallowed subtrahend register",
				p.lexer.line, p.lexer.col)
		}
		if op1.Val.(RegisterData).Size != op2.Val.(RegisterData).Size {
			return errors.MismatchedRegisterSizes(p.lexer.line, p.lexer.col)
		}
		p.currentInst = Instruction{
			Ty: INST_TCMPRR,
			Data: InstCmpData{
				Ty:  ty,
				Min: op1.Val.(RegisterData),
				Sub: op2.Val.(RegisterData),
			},
		}
	case TOKEN_TINTEGER_LIT:
		p.currentInst = Instruction{
			Ty: INST_TCMPIR,
			Data: InstCmpData{
				Ty:  ty,
				Min: op1.Val.(RegisterData),
				Imm: op2.Val.(uint64),
			},
		}
	case TOKEN_TFLOAT_LIT:
		if ty != vm.TY_FLOAT {
			return errors.FailedToParse("cmp instruction",
				"Immediate float value comparison with non FLOAT cmp instruction",
				op2.Line, op2.Col)
		}
		p.currentInst = Instruction{
			Ty: INST_TCMPIR,
			Data: InstCmpData{
				Ty:  ty,
				Min: op1.Val.(RegisterData),
				Imm: op2.Val.(uint64),
			},
		}
	}
	return nil
}
func (p *Parser) parseJmp(ty InstTy) error {
	inst := Instruction{
		Ty: ty,
	}
	var absolute bool
	addr := Token{Ty: INVALID}
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	} else if p.lexer.CurrentToken().Ty == TOKEN_TABSOLUTE {
		absolute = true
	} else {
		p.lexer.UnreadToken()
		absolute = false
	}
	if expr, err := p.parseExpression(0); err != nil {
		return err
	} else if expr.Ty == EXPR_TDEREF && absolute {
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"ABSOLUTE jump disallowed with IP regiter relative offsets",
			p.lexer.line, p.lexer.col,
		)
	} else if expr.Ty == EXPR_TDEREF {
		return p.parseJmpIP(ty, expr)
	} else if eval, ok := TryConstEvaluateExpression(expr); eval.Ty == CONSTEXPR_TIDENT {
		addr.Ty = TOKEN_TIDENT
		addr.Val = eval.Ident
	} else if !ok {
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"Second operand to instruction must be a valid label or a an address",
			p.lexer.line, p.lexer.col)
	} else if eval.Ty == CONSTEXPR_TILIT {
		addr.Ty = TOKEN_TINTEGER_LIT
		addr.Val = eval.Val
	} else {
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"Bad expression",
			p.lexer.line, p.lexer.col)
	}
	switch addr.Ty {
	case TOKEN_TINTEGER_LIT:
		if absolute {
			return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
				"Unnecessary use of ABSOLUTE keyword",
				p.lexer.line, p.lexer.col)
		}
		inst.Data = InstJmpData{
			Address:  addr.Val.(uint64),
			Absolute: absolute,
		}
	case TOKEN_TIDENT:
		inst.Data = InstJmpData{
			Address:  addr.Val.(string),
			Absolute: absolute,
		}
	default:
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"instruction requires a valid address or label as a parameter",
			p.lexer.line, p.lexer.col)
	}
	p.currentInst = inst
	return nil
}
func (p *Parser) parseJmpIP(ty InstTy, expr *Expr) error {
	derefExpr := expr.Val.(DerefExpr)
	deref, err := p.processDeref(derefExpr.Inner)
	if err != nil {
		return err
	}
	if deref.Reg1.Reg != vm.IP_IDX && deref.Reg2.Reg != vm.IP_IDX {
		return errors.FailedToParse("ip relative jump instruction", "expression without the IP register",
			p.lexer.line, p.lexer.col)
	}
	switch deref.Ty {
	case DEREF_T1RO:
		p.currentInst = Instruction{
			Ty: INST_TJMPIP0R,
			Data: InstJmpIPData{
				JmpTy:  ty,
				Offset: deref.Offset,
				OpTy:   deref.OffsetOp,
			},
		}
	case DEREF_T2RO:
		var reg RegisterData
		if deref.Reg1.Reg == vm.IP_IDX {
			reg = deref.Reg2
		} else {
			reg = deref.Reg1
		}
		p.currentInst = Instruction{
			Ty: INST_TJMPIP1R,
			Data: InstJmpIPData{
				JmpTy:  ty,
				Offset: deref.Offset,
				Reg:    reg,
				OpTy:   deref.OffsetOp,
			},
		}
	}
	return nil
}
func (p *Parser) parseCall() error {
	ok := false
	var pruned ConstExpr
	if param, err := p.parseExpression(0); err != nil {
		return err
	} else if param.Ty == EXPR_TDEREF {
		return p.parseCallIP(param)
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
func (p *Parser) parseCallIP(expr *Expr) error {
	derefExpr := expr.Val.(DerefExpr)
	deref, err := p.processDeref(derefExpr.Inner)
	if err != nil {
		return err
	}
	if deref.Reg1.Reg != vm.IP_IDX && deref.Reg2.Reg != vm.IP_IDX {
		return errors.FailedToParse("ip relative call instruction",
			"expression without the IP register",
			p.lexer.line, p.lexer.col)
	}
	switch deref.Ty {
	case DEREF_T1RO:
		p.currentInst = Instruction{
			Ty: INST_TCALLIP0R,
			Data: InstCallIPData{
				Offset: deref.Offset,
				OpTy:   deref.OffsetOp,
			},
		}
	case DEREF_T2RO:
		var reg RegisterData
		if deref.Reg1.Reg == vm.IP_IDX {
			reg = deref.Reg2
		} else {
			reg = deref.Reg1
		}
		p.currentInst = Instruction{
			Ty: INST_TCALLIP1R,
			Data: InstCallIPData{
				Offset: deref.Offset,
				Reg:    reg,
				OpTy:   deref.OffsetOp,
			},
		}
	}
	return nil
}
