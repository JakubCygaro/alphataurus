package assembler

import (
	"fmt"

	"github.com/JakubCygaro/alphataurus/internal/vm"
)

func (p *Parser) parseCmp() error {
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	op1 := p.lexer.CurrentToken()
	ty := vm.TY_INT64
	switch op1.Ty {
	case TOKEN_TEOF:
		return p.prematureEndError()
	case TOKEN_TNEWLINE:
		return fmt.Errorf("Malformed cmp instruction %s", p.lexer.CurrentPosition())
	case TOKEN_TFLOAT:
		ty = vm.TY_FLOAT64
		if err := p.lexer.ReadNextToken(); err != nil {
			return err
		}
		op1 = p.lexer.CurrentToken()
	}
	if op1.Ty != TOKEN_TREG {
		return fmt.Errorf("The first operand to the cmp instruction must be a valid register %s", p.lexer.CurrentPosition())
	}
	if op1.val.(int) > vm.GP_REG_MAX {
		return fmt.Errorf("Disallowed minuend register %s", p.lexer.CurrentPosition())
	}
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	if p.lexer.CurrentToken().Ty != TOKEN_TCOMMA {
		return fmt.Errorf("Instruction missing a comma %s", p.lexer.CurrentPosition())
	}
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	op2 := p.lexer.CurrentToken()

	switch op2.Ty {
	case TOKEN_TREG:
		if op2.val.(int) > vm.GP_REG_MAX {
			return fmt.Errorf("Disallowed subtrahend register %s", p.lexer.CurrentPosition())
		}
		p.currentInst = Instruction{
			Ty: INST_TCMPRR,
			Data: InstCmpData{
				Ty:  ty,
				Min: op1.val.(int),
				Sub: op2.val.(int),
			},
		}
	case TOKEN_TINTEGER_LIT:
		p.currentInst = Instruction{
			Ty: INST_TCMPIR,
			Data: InstCmpData{
				Ty:  ty,
				Min: op1.val.(int),
				Imm: op2.val.(uint64),
			},
		}
	case TOKEN_TFLOAT_LIT:
		p.currentInst = Instruction{
			Ty: INST_TCMPIR,
			Data: InstCmpData{
				Ty:  ty,
				Min: op1.val.(int),
				Imm: op2.val.(uint64),
			},
		}
	}

	return nil
}
func (p *Parser) parseJmp(ty int) error {
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	inst := Instruction{
		Ty: ty,
	}
	addr := p.lexer.CurrentToken()
	switch addr.Ty {
	case TOKEN_TINTEGER_LIT:
		inst.Data = InstJmpData {
			Address: addr.val.(uint64),
		}
	case TOKEN_TIDENT:
		inst.Data = InstJmpData {
			Address: addr.val.(string),
		}
	default:
		return fmt.Errorf("jg instruction requires a valid address as a parameter %s", p.lexer.CurrentPosition())
	}
	p.currentInst = inst
	return nil
}
