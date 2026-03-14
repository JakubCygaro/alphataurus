package assembler

import (
	"bufio"
	"fmt"

	"github.com/JakubCygaro/alphataurus/assembler/errors"
	"github.com/JakubCygaro/alphataurus/internal/vm"
)

const (
	INST_TMOVRR = iota
	INST_TMOVIR
	INST_TMOVDRI
	INST_TMOVDRO1
	INST_TMOVDRO2
	INST_TMOVID
	INST_TMOVRD
	INST_TMOVRDO1
	INST_TMOVIDO1
	INST_TMOVRDO2
	INST_TMOVIDO2
	INST_TADDRR
	INST_TSUBRR
	INST_TDIVRR
	INST_TMULRR
	INST_TADDIR
	INST_TSUBIR
	INST_TINCR
	INST_TDECR
	INST_TNOT
	INST_TORIR
	INST_TORRR
	INST_TANDIR
	INST_TANDRR
	INST_TXORIR
	INST_TXORRR
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
	INST_TPUSHR
	INST_TPUSHI
	INST_TPOP
	INST_TCLR
	INST_TNOP
	INST_TCALL
	INST_TRET
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
const (
	LOG_TNOT = iota
	LOG_TAND
	LOG_TOR
	LOG_TXOR
)

type InstIncDecData struct {
	Reg int
}
type InstArthData struct {
	Source, Dest int
	Imm       uint64
	Ty        int
}
type InstLogicalData struct {
	First, Second int
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
	Label      string
	DeclaredAt string
}
type InstPushPopData struct {
	Reg uint64
	Imm uint64
}
type InstDerefMovData struct {
	Dest   int
	Offset int64
	OReg1  int
	OReg2  int
	Label  string
	OpTy   int
}
type InstMovDerefData struct {
	SourceReg int
	Imm       uint64
	Offset    int64
	OReg1     int
	OReg2     int
	Label     string
	OpTy      int
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
type InstCallData struct {
	Addr uint64
	Ident string
	Expr *Expr
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
		return false, errors.ExtraTokensOnLine(p.lexer.line, p.lexer.col)
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
	case "not":
		return p.parseLogical(LOG_TNOT)
	case "or":
		return p.parseLogical(LOG_TOR)
	case "and":
		return p.parseLogical(LOG_TAND)
	case "xor":
		return p.parseLogical(LOG_TXOR)
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
	case "push":
		return p.parsePush()
	case "pop":
		return p.parsePop()
	case "nop":
		p.currentInst = Instruction{
			Ty: INST_TNOP,
		}
		return nil
	case "clr":
		p.currentInst = Instruction{
			Ty: INST_TCLR,
		}
		return nil
	case "call":
		return p.parseCall()
	case "ret":
		p.currentInst = Instruction{
			Ty: INST_TRET,
		}
		return nil
	default:
		pos := p.lexer.CurrentPosition()
		if _, ok := p.lexer.Expect(TOKEN_TCOLON); ok {
			p.currentInst = Instruction{
				Ty: INST_TLABEL,
				Data: InstLabData{
					Label:      ident,
					DeclaredAt: pos,
				},
			}
			return nil
		}
	}
	p.currentIdent = ""
	return errors.UnknownIdentifier(ident, p.lexer.line, p.lexer.col)
}
