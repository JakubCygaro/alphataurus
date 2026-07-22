package assembler

import (
	"bufio"
	"strings"
	"unicode"

	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
	"github.com/JakubCygaro/alphataurus/pkg/assembler/parser/errors"
)

type ParserWarningData struct {
	Col, Line int
	Message   string
}
type Parser struct {
	lexer             *lx.Lexer
	currentInst       Instruction
	currentIdent      string
	currentStartToken lx.Token
	// pointer to a function that recieves warnings emitted by the parser
	WarningSink func(ParserWarningData)
	// when parsing instructions of Generic...Inst variant immediately attempt
	// to coalesce them into proper non-generic Instructions, this makes it impossible
	// for the assembler to later reparse them after expression evaluations
	ForceCoalesceGenerics bool
}

func pInitialState() Parser {
	return Parser{}
}

func NewParser(reader *bufio.Reader) *Parser {
	this := pInitialState()
	this.lexer = lx.NewLexer(reader)
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
		case lx.TOKEN_TNEWLINE:
			p.lexer.UnreadCurrentToken()
		case lx.TOKEN_TEOF:
			p.lexer.UnreadCurrentToken()
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
		if p.currentStartToken.Ty == lx.TOKEN_TEOF {
			return false, nil
		}
		if p.currentStartToken.Ty == lx.TOKEN_TDOUBLESEMICOLON {
			err = p.SkipCommentLine()
		} else if p.currentStartToken.Ty != lx.TOKEN_TNEWLINE {
			break
		}
	}
	switch p.currentStartToken.Ty {
	case lx.TOKEN_TIDENT:
		err = p.parseStartIdent(p.currentStartToken)
		if err != nil {
			return false, err
		}
	case lx.TOKEN_TAT:
		err = p.ParseAttribute()
		if err != nil {
			return false, err
		}
	default:
		return false, errors.
			ExtraTokensOnLine(
				p.currentStartToken.Line, p.currentStartToken.Col,
				p.currentStartToken,
			)
	}
	err = p.lexer.ReadNextToken()
	if p.lexer.CurrentToken().Ty == lx.TOKEN_TDOUBLESEMICOLON {
		err = p.SkipCommentLine()
	}
	if p.lexer.CurrentToken().Ty != lx.TOKEN_TNEWLINE &&
		p.lexer.CurrentToken().Ty != lx.TOKEN_TEOF {

		t := p.lexer.CurrentToken()
		return false,
			errors.
				ExtraTokensOnLine(
					t.Line, t.Col,
					t,
				)
	}
	p.currentInst.Col, p.currentInst.Line =
		p.currentStartToken.Col, p.currentStartToken.Line
	p.currentStartToken = lx.Token{}
	return true, err
}

