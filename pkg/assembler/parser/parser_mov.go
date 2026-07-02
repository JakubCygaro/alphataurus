package assembler

import (
	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
	// "github.com/JakubCygaro/alphataurus/pkg/vm"
)

func (p *Parser) parseMov() error {
	genericMov := InstGenericMov{}
	var op1 lx.Token
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	op1 = p.lexer.CurrentToken()
	if sz, ok := lx.TokenAsSize(&op1); !ok {
		p.lexer.UnreadCurrentToken()
	} else {
		genericMov.DataSize = &MovSize{
			Size: sz,
			Col: op1.Col,
			Line: op1.Line,
		}
		op1 = lx.Token{}
	}
	if expr, err := p.ParseExpression(); err != nil {
		return err
	} else {
		genericMov.Dest = expr
	}

	// else if expr.Ty == EXPR_TDEREF {
	// 	inner, _ := TryEvaluateExpression(expr.Val.(DerefExpr).Inner)
	// 	return p.parseMovDeref(inner, sized)
	//
	// } else if eval, _ := TryConstEvaluatePruneExpression(expr); eval.Ty != CONSTEXPR_TREG {
	// 	em, _ := expr.Emit()
	// 	return errors.FailedToParse(p.currentIdent,
	// 		expr.Line, expr.Col,
	// 		"First operand to instruction must be a valid register "+
	// 			"or dereference expression. In expression `%s`",
	// 		em,
	// 	)
	//
	// } else {
	// 	op1.Val = eval.UnpackAsRegisterData()
	// }
	// if sized != 0xff {
	// 	s, _ := lx.GetSizeKeyword(sized)
	// 	return errors.UnnecessarySizeParameter(s, sizedL, sizedC)
	// }
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
	// var op2 ConstExpr
	if expr, err := p.ParseExpression(); err != nil {
		return err
	} else {
		genericMov.Src = expr
	}
	p.currentInst = Instruction{
		Data: genericMov,
	}
	// else {
	// 	eval, _ := TryEvaluatePruneExpression(expr)
	// 	switch eval.Ty {
	// 	case EXPR_TCONST:
	// 		op2 = eval.Val.(ConstExpr)
	// 	case EXPR_TDEREF:
	// 		return p.parseDerefMov(op1.Val.(lx.RegisterData), eval.Val.(DerefExpr).Inner)
	// 	default:
	// 		return errors.FailedToParse(p.currentIdent,
	// 			expr.Line, expr.Col,
	// 			"Second operand to instruction has to be a valid register,"+
	// 				" dereference, or a compile time expression. In expression `%s`",
	// 			comma.ForceValAsString(),
	// 		)
	// 	}
	// }
	//
	// destData := op1.Val.(lx.RegisterData)
	// switch op2.Ty {
	// case CONSTEXPR_TREG:
	// 	srcData := op2.UnpackAsRegisterData()
	// 	if destData.Size < srcData.Size {
	// 		return errors.MismatchedRegisterSizes(op1.Line, op1.Col)
	// 	}
	// 	p.currentInst = Instruction{
	// 		Data: InstMovRR{
	// 			Src:      srcData.Reg,
	// 			Dest:     destData.Reg,
	// 			DataSize: destData.Size,
	// 		},
	// 	}
	// case CONSTEXPR_TILIT:
	// 	p.currentInst = Instruction{
	// 		Data: InstMovIR{
	// 			Imm:      op2.Val,
	// 			Dest:     destData.Reg,
	// 			DataSize: destData.Size,
	// 		},
	// 	}
	// case CONSTEXPR_TFLIT:
	// 	p.currentInst = Instruction{
	// 		Data: InstMovIR{
	// 			Imm:      op2.Val,
	// 			Dest:     destData.Reg,
	// 			DataSize: destData.Size,
	// 		},
	// 	}
	// default:
	// 	em, _ := op2.Emit()
	// 	return errors.FailedToParse(p.currentIdent,
	// 		p.currentInst.Line, p.currentInst.Col,
	// 		"Second operand to instruction has to be a valid register"+
	// 			", dereference or a compile time expression. In expression `%s`",
	// 		em,
	// 	)
	// }
	return nil
}

