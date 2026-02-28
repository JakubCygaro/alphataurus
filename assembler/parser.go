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
	INST_TDECR
	INST_TCMPRR
	INST_TCMPIR
	INST_TJMPG
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

type InstIncDecData struct {
	Reg int
}
type InstArthData struct {
	Src, Dest int
	Ty        int
	Imm       uint64
}
type InstCmpData struct {
	Ty       int
	Sub, Min int
	Imm      uint64
}
type InstJmpData struct {
	Address uint64
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
	case "dec":
		return p.parseDec()
	case "cmp":
		return p.parseCmp()
	case "jg":
		return p.parseJmpG()
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
func (p *Parser) prematureEndError() error {
	return fmt.Errorf("Premature end of input %s", p.lexer.CurrentPosition())
}
