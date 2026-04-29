package assembler

import "github.com/JakubCygaro/alphataurus/pkg/assembler/errors"

func (p *Parser) parseMov() error {
	var op1 Token
	var sized byte = 0xff
	var sizedL, sizedC uint64
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	if sz, ok := TokenAsSize(&op1); !ok {
		p.lexer.UnreadToken()
	} else {
		sized = sz
		sizedL = p.lexer.CurrentToken().Line
		sizedC = p.lexer.CurrentToken().Col
		op1 = Token{}
	}
	if expr, err := p.parseExpression(0); err != nil {
		return err

	} else if expr.Ty == EXPR_TDEREF {
		inner, _ := TryEvaluateExpression(expr.Val.(DerefExpr).Inner)
		return p.parseMovDeref(inner, sized)

	} else if eval, _ := TryConstEvaluatePruneExpression(expr); eval.Ty != CONSTEXPR_TREG {
		return errors.FailedToParse("mov instruction",
			"First operand to instruction must be a valid register or dereference expression",
			expr.Line, expr.Col)

	} else {
		op1.Val = eval.UnpackAsRegisterData()
	}
	if sized != 0xff {
		s, _ := GetSizeKeyword(sized)
		return errors.UnnecessarySizeParameter(s, sizedL, sizedC)
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
			return p.parseDerefMov(op1.Val.(RegisterData), eval.Val.(DerefExpr).Inner)
		default:
			return errors.FailedToParse("mov instruction",
				"Second operand to instruction has to be a valid register, dereference, or a compile time expression",
				p.lexer.line, p.lexer.col)
		}
	}

	destData := op1.Val.(RegisterData)
	switch op2.Ty {
	case CONSTEXPR_TREG:
		srcData := op2.UnpackAsRegisterData()
		if destData.Size < srcData.Size {
			return errors.MismatchedRegisterSizes(op1.Line, op1.Col)
		}
		p.currentInst = Instruction{
			Ty: INST_TMOVRR,
			Data: InstMovData{
				Src:      srcData.Reg,
				Dest:     destData.Reg,
				DataSize: destData.Size,
			},
		}
	case CONSTEXPR_TILIT:
		p.currentInst = Instruction{
			Ty: INST_TMOVIR,
			Data: InstMovData{
				Imm:      op2.Val,
				Dest:     destData.Reg,
				DataSize: destData.Size,
			},
		}
	case CONSTEXPR_TFLIT:
		p.currentInst = Instruction{
			Ty: INST_TMOVIR,
			Data: InstMovData{
				Imm:      op2.Val,
				Dest:     destData.Reg,
				DataSize: destData.Size,
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
func (p *Parser) parseDerefMov(reg RegisterData, inner *Expr) error {
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
func (p *Parser) parseMovDeref(inner *Expr, sized byte) error {
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
		mddata.SourceReg = op2.UnpackAsRegisterData()
	case CONSTEXPR_TILIT:
		mddata.Imm = uint64(op2.Val)
		mddata.SourceReg = GetInvalidRegister()
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
		if mddata.SourceReg.IsInvalidRegister() {
			ty = INST_TMOVID
		}
		// if this is an immediate move into a deref we need a size parameter
		// like mov WORD [bp], 100
		if ty == INST_TMOVID && sized == 0xff {
			return errors.MissingDataSize(inner.Line, inner.Col)
		} else if ty != INST_TMOVID && sized != 0xff {
			s, _ := GetSizeKeyword(sized)
			return errors.UnnecessarySizeParameter(s, inner.Line, inner.Col)
		}
		mddata.DataSize = sized
		p.currentInst = Instruction{
			Ty:   ty,
			Data: mddata,
		}
	case DEREF_T1RO:
		mddata.OReg1 = dData.Reg1
		mddata.Offset = dData.Offset
		mddata.OpTy = dData.OffsetOp
		ty := INST_TMOVRDO1
		if mddata.SourceReg.IsInvalidRegister() {
			ty = INST_TMOVIDO1
		}
		// like mov WORD [bp+1], 100
		if ty == INST_TMOVIDO1 && sized == 0xff {
			return errors.MissingDataSize(inner.Line, inner.Col)
		} else if ty != INST_TMOVIDO1  && sized != 0xff {
			s, _ := GetSizeKeyword(sized)
			return errors.UnnecessarySizeParameter(s, inner.Line, inner.Col)
		}
		mddata.DataSize = sized
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
		if mddata.SourceReg.IsInvalidRegister() {
			ty = INST_TMOVIDO2
		}
		// like mov WORD [bp+r0+1], 100
		if ty == INST_TMOVIDO2 && sized == 0xff {
			return errors.MissingDataSize(inner.Line, inner.Col)
		} else if ty != INST_TMOVIDO2  && sized != 0xff {
			s, _ := GetSizeKeyword(sized)
			return errors.UnnecessarySizeParameter(s, inner.Line, inner.Col)
		}
		mddata.DataSize = sized
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