// move deref to somewhere
// e.g: mov r0, [bp]
// func (p *Parser) parseDerefMov(reg lx.RegisterData, inner *Expr) error {
// 	dData, err := p.processDeref(inner)
// 	if err != nil {
// 		return err
// 	}
// 	switch dData.Ty {
// 	case DEREF_T0RO:
// 		p.currentInst = Instruction{
// 			Data: InstMovDRI{
// 				Dest:   reg,
// 				Offset: dData.Offset,
// 			},
// 		}
// 	case DEREF_T1RO:
// 		p.currentInst = Instruction{
// 			Data: InstMovDRO1{
// 				Dest:   reg,
// 				OReg1:  dData.Reg1,
// 				Offset: dData.Offset,
// 				OpTy:   dData.OffsetOp,
// 			},
// 		}
// 	case DEREF_T2RO:
// 		p.currentInst = Instruction{
// 			Data: InstMovDRO2{
// 				Dest:   reg,
// 				OReg1:  dData.Reg1,
// 				OReg2:  dData.Reg2,
// 				Offset: dData.Offset,
// 				OpTy:   dData.OffsetOp,
// 			},
// 		}
// 	default:
// 		em, _ := inner.Emit()
// 		return errors.FailedToParse(p.currentIdent,
// 			inner.Line, inner.Col,
// 			"Invalid dereference expression `%s`",
// 			em,
// 		)
// 	}
// 	return nil
// }
//
// // move something into deref
// // e.g: mov [bp], r0
// func (p *Parser) parseMovDeref(inner *Expr, sized byte) error {
// 	dData, err := p.processDeref(inner)
// 	if err != nil {
// 		return err
// 	}
// 	if err := p.lexer.ReadNextToken(); err != nil {
// 		return err
// 	}
// 	comma := p.lexer.CurrentToken()
// 	if comma.Ty != lx.TOKEN_TCOMMA {
// 		return errors.FailedToParse(p.currentIdent,
// 			comma.Line, comma.Col,
// 			"Instruction missing a comma, got `%s`",
// 			comma.ForceValAsString(),
// 		)
// 	}
// 	var op2 ConstExpr
// 	if expr, err := p.parseExpression(0); err != nil {
// 		return err
// 	} else {
// 		eval, _ := TryEvaluatePruneExpression(expr)
// 		switch eval.Ty {
// 		case EXPR_TCONST:
// 			op2 = eval.Val.(ConstExpr)
// 		default:
// 			em, _ := expr.Emit()
// 			return errors.FailedToParse(p.currentIdent,
// 				expr.Line, expr.Col,
// 				"Second operand to instruction has to be a valid"+
// 					" register or a compile time expression. In expression `%s`",
// 				em,
// 			)
// 		}
// 	}
// 	var src lx.RegisterData
// 	var imm uint64
// 	switch op2.Ty {
// 	case CONSTEXPR_TREG:
// 		src = op2.UnpackAsRegisterData()
// 	case CONSTEXPR_TILIT:
// 		imm = uint64(op2.Val)
// 		src = lx.GetInvalidRegister()
// 	//TODO: label support
// 	default:
// 		em, _ := op2.Emit()
// 		return errors.FailedToParse(p.currentIdent,
// 			p.currentInst.Line, p.currentInst.Line,
// 			"Second operand to instruction has to be a valid"+
// 				" register or a compile time expression. In expression `%s`",
// 			em,
// 		)
// 	}
// 	switch dData.Ty {
// 	case DEREF_T0RO:
// 		if src.IsInvalidRegister() {
// 			mddata := InstMovID{}
// 			mddata.Offset = dData.Offset
// 			mddata.Imm = imm
// 			// if this is an immediate move into a deref we need a size parameter
// 			// like mov WORD [bp], 100
// 			if sized == 0xff {
// 				return errors.MissingDataSize(inner.Line, inner.Col)
// 			}
// 			mddata.DataSize = sized
// 			p.currentInst = Instruction{
// 				Data: mddata,
// 			}
// 		} else {
// 			mddata := InstMovRD{}
// 			mddata.Offset = dData.Offset
// 			mddata.Src = src
// 			// we dont want a size parameter
// 			if sized != 0xff {
// 				s, _ := lx.GetSizeKeyword(sized)
// 				return errors.UnnecessarySizeParameter(s, inner.Line, inner.Col)
// 			}
// 			p.currentInst = Instruction{
// 				Data: mddata,
// 			}
// 		}
// 	case DEREF_T1RO:
// 		if src.IsInvalidRegister() {
// 			mddata := InstMovIDO1{}
// 			mddata.OReg1 = dData.Reg1
// 			mddata.Offset = dData.Offset
// 			mddata.OpTy = dData.OffsetOp
// 			mddata.Imm = imm
// 			if sized == 0xff {
// 				return errors.MissingDataSize(inner.Line, inner.Col)
// 			}
// 			mddata.DataSize = sized
// 			mddata.NoOff = mddata.Offset == 0
// 			p.currentInst = Instruction{
// 				Data: mddata,
// 			}
// 		} else {
// 			mddata := InstMovRDO1{}
// 			mddata.OReg1 = dData.Reg1
// 			mddata.Offset = dData.Offset
// 			mddata.OpTy = dData.OffsetOp
// 			mddata.Src = src
// 			if sized != 0xff {
// 				s, _ := lx.GetSizeKeyword(sized)
// 				return errors.UnnecessarySizeParameter(s, inner.Line, inner.Col)
// 			}
// 			p.currentInst = Instruction{
// 				Data: mddata,
// 			}
// 		}
// 	case DEREF_T2RO:
// 		if src.IsInvalidRegister() {
// 			mddata := InstMovIDO2{}
// 			mddata.OReg1 = dData.Reg1
// 			mddata.OReg2 = dData.Reg2
// 			mddata.Offset = dData.Offset
// 			mddata.OpTy = dData.OffsetOp
// 			mddata.Imm = imm
// 			if sized == 0xff {
// 				return errors.MissingDataSize(inner.Line, inner.Col)
// 			}
// 			mddata.DataSize = sized
// 			mddata.NoOff = mddata.Offset == 0
// 			p.currentInst = Instruction{
// 				Data: mddata,
// 			}
// 		} else {
// 			mddata := InstMovRDO2{}
// 			mddata.OReg1 = dData.Reg1
// 			mddata.OReg2 = dData.Reg2
// 			mddata.Offset = dData.Offset
// 			mddata.OpTy = dData.OffsetOp
// 			mddata.Src = src
// 			if sized != 0xff {
// 				s, _ := lx.GetSizeKeyword(sized)
// 				return errors.UnnecessarySizeParameter(s, inner.Line, inner.Col)
// 			}
// 			p.currentInst = Instruction{
// 				Data: mddata,
// 			}
// 		}
// 	default:
// 		em, _ := op2.Emit()
// 		return errors.FailedToParse(p.currentIdent,
// 			p.currentInst.Line, p.currentInst.Line,
// 			"Invalid dereference expression"+
// 				". In expression `%s`",
// 			em,
// 		)
// 	}
// 	return nil
// }
