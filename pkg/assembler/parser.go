package assembler

import (
	"bufio"
	"fmt"
	"strings"
	"unicode"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
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
	INST_TLSHRR
	INST_TLSHIR
	INST_TRSHRR
	INST_TRSHIR
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
	INST_TJMPIP0R
	INST_TJMPIP1R
	INST_TCALLIP0R
	INST_TCALLIP1R
	INST_TLABEL
	INST_TPUSHR
	INST_TPUSHI
	INST_TPOP
	INST_TCLR
	INST_TNOP
	INST_TCALL
	INST_TRET
	INST_TSECCODE
	INST_TSECDATA
	INST_TIMPORT
	INST_TEXPORT
	INST_TATTRENTRY
	INST_TEXIT
)

const (
	ARTH_TUNSIGNED = vm.TY_UINT
	ARTH_TSIGNED   = vm.TY_SINT
	ARTH_TFLOAT    = vm.TY_FLOAT
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
	LOG_TLSH
	LOG_TRSH
)

type InstMovData struct {
	Src, Dest int
	Imm       uint64
	DataSize  byte
}
type InstIncDecData struct {
	Reg RegisterData
}
type InstArthData struct {
	Source, Dest RegisterData
	Imm          uint64
	Ty           int
	DataSize     byte
}
type InstLogicalData struct {
	First, Second RegisterData
	Imm           uint64
}
type InstCmpData struct {
	Ty       int
	Sub, Min RegisterData
	Imm      uint64
	ImmIsFloat bool
}
type InstJmpData struct {
	Address any
}
type InstJmpIPData struct {
	Reg    RegisterData
	Offset int64
	JmpTy  int
	OpTy   int
}
type InstCallIPData struct {
	Reg    RegisterData
	Offset int64
	OpTy   int
}
type InstLabData struct {
	Label      string
	DeclaredAt string
}
type InstPushPopData struct {
	Reg    uint64
	Imm    uint64
	DataSz byte
}
type InstDerefMovData struct {
	Dest   RegisterData
	Offset int64
	OReg1  RegisterData
	OReg2  RegisterData
	Label  string
	OpTy   int
}
type InstMovDerefData struct {
	DataSize  byte
	SourceReg RegisterData
	Imm       uint64
	Offset    int64
	OReg1     RegisterData
	OReg2     RegisterData
	Label     string
	OpTy      int
}
type Parser struct {
	lexer        Lexer
	currentInst  Instruction
	currentIdent string
}
type InstCallData struct {
	Addr  uint64
	Ident string
	Expr  *Expr
}
type InstImportExportData struct {
	Name string
	Weak bool
}
type InstExitData struct {
	Val uint64
}
type Instruction struct {
	Ty        int
	Data      any
	Line, Col uint64
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
	p.currentInst.Col, p.currentInst.Line = start.Col, start.Line
	switch start.Ty {
	case TOKEN_TIDENT:
		err = p.parseStartIdent(start)
		if err != nil {
			return false, err
		}
	case TOKEN_TAT:
		err = p.parseAttribute()
		if err != nil {
			return false, err
		}
	default:
		return false, fmt.Errorf("Unimplemented instruction %s", p.lexer.CurrentPosition())
	}
	err = p.lexer.ReadNextToken()
	if p.lexer.CurrentToken().Ty != TOKEN_TNEWLINE &&
		p.lexer.CurrentToken().Ty != TOKEN_TEOF {

		t := p.lexer.CurrentToken()
		return false, errors.ExtraTokensOnLine(t.Line, t.Col)
	}
	return true, err
}
func (p *Parser) parseStartIdent(t Token) error {
	ident := t.Val.(string)
	p.currentIdent = ident

	if err := p.lexer.ReadNextToken(); err != nil {
		p.lexer.UnreadToken()
	} else if next := p.lexer.CurrentToken(); next.Ty != TOKEN_TCOLON {
		p.lexer.UnreadToken()
	} else {
		p.currentInst = Instruction{
			Ty: INST_TLABEL,
			Data: InstLabData{
				Label: ident,
			},
		}
		return nil
	}

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
	case "lsh":
		return p.parseLogical(LOG_TLSH)
	case "rsh":
		return p.parseLogical(LOG_TRSH)
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
	case "section":
		return p.parseSection()
	case "export":
		return p.parseExport()
	case "import":
		return p.parseImport()
	case "exit":
		return p.parseExit()
	}
	p.currentIdent = ""
	return errors.UnknownIdentifier(ident, t.Line, t.Col)
}
func (p *Parser) parseSection() error {
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	op := p.lexer.CurrentToken()
	switch op.Ty {
	case TOKEN_TSINGLEQ:
		ty := op.Val.(string)
		switch ty {
		case ".code":
			p.currentInst = Instruction{
				Ty: INST_TSECCODE,
			}
		default:
			return errors.FailedToParse("section",
				fmt.Sprintf("Unknown section type '%s'", ty), op.Line, op.Col)
		}
	default:
		return errors.FailedToParse("section",
			"Bad argument", op.Line, op.Col)
	}
	return nil
}
func (p *Parser) parseImport() error {
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	weak := false
	op := p.lexer.CurrentToken()
	if op.Ty == TOKEN_TWEAK {
		weak = true
		if err := p.lexer.ReadNextToken(); err != nil {
			return err
		}
		op = p.lexer.CurrentToken()
	}
	if op.Ty != TOKEN_TSINGLEQ {
		return errors.FailedToParse("import statement", "expected single quoted string parameter",
			op.Line, op.Col)
	}
	name := op.Val.(string)
	if strings.ContainsFunc(name, unicode.IsSpace) {
		return errors.FailedToParse("import statement", "parameter not a valid identifier",
			op.Line, op.Col)
	}
	p.currentInst = Instruction{
		Ty: INST_TIMPORT,
		Data: InstImportExportData{
			Name: name,
			Weak: weak,
		},
		Line: op.Line,
		Col:  op.Col,
	}
	return nil
}
func (p *Parser) parseExport() error {
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	op := p.lexer.CurrentToken()
	if op.Ty != TOKEN_TSINGLEQ {
		return errors.FailedToParse("export statement", "expected single quoted string parameter",
			op.Line, op.Col)
	}
	name := op.Val.(string)
	if strings.ContainsFunc(name, unicode.IsSpace) {
		return errors.FailedToParse("export statement", "parameter not a valid identifier",
			op.Line, op.Col)
	}
	p.currentInst = Instruction{
		Ty: INST_TEXPORT,
		Data: InstImportExportData{
			Name: name,
		},
		Line: op.Line,
		Col:  op.Col,
	}
	return nil
}
func (p *Parser) parseAttribute() error {
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	op := p.lexer.CurrentToken()
	if op.Ty != TOKEN_TIDENT {
		return errors.FailedToParse("attribute", "invalid parameter",
			op.Line, op.Col)
	}
	switch op.Val.(string) {
	case "entry":
		p.currentInst = Instruction{
			Ty:   INST_TATTRENTRY,
			Data: nil,
		}
	default:
		return errors.FailedToParse("attribute", "unrecognized attribute type",
			op.Line, op.Col)
	}
	return nil
}
func (p *Parser) parseExit() error {
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	start := p.lexer.CurrentToken()
	p.lexer.UnreadToken()
	expr, err := p.parseExpression(0)
	if err != nil {
		return err
	}
	if cexpr, ok := TryConstEvaluatePruneExpression(expr); !ok {
		return errors.FailedToParse("exit instruction", "non comp-time expression parameter",
			start.Line, start.Col)
	} else if cexpr.Ty != CONSTEXPR_TILIT {
		return errors.FailedToParse("exit instruction", "invalid expression value type",
			start.Line, start.Col)
	} else {
		p.currentInst = Instruction{
			Ty: INST_TEXIT,
			Data: InstExitData{
				Val: cexpr.Val,
			},
		}
	}
	return nil
}
