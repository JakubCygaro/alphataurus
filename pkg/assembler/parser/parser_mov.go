package assembler

import (
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
	"github.com/JakubCygaro/alphataurus/pkg/assembler/parser/errors"
)

func (p *Parser) parseMov() error {
	genericMov := InstGenericMov{}
	var op1 lx.Token
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	op1 = p.lexer.CurrentToken()
	if sz, ok := lx.TokenAsSize(&op1); !ok {
		p.lexer.UnreadCurrentToken()
	} else {
		genericMov.DataSize = &DataSize{
			Size: sz,
			Col:  op1.Col,
			Line: op1.Line,
		}
		op1 = lx.Token{}
	}
	if expr, err := p.ParseExpression(); err != nil {
		return err
	} else {
		genericMov.Dest = expr
	}
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	comma := p.lexer.CurrentToken()
	if comma.Ty != lx.TOKEN_TCOMMA {
		return errors.MissingComma(
			comma.Line, comma.Col,
			comma,
		)
	}
	if expr, err := p.ParseExpression(); err != nil {
		return err
	} else {
		genericMov.Src = expr
	}
	if concrete, err :=
		GetConcreteMovInst(genericMov, &p.currentInst); concrete != nil && err == nil {
		p.currentInst = Instruction{
			Data: concrete,
		}
	} else if err != nil && p.ForceCoalesceGenerics {
		return err
	} else {
		p.currentInst = Instruction{
			Data: genericMov,
		}
	}
	return nil
}
func (p *Parser) parseSXMov() error {
	if err := p.parseMov(); err != nil {
		return err
	}
	if movRR, ok := p.currentInst.Data.(InstMovRR); !ok {
		return errors.MakeParserError(
			p.currentInst.Line,
			p.currentInst.Col,
			"Sign extend move allowed only for register to register data move.",
		)
	} else {
		p.currentInst.Data = InstMovSXRR{
			Src:  movRR.Src,
			Dest: movRR.Dest,
		}
	}
	return nil
}