func (p *Parser) parseStartIdent(t lx.Token) error {
	ident := t.Val.(string)
	p.currentIdent = ident

	if err := p.lexer.ReadNextToken(); err != nil {
		p.lexer.UnreadCurrentToken()
	} else if next := p.lexer.CurrentToken(); next.Ty != lx.TOKEN_TCOLON {
		p.lexer.UnreadCurrentToken()
	} else {
		p.currentInst = Instruction{
			Data: InstLab{
				LabelName: ident,
			},
		}
		return nil
	}

	switch ident {
	case "mov":
		return p.parseMov()
	case "movsx":
		return p.parseSXMov()
	case "movzx":
		return p.parseZXMov()
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
		return p.parseJmp(JMP)
	case "je":
		return p.parseJmp(JMPE)
	case "jne":
		return p.parseJmp(JMPNE)
	case "jz":
		return p.parseJmp(JMPZ)
	case "jnz":
		return p.parseJmp(JMPNZ)
	case "jg":
		return p.parseJmp(JMPG)
	case "jge":
		return p.parseJmp(JMPGE)
	case "jl":
		return p.parseJmp(JMPL)
	case "jle":
		return p.parseJmp(JMPLE)
	case "js":
		return p.parseJmp(JMPS)
	case "jns":
		return p.parseJmp(JMPNS)
	case "jc":
		return p.parseJmp(JMPC)
	case "jnc":
		return p.parseJmp(JMPNC)
	case "jo":
		return p.parseJmp(JMPO)
	case "jno":
		return p.parseJmp(JMPNO)
	case "push":
		return p.parsePush()
	case "pop":
		return p.parsePop()
	case "nop":
		p.currentInst = Instruction{
			Data: InstNop{},
		}
		return nil
	case "clr":
		p.currentInst = Instruction{
			Data: InstClr{},
		}
		return nil
	case "call":
		return p.parseCall()
	case "ret":
		p.currentInst = Instruction{
			Data: InstRet{},
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
	return errors.UnknownIdentifier(t.Line, t.Col, ident)
}

func (p *Parser) parseSection() error {
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	op := p.lexer.CurrentToken()
	switch op.Ty {
	case lx.TOKEN_TSINGLEQ:
		ty := op.Val.(string)
		switch ty {
		case ".code":
			p.currentInst = Instruction{
				Data: InstSecCode{},
			}
		default:
			return errors.
				MakeParserError(
					op.Line, op.Col,
					"Unknown section name `%s`.",
					ty,
				)
		}
	default:
		return errors.
			MakeParserError(
				op.Line, op.Col,
				"Bad section type argument `%s`.",
				op.ForceValAsString(),
			)
	}
	return nil
}
func (p *Parser) parseImport() error {
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	weak := false
	op := p.lexer.CurrentToken()
	if op.Ty == lx.TOKEN_TWEAK {
		weak = true
		if err := p.lexer.ReadNextToken(); err != nil {
			return err
		}
		op = p.lexer.CurrentToken()
	}
	if op.Ty != lx.TOKEN_TSINGLEQ {
		return errors.
			MakeParserError(
				op.Line, op.Col,
				"Expected a single quoted string parameter, got `%s`.",
				op.ForceValAsString(),
			)
	}
	name := op.Val.(string)
	if strings.ContainsFunc(name, unicode.IsSpace) {
		return errors.
			MakeParserError(
				op.Line, op.Col,
				"`%s` is not a valid identifier",
				name,
			)
	}
	p.currentInst = Instruction{
		Data: InstImport{
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
	if op.Ty != lx.TOKEN_TSINGLEQ {
		return errors.
			MakeParserError(
				op.Line, op.Col,
				"Expected a single quoted string parameter, got `%s`.",
				op.ForceValAsString(),
			)
	}
	name := op.Val.(string)
	if strings.ContainsFunc(name, unicode.IsSpace) {
		return errors.
			MakeParserError(
				op.Line, op.Col,
				"`%s` is not a valid identifier",
				name,
			)
	}
	p.currentInst = Instruction{
		Data: InstExport{
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
	if op.Ty != lx.TOKEN_TIDENT {
		return errors.
			MakeParserError(
				op.Line, op.Col,
				"`%s` is not a valid identifier",
				op.ForceValAsString(),
			)
	}
	attr := op.Val.(string)
	switch attr {
	case "entry":
		p.currentInst = Instruction{
			Data: InstEntry{},
		}
	default:
		return errors.
			MakeParserError(
				op.Line, op.Col,
				"Unrecognized attribute `%s`.",
				attr,
			)
	}
	return nil
}
func (p *Parser) parseExit() error {
	expr, err := p.ParseExpression()
	if err != nil {
		return err
	}
	if v, ok := IsConstexprType[ConstExprILit](expr); ok {
		p.currentInst = Instruction{
			Data: InstExitI{
				Val: v.Integer,
			},
		}
	} else if expr.IsRegexpr() {
		p.currentInst = Instruction{
			Data: InstExitR{
				Reg: expr.Val.(RegExpr).Reg,
			},
		}
	} else {
		p.currentInst = Instruction{
			Data: InstGenericExit{
				Expr: expr,
			},
		}
	}
	return nil
}
