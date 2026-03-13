package assembler

import (
	"fmt"
	"math"
)

func (p *Parser) parsePush() error {
	arg, err := p.parseExpression(0)
	if err != nil {
		return err
	}
	eval, ok := TryConstEvaluateExpression(arg)
	if !ok {
		return fmt.Errorf("Operand to push instruction must be a constant expression or a register name %s",
			p.lexer.CurrentPosition())
	}
	switch eval.Ty {
	case CONSTEXPR_TREG:
		p.currentInst = Instruction{
			Ty: INST_TPUSHR,
			Data: InstPushPopData{
				Reg: eval.Val,
			},
		}
	case CONSTEXPR_TILIT:
		p.currentInst = Instruction{
			Ty: INST_TPUSHI,
			Data: InstPushPopData{
				Imm: eval.Val,
			},
		}
	case CONSTEXPR_TFLIT:
		p.currentInst = Instruction{
			Ty: INST_TPUSHI,
			Data: InstPushPopData{
				Imm: eval.Val,
			},
		}
	default:
		return fmt.Errorf("Unsupported operand expression type %s", p.lexer.CurrentPosition())
	}
	return nil
}
func (p *Parser) parsePop() error {
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	switch p.lexer.CurrentToken().Ty {
	case TOKEN_TNEWLINE:
		p.lexer.UnreadToken()
	case TOKEN_TEOF:
		p.lexer.UnreadToken()
	default:
		p.lexer.UnreadToken()
		arg, err := p.parseExpression(0)
		if err != nil {
			return err
		}
		eval, ok := TryConstEvaluateExpression(arg)
		if !ok || eval.Ty != CONSTEXPR_TREG {
			return fmt.Errorf("Operand to pop instruction can only be a register name or none %s",
				p.lexer.CurrentPosition())
		}
		p.currentInst = Instruction{
			Ty: INST_TPOP,
			Data: InstPushPopData{
				Reg: eval.Val,
			},
		}
		return nil
	}
	p.currentInst = Instruction{
		Ty: INST_TPOP,
		Data: InstPushPopData{
			Imm: uint64(math.MaxUint64),
			Reg: 0xff,
		},
	}
	return nil
}
