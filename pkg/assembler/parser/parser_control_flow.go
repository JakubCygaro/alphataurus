package assembler

import (
	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func (p *Parser) parseCmp() error {
	genericCmp := InstGenericCmp{}
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	op1 := p.lexer.CurrentToken()
	ty := vm.TY_UINT
	switch op1.Ty {
	case lx.TOKEN_TFLOAT:
		ty = vm.TY_FLOAT
	case lx.TOKEN_TSIGNED:
		ty = vm.TY_SINT
	case lx.TOKEN_TUNSIGNED:
		ty = vm.TY_UINT
	default:
		p.lexer.UnreadCurrentToken()
	}
	genericCmp.Ty = ty
	if expr, err := p.ParseExpression(); err != nil {
		return err
	} else {
		genericCmp.Min = expr
	}
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	if p.lexer.CurrentToken().Ty != lx.TOKEN_TCOMMA {
		t := p.lexer.CurrentToken()
		return errors.FailedToParse(p.currentIdent,
			t.Line, t.Col,
			"Instruction missing a comma, got `%s`",
			t.ForceValAsString(),
		)
	}
	if expr, err := p.ParseExpression(); err != nil {
		return err
	} else {
		genericCmp.Sub = expr
	}
	if concrete, err := GetConcreteCmpInst(genericCmp); concrete != nil && err == nil {
		p.currentInst = Instruction{
			Data: concrete,
		}
	} else {
		p.currentInst = Instruction{
			Data: genericCmp,
		}
	}
	return nil
}
func (p *Parser) parseJmp(ty JmpVariant) error {
	genericJmp := InstGenericJmp{}
	var absolute bool
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	} else if p.lexer.CurrentToken().Ty == lx.TOKEN_TABSOLUTE {
		absolute = true
	} else {
		p.lexer.UnreadCurrentToken()
		absolute = false
	}
	if expr, err := p.ParseExpression(); err != nil {
		return err
	} else {
		genericJmp.Address = expr
	}
	genericJmp.Absolute = absolute
	genericJmp.Variant = ty
	// else if expr.Ty == EXPR_TDEREF && absolute {
	// 	em, _ := expr.Emit()
	// 	return errors.FailedToParse(p.currentIdent,
	// 		p.currentStartToken.Line, p.currentStartToken.Col,
	// 		"ABSOLUTE disallowed relative jumps. "+
	// 			"In expression `%s`.",
	// 		em,
	// 	)
	// } else if expr.Ty == EXPR_TDEREF {
	// 	return p.parseJmpIP(ty, expr)
	// } else if eval, ok := TryConstEvaluateExpression(expr); eval.Ty == CONSTEXPR_TIDENT {
	// 	addr.Ty = lx.TOKEN_TIDENT
	// 	addr.Val = eval.Ident
	// } else if !ok {
	// 	em, _ := expr.Emit()
	// 	return errors.FailedToParse(p.currentIdent,
	// 		p.currentStartToken.Line, p.currentStartToken.Col,
	// 		"Second operand to instruction must be a valid label or a an address. "+
	// 			"In expression `%s`.",
	// 		em,
	// 	)
	// } else if eval.Ty == CONSTEXPR_TILIT {
	// 	addr.Ty = lx.TOKEN_TINTEGER_LIT
	// 	addr.Val = eval.Val
	// } else {
	// 	em, _ := expr.Emit()
	// 	return errors.FailedToParse(p.currentIdent,
	// 		p.currentStartToken.Line, p.currentStartToken.Col,
	// 		"Bad expression `%s`.",
	// 		em,
	// 	)
	// }
	// switch addr.Ty {
	// case lx.TOKEN_TINTEGER_LIT:
	// 	if absolute {
	// 		p.issueWarning(
	// 			p.currentStartToken.Line, p.currentStartToken.Col,
	// 			"Unnecessary use of ABSOLUTE keyword",
	// 		)
	// 		absolute = false
	// 	}
	// 	inst.Data = InstJmp{
	// 		Address:  addr.Val.(uint64),
	// 		Absolute: absolute,
	// 		Variant:  ty,
	// 	}
	// case lx.TOKEN_TIDENT:
	// 	inst.Data = InstJmp{
	// 		Address:  addr.Val.(string),
	// 		Absolute: absolute,
	// 		Variant:  ty,
	// 	}
	// default:
	// 	return errors.FailedToParse(p.currentIdent,
	// 		addr.Line, addr.Col,
	// 		"Instruction requires a valid address or label as a parameter, got `%s`",
	// 		addr.ForceValAsString(),
	// 	)
	// }
	p.currentInst = Instruction{
		Data: genericJmp,
	}
	return nil
}

