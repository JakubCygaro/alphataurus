package assembler

import (
	"bufio"
	"strings"
	"unicode"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
)

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
			Data: InstLab{
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
				Data: InstSecCode{},
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
	if op.Ty != TOKEN_TIDENT {
		return errors.FailedToParse("attribute", op.Line, op.Col,
			"`%s` is not a valid parameter", op.ForceValAsString())
	}
	attr := op.Val.(string)
	switch attr {
	case "entry":
		p.currentInst = Instruction{
			Data: InstEntry{},
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
			Data: InstExitI{
				Val: cexpr.Val,
			},
		}
	} else if cexpr.Ty == CONSTEXPR_TREG {
		p.currentInst = Instruction{
			Data: InstExitR{
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
