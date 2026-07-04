package assembler

import (
	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
)

func (p *Parser) parsePush() error {
	genericPush := InstGenericPush{}
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	nextT := p.lexer.CurrentToken()
	// WORD by default
	if sz, ok := lx.TokenAsSize(&nextT); ok {
		genericPush.DataSz = &DataSize{
			Line: nextT.Line,
			Col: nextT.Col,
			Size: sz,
		}
	} else {
		p.lexer.UnreadCurrentToken()
	}
	if arg, err := p.ParseExpression(); err != nil {
		return err
	} else {
		genericPush.Expr = arg
	}
	p.currentInst = Instruction{
		Data: genericPush,
	}
	// eval, ok := TryConstEvaluateExpression(arg)
	// if !ok {
	// 	return fmt.Errorf("Operand to push instruction must be a constant expression or a register name %s",
	// 		p.lexer.CurrentPosition())
	// }
	// switch eval.Ty {
	// case CONSTEXPR_TREG:
	// 	if dataSz != 0xff {
	// 		p, _ := lx.GetSizeKeyword(dataSz)
	// 		return errors.UnnecessarySizeParameter(p, nextT.Line, nextT.Col)
	// 	}
	// 	regData := eval.UnpackAsRegisterData()
	// 	p.currentInst = Instruction{
	// 		Data: InstPushR{
	// 			Reg:    uint64(regData.Reg),
	// 			DataSz: regData.Size,
	// 		},
	// 	}
	// case CONSTEXPR_TILIT:
	// 	p.currentInst = Instruction{
	// 		Data: InstPushI{
	// 			Imm:    eval.Val,
	// 			DataSz: dataSz,
	// 		},
	// 	}
	// case CONSTEXPR_TFLIT:
	// 	if dataSz != vm.SZ_64 {
	// 		g, _ := lx.GetSizeKeyword(dataSz)
	// 		n, _ := lx.GetSizeKeyword(vm.SZ_64)
	// 		return errors.BadSizeArgument(
	// 			g,
	// 			n,
	// 			nextT.Line,
	// 			nextT.Col,
	// 		)
	// 	}
	// 	p.currentInst = Instruction{
	// 		Data: InstPushI{
	// 			Imm:    eval.Val,
	// 			DataSz: dataSz,
	// 		},
	// 	}
	// default:
	// 	em, _ := arg.Emit()
	// 	return errors.FailedToParse(p.currentIdent,
	// 		arg.Line, arg.Col,
	// 		"Unsupported operand expression type `%s`",
	// 		em,
	// 	)
	// }
	return nil
}
func (p *Parser) parsePop() error {
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	t := p.lexer.CurrentToken()
	// pop BYTE/QUARTER/HALF/WORD case
	if sz, ok := lx.TokenAsSize(&t); ok {
		p.currentInst = Instruction{
			Data: InstPop{
				DataSz: sz,
			},
		}
		return nil
	} else {
		p.lexer.UnreadCurrentToken()
	}
	// pop r0, r0b, r0q, r0h case
	arg, err := p.parseExpression(0)
	if err != nil {
		return err
	}
	if !arg.IsRegexpr() {
		em, _ := arg.Emit()
		return errors.FailedToParse(p.currentIdent,
			arg.Line, arg.Col,
			"Operand to pop instruction can only be a register name or none, got `%s`",
			em)
	}
	regData := arg.Val.(RegExpr).Reg
	p.currentInst = Instruction{
		Data: InstPopR{
			Reg:    uint64(regData.Reg),
			DataSz: regData.Size,
		},
	}
	return nil
}
