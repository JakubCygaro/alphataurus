package assembler

import (
	"bufio"
	"strings"
	"unicode"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

//go:generate stringer -type=InstTy
type InstTy int

const (
	INST_TMOVRR InstTy = iota
	INST_TMOVIR
	INST_TMOVDRI
	INST_TMOVDRO0
	INST_TMOVDRO1
	INST_TMOVDRO2
	INST_TMOVID
	INST_TMOVRD
	INST_TMOVRDO1
	INST_TMOVIDO1
	INST_TMOVIDO1_NO
	INST_TMOVRDO2
	INST_TMOVIDO2
	INST_TMOVIDO2_NO
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
	INST_TJMPS
	INST_TJMPNS
	INST_TJMPC
	INST_TJMPNC
	INST_TJMPO
	INST_TJMPNO
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
	INST_TEXITI
	INST_TEXITR
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
	Src, Dest RegisterData
	Imm       uint64
	Ty        int
	DataSize  byte
}
type InstLogicalData struct {
	First, Second RegisterData
	Imm           uint64
}
type InstCmpData struct {
	Ty       int
	Sub, Min RegisterData
	Imm      uint64
}
type InstJmpData struct {
	Address  any
	Absolute bool
}
type InstJmpIPData struct {
	Reg    RegisterData
	Offset int64
	JmpTy  InstTy
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
	DataSize byte
	Src      RegisterData
	Imm      uint64
	Offset   int64
	OReg1    RegisterData
	OReg2    RegisterData
	Label    string
	OpTy     int
}
type ParserWarningData struct {
	Col, Line int
	Message   string
}
type Parser struct {
	lexer             *Lexer
	currentInst       Instruction
	currentIdent      string
	currentStartToken Token
	// pointer to a function that recieves warnings emitted by the parser
	WarningSink func(ParserWarningData)
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
	Reg RegisterData
}
type Instruction struct {
	Ty        InstTy
	Data      any
	Line, Col int
}

func pInitialState() Parser {
	return Parser{}
}

func NewParser(reader *bufio.Reader) *Parser {
	this := pInitialState()
	this.lexer = NewLexer(reader)
	return &this
}
// reset the state of the parser and load new reader input
func (p *Parser) LoadNew(reader *bufio.Reader) {
	ws := p.WarningSink
	l := p.lexer
	l.LoadNew(reader)
	*p = pInitialState()
	p.WarningSink = ws
	p.lexer = l
}

func (p *Parser) CurrentInst() Instruction {
	return p.currentInst
}
func (p *Parser) SkipCommentLine() error {
	for {
		if err := p.lexer.ReadNextToken(); err != nil {
			return err
		}
		switch p.lexer.CurrentToken().Ty {
		case TOKEN_TNEWLINE:
			p.lexer.UnreadToken()
		case TOKEN_TEOF:
			p.lexer.UnreadToken()
		default:
			continue
		}
		break
	}
	return nil
}
func (p *Parser) ParseNext() (bool, error) {
	var err error = nil
	for {
		err = p.lexer.ReadNextToken()
		if err != nil {
			return false, err
		}
		p.currentStartToken = p.lexer.CurrentToken()
		if p.currentStartToken.Ty == TOKEN_TEOF {
			return false, nil
		}
		if p.currentStartToken.Ty == TOKEN_TDOUBLESEMICOLON {
			err = p.SkipCommentLine()
		} else if p.currentStartToken.Ty != TOKEN_TNEWLINE {
			break
		}
	}
	switch p.currentStartToken.Ty {
	case TOKEN_TIDENT:
		err = p.parseStartIdent(p.currentStartToken)
		if err != nil {
			return false, err
		}
	case TOKEN_TAT:
		err = p.ParseAttribute()
		if err != nil {
			return false, err
		}
	default:
		return false, errors.
			ExtraTokensOnLine(p.currentStartToken.Line, p.currentStartToken.Col,
				p.currentStartToken.ForceValAsString())
	}
	err = p.lexer.ReadNextToken()
	if p.lexer.CurrentToken().Ty == TOKEN_TDOUBLESEMICOLON {
		err = p.SkipCommentLine()
	}
	if p.lexer.CurrentToken().Ty != TOKEN_TNEWLINE &&
		p.lexer.CurrentToken().Ty != TOKEN_TEOF {

		t := p.lexer.CurrentToken()
		return false, errors.ExtraTokensOnLine(t.Line, t.Col, t.ForceValAsString())
	}
	p.currentInst.Col, p.currentInst.Line =
		p.currentStartToken.Col, p.currentStartToken.Line
	p.currentStartToken = Token{}
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
	case "js":
		return p.parseJmp(INST_TJMPS)
	case "jns":
		return p.parseJmp(INST_TJMPNS)
	case "jc":
		return p.parseJmp(INST_TJMPC)
	case "jnc":
		return p.parseJmp(INST_TJMPNC)
	case "jo":
		return p.parseJmp(INST_TJMPO)
	case "jno":
		return p.parseJmp(INST_TJMPNO)
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
			return errors.FailedToParse(p.currentIdent, op.Line, op.Col,
				"Unknown section name `%s`", ty)
		}
	default:
		return errors.FailedToParse(p.currentIdent, op.Line, op.Col,
			"Bad section type argument `%s`", op.ForceValAsString())
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
		return errors.FailedToParse(p.currentIdent, op.Line, op.Col,
			"Expected a single quoted string parameter, got `%s`",
			op.ForceValAsString())
	}
	name := op.Val.(string)
	if strings.ContainsFunc(name, unicode.IsSpace) {
		return errors.FailedToParse(p.currentIdent,
			op.Line, op.Col,
			"`%s` is not a valid identifier",
			name,
		)
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
		return errors.FailedToParse(p.currentIdent, op.Line, op.Col,
			"Expected a single quoted string parameter, got `%s`",
			op.ForceValAsString())
	}
	name := op.Val.(string)
	if strings.ContainsFunc(name, unicode.IsSpace) {
		return errors.FailedToParse(p.currentIdent, op.Line, op.Col,
			"`%s` is not a valid identifier", name)
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
func (p *Parser) ParseAttribute() error {
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	op := p.lexer.CurrentToken()
	if op.Ty != TOKEN_TIDENT {
		return errors.FailedToParse("attribute", op.Line, op.Col,
			"`%s` is not a valid parameter", op.ForceValAsString())
	}
	attr := op.Val.(string)
	switch attr {
	case "entry":
		p.currentInst = Instruction{
			Ty:   INST_TATTRENTRY,
			Data: nil,
		}
	default:
		return errors.FailedToParse("attribute", op.Line, op.Col,
			"Unrecognized attribute type `%s`", attr)
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
		em, _ := cexpr.Emit()
		return errors.FailedToParse(p.currentIdent, start.Line, start.Col,
			"Non comp-time expression as parameter `%s`."+
				"\nThe argument to this instruction must be either a valid register "+
				"or a compile time expression.", em)
	} else if cexpr.Ty == CONSTEXPR_TILIT {
		p.currentInst = Instruction{
			Ty: INST_TEXITI,
			Data: InstExitData{
				Val: cexpr.Val,
			},
		}
	} else if cexpr.Ty == CONSTEXPR_TREG {
		p.currentInst = Instruction{
			Ty: INST_TEXITR,
			Data: InstExitData{
				Reg: cexpr.UnpackAsRegisterData(),
			},
		}
	} else {
		em, _ := cexpr.Emit()
		return errors.FailedToParse(p.currentIdent, start.Line, start.Col,
			"Invalid expression as parameter `%s`."+
				"\nThe argument to this instruction must be either a valid register "+
				"or a compile time expression.", em)
	}
	return nil
}
