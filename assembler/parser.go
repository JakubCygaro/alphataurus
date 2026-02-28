package assembler

import (
	"bufio"
	"fmt"

	"github.com/JakubCygaro/alphataurus/internal/vm"
)

const (
	INST_TMOVRR = iota
	INST_TMOVIR
	INST_TADDRR
	INST_TSUBRR
	INST_TDIVRR
	INST_TMULRR
	INST_TADDIR
	INST_TSUBIR
	INST_TINCR
)

type InstMovData struct {
	Src, Dest int
	Imm       uint64
}

const (
	ARTH_TUNSIGNED = vm.TY_UINT64
	ARTH_TSIGNED   = vm.TY_INT64
	ARTH_TFLOAT    = vm.TY_FLOAT64
)

const (
	ARTH_TADD = iota
	ARTH_TSUB
	ARTH_TDIV
	ARTH_TMUL
)

type InstIncData struct {
	Reg int
}

type InstArthData struct {
	Src, Dest int
	Ty        int
	Imm       uint64
}

type Instruction struct {
	Ty   int
	Data any
}
type Parser struct {
	lexer       Lexer
	currentInst Instruction
}

func NewParser(reader bufio.Reader) Parser {
	return Parser{
		lexer: NewLexer(reader),
	}
}

func (p *Parser) CurrentInst() Instruction {
	return p.currentInst
}
func (p *Parser) ParseNext() (bool, error) {
	var err error = nil
	var start Token
	for {
		err = p.lexer.ReadNextToken()
		if err != nil {
			return false, err
		}
		start = p.lexer.CurrentToken()
		if start.Ty == TOKEN_TEOF {
			return false, nil
		}
		if start.Ty != TOKEN_TNEWLINE {
			break
		}
	}
	switch start.Ty {
	case TOKEN_TIDENT:
		err = p.parseStartIdent(start)
	default:
		return false, fmt.Errorf("Unimplemented instruction %s", p.lexer.CurrentPosition())
	}
	err = p.lexer.ReadNextToken()
	if p.lexer.CurrentToken().Ty != TOKEN_TNEWLINE && p.lexer.CurrentToken().Ty != TOKEN_TEOF {
		return false, fmt.Errorf("Extra tokens on line %s", p.lexer.CurrentPosition())
	}
	return true, err
}
func (p *Parser) parseStartIdent(t Token) error {
	ident := t.val.(string)
	switch ident {
	case "mov":
		return p.parseMov()
	case "add":
		return p.parseAddOrSub(ARTH_TADD)
	case "sub":
		return p.parseAddOrSub(ARTH_TSUB)
	case "div":
		return p.parseDivOrMul(ARTH_TDIV)
	case "mul":
		return p.parseDivOrMul(ARTH_TMUL)
	case "inc":
		return p.parseInc()
	}
	return fmt.Errorf("Unknown identifier '%s' %s", ident, p.lexer.CurrentPosition())
}
func (p *Parser) parseMov() error {
	err := p.lexer.ReadNextToken()
	if err != nil {
		return err
	}
	op1 := p.lexer.CurrentToken()
	if op1.Ty == TOKEN_TEOF {
		return p.prematureEndError()
	}
	if op1.Ty != TOKEN_TREG {
		return fmt.Errorf("First operand to mov instruction must be a valid register %s", p.lexer.CurrentPosition())
	}
	err = p.lexer.ReadNextToken()
	if err != nil {
		return err
	}
	comma := p.lexer.CurrentToken()

	if comma.Ty != TOKEN_TCOMMA {
		return fmt.Errorf("mov instruction missing a comma %s", p.lexer.CurrentPosition())
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
		p.currentInst = Instruction{
			Ty: INST_TMOVRR,
			Data: InstMovData{
				Src:  op2.val.(int),
				Dest: op1.val.(int),
			},
		}
	case TOKEN_TINTEGER_LIT:
		p.currentInst = Instruction{
			Ty: INST_TMOVIR,
			Data: InstMovData{
				Imm:  op2.val.(uint64),
				Dest: op1.val.(int),
			},
		}
	case TOKEN_TFLOAT_LIT:
		p.currentInst = Instruction{
			Ty: INST_TMOVIR,
			Data: InstMovData{
				Imm:  op2.val.(uint64),
				Dest: op1.val.(int),
			},
		}
	default:
		return fmt.Errorf("Second operand to mov instruction must be a valid register or an immediate value %s", p.lexer.CurrentPosition())
	}
	return nil
}
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
		return fmt.Errorf("First operand to add instruction must be a valid register %s", p.lexer.CurrentPosition())
	}

	err = p.lexer.ReadNextToken()
	if err != nil {
		return err
	}
	comma := p.lexer.CurrentToken()

	if comma.Ty != TOKEN_TCOMMA {
		return fmt.Errorf("add instruction missing a comma %s", p.lexer.CurrentPosition())
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
	p.currentInst = Instruction{
		Ty: INST_TINCR,
		Data: InstIncData{
			Reg: op1.val.(int),
		},
	}
	return nil
}
func (p *Parser) prematureEndError() error {
	return fmt.Errorf("Premature end of input %s", p.lexer.CurrentPosition())
}
