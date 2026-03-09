package assembler

import "github.com/JakubCygaro/alphataurus/assembler/errors"

func (p *Parser) parseMov() error {
	var op1 Token
	if expr, err := p.parseExpression(0); err != nil {
		return err
	} else if expr.Ty == EXPR_TDEREF {
		inner, _ := TryEvaluateExpression(expr.Val.(DerefExpr).Inner)
		return p.parseMovDeref(inner)
	} else if eval, _ := TryConstEvaluatePruneExpression(expr); eval.Ty != CONSTEXPR_TREG {
		return errors.FailedToParse("mov instruction",
			"First operand to instruction must be a valid register or dereference expression",
			p.lexer.line, p.lexer.col)
	} else {
		op1.val = int(eval.Val)
	}
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	comma := p.lexer.CurrentToken()
	if comma.Ty != TOKEN_TCOMMA {
		return errors.FailedToParse("mov instruction",
			"Instruction missing a comma",
			p.lexer.line, p.lexer.col)
	}

	var op2 ConstExpr
	if expr, err := p.parseExpression(0); err != nil {
		return err
	} else {
		eval, _ := TryEvaluatePruneExpression(expr)
		switch eval.Ty {
		case EXPR_TCONST:
			op2 = eval.Val.(ConstExpr)
		case EXPR_TDEREF:
			return p.parseDerefMov(op1.val.(int), eval.Val.(DerefExpr).Inner)
		default:
			return errors.FailedToParse("mov instruction",
				"Second operand to instruction has to be a valid register, dereference, or a compile time expression",
				p.lexer.line, p.lexer.col)
		}
	}
	switch op2.Ty {
	case CONSTEXPR_TREG:
		p.currentInst = Instruction{
			Ty: INST_TMOVRR,
			Data: InstMovData{
				Src:  int(op2.Val),
				Dest: op1.val.(int),
			},
		}
	case CONSTEXPR_TILIT:
		p.currentInst = Instruction{
			Ty: INST_TMOVIR,
			Data: InstMovData{
				Imm:  op2.Val,
				Dest: op1.val.(int),
			},
		}
	case CONSTEXPR_TFLIT:
		p.currentInst = Instruction{
			Ty: INST_TMOVIR,
			Data: InstMovData{
				Imm:  op2.Val,
				Dest: op1.val.(int),
			},
		}
	default:
		return errors.FailedToParse("mov instruction",
			"Second operand to instruction has to be a valid register, dereference or a compile time expression",
			p.lexer.line, p.lexer.col)
	}
	return nil
}

// move deref to somewhere
// e.g: mov r0, [bp]
func (p *Parser) parseDerefMov(reg int, inner *Expr) error {
	dData, err := p.processDeref(inner)
	if err != nil {
		return err
	}
	switch dData.Ty {
	case DEREF_T0RO:
		p.currentInst = Instruction{
			Ty: INST_TMOVDRI,
			Data: InstDerefMovData{
				Dest:   reg,
				Offset: dData.Offset,
			},
		}
	case DEREF_T1RO:
		p.currentInst = Instruction{
			Ty: INST_TMOVDRO1,
			Data: InstDerefMovData{
				Dest:   reg,
				OReg1:  dData.Reg1,
				Offset: dData.Offset,
				OpTy:   dData.OffsetOp,
			},
		}
	case DEREF_T2RO:
		p.currentInst = Instruction{
			Ty: INST_TMOVDRO2,
			Data: InstDerefMovData{
				Dest:   reg,
				OReg1:  dData.Reg1,
				OReg2:  dData.Reg2,
				Offset: dData.Offset,
				OpTy:   dData.OffsetOp,
			},
		}
	default:
		return errors.FailedToParse("mov instruction",
			"Invalid dereference expression",
			p.lexer.line, p.lexer.col)
	}
	return nil
}

// move something into deref
// e.g: mov [bp], r0
func (p *Parser) parseMovDeref(inner *Expr) error {
	dData, err := p.processDeref(inner)
	if err != nil {
		return err
	}
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	comma := p.lexer.CurrentToken()
	if comma.Ty != TOKEN_TCOMMA {
		return errors.FailedToParse("mov instruction",
			"Instruction missing a comma",
			p.lexer.line, p.lexer.col)
	}
	var op2 ConstExpr
	if expr, err := p.parseExpression(0); err != nil {
		return err
	} else {
		eval, _ := TryEvaluatePruneExpression(expr)
		switch eval.Ty {
		case EXPR_TCONST:
			op2 = eval.Val.(ConstExpr)
		default:
			return errors.FailedToParse("mov instruction",
				"Second operand to instruction has to be a valid register or a compile time expression",
				p.lexer.line, p.lexer.col)
		}
	}
	mddata := InstMovDerefData{}
	switch op2.Ty {
	case CONSTEXPR_TREG:
		mddata.SourceReg = int(op2.Val)
	case CONSTEXPR_TILIT:
		mddata.Imm = uint64(op2.Val)
		mddata.SourceReg = INVALID
	//TODO: label support
	default:
		return errors.FailedToParse("mov instruction",
			"Second operand to instruction has to be a valid register or a compile time expression",
			p.lexer.line, p.lexer.col)
	}
	switch dData.Ty {
	case DEREF_T0RO:
		mddata.Offset = dData.Offset
		ty := INST_TMOVRD
		if mddata.SourceReg == INVALID {
			ty = INST_TMOVID
		}
		p.currentInst = Instruction{
			Ty:   ty,
			Data: mddata,
		}
	case DEREF_T1RO:
		mddata.OReg1 = dData.Reg1
		mddata.Offset = dData.Offset
		mddata.OpTy = dData.OffsetOp
		ty := INST_TMOVRDO1
		if mddata.SourceReg == INVALID {
			ty = INST_TMOVIDO1
		}
		p.currentInst = Instruction{
			Ty:   ty,
			Data: mddata,
		}
	case DEREF_T2RO:
		mddata.OReg1 = dData.Reg1
		mddata.OReg2 = dData.Reg2
		mddata.Offset = dData.Offset
		mddata.OpTy = dData.OffsetOp
		ty := INST_TMOVRDO2
		if mddata.SourceReg == INVALID {
			ty = INST_TMOVIDO2
		}
		p.currentInst = Instruction{
			Ty:   ty,
			Data: mddata,
		}
	default:
		return errors.FailedToParse("mov instruction",
			"Invalid dereference expression",
			p.lexer.line, p.lexer.col)
	}
	return nil
}