//	func (p *Parser) parseJmpIP(ty JmpVariant, expr *Expr) error {
//		derefExpr := expr.Val.(DerefExpr)
//		deref, err := p.processDeref(derefExpr.Inner)
//		if err != nil {
//			return err
//		}
//		if deref.Reg1.Reg != vm.IP_IDX && deref.Reg2.Reg != vm.IP_IDX {
//			em, err := expr.Emit()
//			if err != nil {
//				return err
//			}
//			return errors.FailedToParse(p.currentIdent,
//				p.currentStartToken.Line, p.currentStartToken.Col,
//				"IP-relative jump instruction dereference "+
//					"expression without the IP register. "+
//					"In expression: `%s`",
//				em,
//			)
//		}
//		switch deref.Ty {
//		case DEREF_T1RO:
//			p.currentInst = Instruction{
//				Data: InstJmpIP0R{
//					JmpTy:  ty,
//					Offset: deref.Offset,
//					OpTy:   deref.OffsetOp,
//				},
//			}
//		case DEREF_T2RO:
//			var reg lx.RegisterData
//			if deref.Reg1.Reg == vm.IP_IDX {
//				reg = deref.Reg2
//			} else {
//				reg = deref.Reg1
//			}
//			p.currentInst = Instruction{
//				Data: InstJmpIP1R{
//					JmpTy:  ty,
//					Offset: deref.Offset,
//					Reg:    reg,
//					OpTy:   deref.OffsetOp,
//				},
//			}
//		}
//		return nil
//	}
func (p *Parser) parseCall() error {
	genericCall := InstGenericCall{}
	// ok := false
	// var pruned ConstExpr
	if param, err := p.ParseExpression(); err != nil {
		return err
	} else {
		genericCall.Expr = param
	}
	// else if param.Ty == EXPR_TDEREF {
	// 	return p.parseCallIP(param)
	// } else if pruned, ok = TryConstEvaluatePruneExpression(param); !ok {
	// 	em, _ := pruned.Emit()
	// 	return errors.FailedToParse(p.currentIdent,
	// 		p.currentStartToken.Line, p.currentStartToken.Col,
	// 		"Parameter of call instruction must be a compile time expression. "+
	// 			"Expression `%s`",
	// 		em,
	// 	)
	// }
	// switch pruned.Ty {
	// case CONSTEXPR_TILIT:
	// 	p.currentInst = Instruction{
	// 		Data: InstCall{
	// 			Addr: pruned.Val,
	// 			Expr: nil,
	// 		},
	// 	}
	// case CONSTEXPR_TIDENT:
	// 	p.currentInst = Instruction{
	// 		Data: InstCall{
	// 			Addr:  pruned.Val,
	// 			Ident: pruned.Ident,
	// 			Expr:  nil,
	// 		},
	// 	}
	// default:
	// 	em, _ := pruned.Emit()
	// 	return errors.FailedToParse(p.currentIdent,
	// 		p.currentStartToken.Line, p.currentStartToken.Col,
	// 		"Parameter of call instruction must be an address literal or a label. "+
	// 			"In expression `%s`",
	// 		em,
	// 	)
	// }
	return nil
}

// func (p *Parser) parseCallIP(expr *Expr) error {
// 	derefExpr := expr.Val.(DerefExpr)
// 	deref, err := p.processDeref(derefExpr.Inner)
// 	if err != nil {
// 		return err
// 	}
// 	if deref.Reg1.Reg != vm.IP_IDX && deref.Reg2.Reg != vm.IP_IDX {
// 		em, _ := expr.Emit()
// 		return errors.FailedToParse(p.currentIdent,
// 			p.currentStartToken.Line, p.currentStartToken.Col,
// 			"IP-relative jump instruction dereference "+
// 				"expression without the IP register. "+
// 				"In expression: `%s`",
// 			em,
// 		)
// 	}
// 	switch deref.Ty {
// 	case DEREF_T1RO:
// 		p.currentInst = Instruction{
// 			Data: InstCallIP0R{
// 				Offset: deref.Offset,
// 				OpTy:   deref.OffsetOp,
// 			},
// 		}
// 	case DEREF_T2RO:
// 		var reg lx.RegisterData
// 		if deref.Reg1.Reg == vm.IP_IDX {
// 			reg = deref.Reg2
// 		} else {
// 			reg = deref.Reg1
// 		}
// 		p.currentInst = Instruction{
// 			Data: InstCallIP1R{
// 				Offset: deref.Offset,
// 				Reg:    reg,
// 				OpTy:   deref.OffsetOp,
// 			},
// 		}
// 	}
// 	return nil
// }
