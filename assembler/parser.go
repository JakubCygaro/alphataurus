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
	INST_TJMP
	INST_TJMPE
	INST_TJMPZ
	INST_TJMPNE
	INST_TJMPNZ
	INST_TJMPG
	INST_TJMPGE
	INST_TJMPL
	INST_TJMPLE
	INST_TLABEL
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
	Address any
}
type InstLabData struct {
	Label string
	DeclaredAt string
}
type Instruction struct {
	Ty   int
	Data any
}
type Parser struct {
	lexer        Lexer
	currentInst  Instruction
	currentIdent string
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
		if err != nil {
			return false, err
		}
	default:
		return false, fmt.Errorf("Unimplemented instruction %s", p.lexer.CurrentPosition())
	}
	err = p.lexer.ReadNextToken()
	if p.lexer.CurrentToken().Ty != TOKEN_TNEWLINE && p.lexer.CurrentToken().Ty != TOKEN_TEOF {
		return false, fmt.Errorf("Extra tokens on line (%v) %s", p.lexer.CurrentToken(), p.lexer.CurrentPosition())
	}
	return true, err
}
func (p *Parser) parseStartIdent(t Token) error {
	ident := t.val.(string)
	p.currentIdent = ident
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
	case "jmp":
		return p.parseJmp(INST_TJMP)
	case "je":
		return p.parseJmp(INST_TJMPE)
	case "jne":
		return p.parseJmp(INST_TJMPNE)
	case "jz":
		return p.parseJmp(INST_TJMPZ)
	case "jnz":
		return p.parseJmp(INST_TJMPNZ)
	case "jg":
		return p.parseJmp(INST_TJMPG)
	case "jge":
		return p.parseJmp(INST_TJMPGE)
	case "jl":
		return p.parseJmp(INST_TJMPL)
	case "jle":
		return p.parseJmp(INST_TJMPLE)
	default:
		pos := p.lexer.CurrentPosition()
		if _, ok := p.lexer.Expect(TOKEN_TCOLON); ok {
			p.currentInst = Instruction{
				Ty: INST_TLABEL,
				Data: InstLabData{
					Label: ident,
					DeclaredAt: pos,
				},
			}
			return nil
		}
	}
	p.currentIdent = ""
	return fmt.Errorf("Unknown identifier '%s' %s", ident, p.lexer.CurrentPosition())
}
func (p *Parser) parseMov() error {
	var op1 Token
	if expr, err := p.parseExpression(0); err != nil {
		return err
	} else if eval, _ := TryConstEvaluateExpression(&expr); eval.Ty != CONSTEXPR_TREG {
		return fmt.Errorf("First operand to mov instruction must be a valid register %s", p.lexer.CurrentPosition())
	} else {
		op1.Ty = TOKEN_TREG
		op1.val = int(expr.Val.(ConstExpr).Val)
	}
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	comma := p.lexer.CurrentToken()
	if comma.Ty != TOKEN_TCOMMA {
		return fmt.Errorf("Instruction '%s' missing a comma %s", p.currentIdent, p.lexer.CurrentPosition())
	}

	var op2 ConstExpr
	if expr, err := p.parseExpression(0); err != nil {
		return err
	} else {
		eval, ok := TryConstEvaluateExpression(&expr)
		if !ok{
			return fmt.Errorf("Second operand to %s instruction has to be a valid register or a compile time expression %s",
				p.currentIdent,
				p.lexer.CurrentPosition())
		}
		op2 = eval
	}
	switch op2.Ty {
	case CONSTEXPR_TREG:
		p.currentInst = Instruction{
			Ty: INST_TMOVRR,
			Data: InstMovData{
				Src:  int(op2.Val),
				Dest: op1.val.(int),
			},
		}
	case CONSTEXPR_TILIT:
		p.currentInst = Instruction{
			Ty: INST_TMOVIR,
			Data: InstMovData{
				Imm:  op2.Val,
				Dest: op1.val.(int),
			},
		}
	case CONSTEXPR_TFLIT:
		p.currentInst = Instruction{
			Ty: INST_TMOVIR,
			Data: InstMovData{
				Imm:  op2.Val,
				Dest: op1.val.(int),
			},
		}
	default:
		return fmt.Errorf("Second operand to mov instruction must be a valid register or an immediate value %s",
			p.lexer.CurrentPosition())
	}
	return nil
}
func (p *Parser) prematureEndError() error {
	return fmt.Errorf("Premature end of input %s", p.lexer.CurrentPosition())
}
