package assembler

import (
	"fmt"

	"github.com/JakubCygaro/alphataurus/internal/vm"
)

func (p *Parser) parseAddOrSub(arthTy int) error {
	err := p.lexer.ReadNextToken()
	if err != nil {
		return err
	}
	op1 := p.lexer.CurrentToken()
	valTy := ARTH_TUNSIGNED
	if op1.Ty == TOKEN_TSIGNED || op1.Ty == TOKEN_TFLOAT || op1.Ty == TOKEN_TUNSIGNED {
		switch op1.Ty {
		case TOKEN_TSIGNED:
			valTy = ARTH_TSIGNED
		case TOKEN_TUNSIGNED:
			valTy = ARTH_TUNSIGNED
		case TOKEN_TFLOAT:
			valTy = ARTH_TFLOAT
		}
		err = p.lexer.ReadNextToken()
		if err != nil {
			return err
		}
		op1 = p.lexer.CurrentToken()
	}
	if op1.Ty == TOKEN_TEOF {
		return p.prematureEndError()
	}
	if op1.Ty != TOKEN_TREG {
		return fmt.Errorf("First operand to instruction must be a valid register %s", p.lexer.CurrentPosition())
	}
	switch op1.val.(int) {
	case vm.IP_IDX:
		return fmt.Errorf("Disallowed source register %s", p.lexer.CurrentPosition())
	}

	err = p.lexer.ReadNextToken()
	if err != nil {
		return err
	}
	comma := p.lexer.CurrentToken()

	if comma.Ty != TOKEN_TCOMMA {
		return fmt.Errorf("Instruction missing a comma %s", p.lexer.CurrentPosition())
	}
	err = p.lexer.ReadNextToken()
	if err != nil {
		return err
	}
	op2 := p.lexer.CurrentToken()
	if op2.Ty == TOKEN_TEOF {
		return p.prematureEndError()
	}
	switch op2.Ty {
	case TOKEN_TREG:
		switch op2.val.(int) {
		case vm.IP_IDX:
			return fmt.Errorf("Disallowed destination register %s", p.lexer.CurrentPosition())
		}
		var ty int
		switch arthTy {
		case ARTH_TADD:
			ty = INST_TADDRR
		case ARTH_TSUB:
			ty = INST_TSUBRR
		}
		p.currentInst = Instruction{
			Ty: ty,
			Data: InstArthData{
				Src:  op2.val.(int),
				Dest: op1.val.(int),
				Ty:   valTy,
			},
		}
	case TOKEN_TINTEGER_LIT:
		var ty int
		switch arthTy {
		case ARTH_TADD:
			ty = INST_TADDIR
		case ARTH_TSUB:
			ty = INST_TSUBIR
		}
		p.currentInst = Instruction{
			Ty: ty,
			Data: InstArthData{
				Imm:  op2.val.(uint64),
				Dest: op1.val.(int),
				Ty:   valTy,
			},
		}
	case TOKEN_TFLOAT_LIT:
		var ty int
		switch arthTy {
		case ARTH_TADD:
			ty = INST_TADDIR
		case ARTH_TSUB:
			ty = INST_TSUBIR
		}
		p.currentInst = Instruction{
			Ty: ty,
			Data: InstArthData{
				Imm:  op2.val.(uint64),
				Dest: op1.val.(int),
				Ty:   valTy,
			},
		}
	default:
		return fmt.Errorf("Second operand to add instruction must be a valid register or an immediate value %s", p.lexer.CurrentPosition())
	}
	return nil
}
func (p *Parser) parseDivOrMul(arthTy int) error {
	err := p.lexer.ReadNextToken()
	if err != nil {
		return err
	}
	op1 := p.lexer.CurrentToken()
	valTy := ARTH_TUNSIGNED
	switch op1.Ty {
	case TOKEN_TSIGNED:
		valTy = ARTH_TSIGNED
	case TOKEN_TUNSIGNED:
		valTy = ARTH_TUNSIGNED
	case TOKEN_TFLOAT:
		valTy = ARTH_TFLOAT
	case TOKEN_TNEWLINE:
		p.lexer.UnreadToken()
	case TOKEN_TEOF:
		p.lexer.UnreadToken()
	default:
		var opName string
		switch arthTy {
		case ARTH_TDIV:
			opName = "div"
		case ARTH_TMUL:
			opName = "mul"
		}
		return fmt.Errorf("Invalid %s instruction %s", opName, p.lexer.CurrentPosition())
	}
	var ty int
	switch arthTy {
	case ARTH_TDIV:
		ty = INST_TDIVRR
	case ARTH_TMUL:
		ty = INST_TMULRR
	}
	p.currentInst = Instruction{
		Ty: ty,
		Data: InstArthData{
			Ty: valTy,
		},
	}
	return nil
}
func (p *Parser) parseInc() error {
	err := p.lexer.ReadNextToken()
	if err != nil {
		return err
	}
	op1 := p.lexer.CurrentToken()
	if op1.Ty == TOKEN_TEOF {
		return p.prematureEndError()
	}
	if op1.Ty != TOKEN_TREG {
		return fmt.Errorf("The operand to the inc instruction must be a valid register %s", p.lexer.CurrentPosition())
	}
	switch op1.val.(int) {
	case vm.IP_IDX:
		return fmt.Errorf("Disallowed register %s", p.lexer.CurrentPosition())
	}
	p.currentInst = Instruction{
		Ty: INST_TINCR,
		Data: InstIncDecData{
			Reg: op1.val.(int),
		},
	}
	return nil
}
func (p *Parser) parseDec() error {
	err := p.lexer.ReadNextToken()
	if err != nil {
		return err
	}
	op1 := p.lexer.CurrentToken()
	if op1.Ty == TOKEN_TEOF {
		return p.prematureEndError()
	}
	if op1.Ty != TOKEN_TREG {
		return fmt.Errorf("The operand to the dec instruction must be a valid register %s", p.lexer.CurrentPosition())
	}
	switch op1.val.(int) {
	case vm.IP_IDX:
		return fmt.Errorf("Disallowed register %s", p.lexer.CurrentPosition())
	}
	p.currentInst = Instruction{
		Ty: INST_TDECR,
		Data: InstIncDecData{
			Reg: op1.val.(int),
		},
	}
	return nil
}
