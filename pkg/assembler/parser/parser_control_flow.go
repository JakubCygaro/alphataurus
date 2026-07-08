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
		genericJmp.Expr = expr
	}
	genericJmp.Absolute = absolute
	genericJmp.Variant = ty
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
	p.currentInst = Instruction{
		Data: genericCall,
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
