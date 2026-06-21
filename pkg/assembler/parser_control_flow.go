package assembler

import (

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
		em, _ := expr.Emit()
		return errors.FailedToParse(p.currentIdent,
			p.currentStartToken.Line, p.currentStartToken.Col,
			"First operand to instruction must be a valid register, got `%s`",
			em,
		)
	}
	if op1.Val.(RegisterData).Reg > vm.MAX_REG_IDX {
		return errors.FailedToParse(p.currentIdent,
			p.currentStartToken.Line, p.currentStartToken.Col,
			"Disallowed minuend register `%s`",
			op1.ForceValAsString(),
		)
	}
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	if p.lexer.CurrentToken().Ty != TOKEN_TCOMMA {
		t := p.lexer.CurrentToken()
		return errors.FailedToParse(p.currentIdent,
			t.Line, t.Col,
			"Instruction missing a comma, got `%s`",
			t.ForceValAsString(),
		)
	}

	op2 := Token{Ty: INVALID}
	if expr, err := p.parseExpression(0); err != nil {
		return err
	} else if eval, ok := TryConstEvaluateExpression(expr); eval.Ty == CONSTEXPR_TREG {
		op2.Ty = TOKEN_TREG
		op2.Val = eval.UnpackAsRegisterData()
	} else if !ok {
		em, _ := expr.Emit()
		return errors.FailedToParse(p.currentIdent,
			p.currentStartToken.Line, p.currentStartToken.Col,
			"Second operand to instruction must be a valid register "+
				"or a constant expression. Got `%s` instead.",
			em,
		)
	} else if eval.Ty == CONSTEXPR_TILIT {
		op2.Ty = TOKEN_TINTEGER_LIT
		op2.Val = eval.Val
	} else if eval.Ty == CONSTEXPR_TFLIT {
		op2.Ty = TOKEN_TFLOAT_LIT
		op2.Val = eval.Val
	} else {
		em, _ := expr.Emit()
		return errors.FailedToParse(p.currentIdent,
			p.currentStartToken.Line, p.currentStartToken.Col,
			"Bad expression `%s`"+
				em,
		)
	}
	switch op2.Ty {
	case TOKEN_TREG:
		if op1.Val.(RegisterData).Size != op2.Val.(RegisterData).Size {
			return errors.MismatchedRegisterSizes(op1.Line, op1.Col)
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
			p.issueWarning(
				op2.Line, op2.Col,
				"Immediate float value comparison with non FLOAT cmp instruction",
			)
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
		em, _ := expr.Emit()
		return errors.FailedToParse(p.currentIdent,
			p.currentStartToken.Line, p.currentStartToken.Col,
			"ABSOLUTE jump disallowed with IP regiter relative offsets. "+
				"In expression `%s`.",
			em,
		)
	} else if expr.Ty == EXPR_TDEREF {
		return p.parseJmpIP(ty, expr)
	} else if eval, ok := TryConstEvaluateExpression(expr); eval.Ty == CONSTEXPR_TIDENT {
		addr.Ty = TOKEN_TIDENT
		addr.Val = eval.Ident
	} else if !ok {
		em, _ := expr.Emit()
		return errors.FailedToParse(p.currentIdent,
			p.currentStartToken.Line, p.currentStartToken.Col,
			"Second operand to instruction must be a valid label or a an address. "+
				"In expression `%s`.",
			em,
		)
	} else if eval.Ty == CONSTEXPR_TILIT {
		addr.Ty = TOKEN_TINTEGER_LIT
		addr.Val = eval.Val
	} else {
		em, _ := expr.Emit()
		return errors.FailedToParse(p.currentIdent,
			p.currentStartToken.Line, p.currentStartToken.Col,
			"Bad expression `%s`.",
			em,
		)
	}
	switch addr.Ty {
	case TOKEN_TINTEGER_LIT:
		if absolute {
			p.issueWarning(
				p.currentStartToken.Line, p.currentStartToken.Col,
				"Unnecessary use of ABSOLUTE keyword",
			)
			absolute = false
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
		return errors.FailedToParse(p.currentIdent,
			addr.Line, addr.Col,
			"Instruction requires a valid address or label as a parameter, got `%s`",
			addr.ForceValAsString(),
		)
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
		em, _ := expr.Emit()
		return errors.FailedToParse(p.currentIdent,
			p.currentStartToken.Line, p.currentStartToken.Col,
			"IP-relative jump instruction dereference "+
				"expression without the IP register. "+
				"In expression: `%s`",
			em,
		)
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
		em, _ := pruned.Emit()
		return errors.FailedToParse(p.currentIdent,
			p.currentStartToken.Line, p.currentStartToken.Col,
			"Parameter of call instruction must be a compile time expression. "+
				"Expression `%s`",
			em,
		)
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
		em, _ := pruned.Emit()
		return errors.FailedToParse(p.currentIdent,
			p.currentStartToken.Line, p.currentStartToken.Col,
			"Parameter of call instruction must be an address literal or a label. "+
				"In expression `%s`",
			em,
		)
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
		em, _ := expr.Emit()
		return errors.FailedToParse(p.currentIdent,
			p.currentStartToken.Line, p.currentStartToken.Col,
			"IP-relative jump instruction dereference "+
				"expression without the IP register. "+
				"In expression: `%s`",
			em,
		)
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
