package assembler

import (
	"fmt"
	"math"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func (p *Parser) parsePush() error {
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	nextT := p.lexer.CurrentToken()
	// WORD by default
	dataSz := byte(0xff)
	if sz, ok := TokenAsSize(&nextT); ok {
		dataSz = sz
		if err := p.lexer.ReadNextToken(); err != nil {
			return err
		}
	} else {
		p.lexer.UnreadToken()
	}
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
		if dataSz != 0xff {
			p, _ := GetSizeKeyword(dataSz)
			return errors.UnnecessarySizeParameter(p, nextT.Line, nextT.Col)
		}
		regData := eval.UnpackAsRegisterData()
		p.currentInst = Instruction{
			Ty: INST_TPUSHR,
			Data: InstPushPopData{
				Reg:    uint64(regData.Reg),
				DataSz: regData.Size,
			},
		}
	case CONSTEXPR_TILIT:
		p.currentInst = Instruction{
			Ty: INST_TPUSHI,
			Data: InstPushPopData{
				Imm:    eval.Val,
				DataSz: dataSz,
			},
		}
	case CONSTEXPR_TFLIT:
		if dataSz != vm.SZ_64 {
			g, _ := GetSizeKeyword(dataSz)
			n, _ := GetSizeKeyword(vm.SZ_64)
			return errors.BadSizeArgument(
				g,
				n,
				nextT.Line,
				nextT.Col,
			)
		}
		p.currentInst = Instruction{
			Ty: INST_TPUSHI,
			Data: InstPushPopData{
				Imm:    eval.Val,
				DataSz: dataSz,
			},
		}
	default:
		return errors.FailedToParse("push instruction",
			"Unsupported operand expression type %s", arg.Line, arg.Col)
	}
	return nil
}
func (p *Parser) parsePop() error {
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	t := p.lexer.CurrentToken()
	// pop BYTE/QUARTER/HALF/WORD case
	if sz, ok := TokenAsSize(&t); ok {
		p.currentInst = Instruction{
			Ty: INST_TPOP,
			Data: InstPushPopData{
				Imm:    uint64(math.MaxUint64),
				Reg:    uint64(math.MaxUint64),
				DataSz: sz,
			},
		}
		return nil
	} else {
		p.lexer.UnreadToken()
	}
	// pop r0, r0b, r0q, r0h case
	arg, err := p.parseExpression(0)
	if err != nil {
		return err
	}
	eval, ok := TryConstEvaluateExpression(arg)
	if !ok || eval.Ty != CONSTEXPR_TREG {
		return fmt.Errorf(
			"Operand to pop instruction can only be a register name or none %s",
			p.lexer.CurrentPosition())
	}
	regData := eval.UnpackAsRegisterData()
	p.currentInst = Instruction{
		Ty: INST_TPOP,
		Data: InstPushPopData{
			Reg:    uint64(regData.Reg),
			DataSz: regData.Size,
		},
	}
	return nil
}
