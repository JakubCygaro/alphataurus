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
			Mov: movRR,
		}
	}
	return nil
}
func (p *Parser) parseZXMov() error {
	if err := p.parseMov(); err != nil {
		return err
	}
	switch mov := p.currentInst.Data.(type) {
	case InstMovRR:
		p.currentInst.Data = InstMovZXRR{
			Mov: mov,
		}
	case InstMovIR:
		p.currentInst.Data = InstMovZXIR{
			Mov: mov,
		}
	case InstMovDR:
		p.currentInst.Data = InstMovZXDR{
			Mov: mov,
		}
	case InstMovDRO1:
		p.currentInst.Data = InstMovZXDRO1{
			Mov: mov,
		}
	case InstMovDRO2:
		p.currentInst.Data = InstMovZXDRO2{
			Mov: mov,
		}
	default:
		return errors.MakeParserError(
			p.currentInst.Line,
			p.currentInst.Col,
			"Zero extend move not allowed for this type of data move.",
		)
	}
	return nil
}
func (p *Parser) parseXChg() error {
	if err := p.parseMov(); err != nil {
		return err
	}
	helpErr := func() error {
		return errors.
			MakeParserError(
				p.currentInst.Line,
				p.currentInst.Col,
				"Exchange with deference only allowed when the address is the source",
			)
	}
	switch mov := p.currentInst.Data.(type) {
	case InstMovRR:
		if mov.Src.Size != mov.Dest.Size {
			return errors.
				MakeParserError(
					p.currentInst.Line,
					p.currentInst.Col,
					"Exchange between registers of different sizes is not allowed",
				)
		}
		p.currentInst.Data = InstXCHGRR{
			Mov: mov,
		}
	case InstMovDR:
		p.currentInst.Data = InstXCHGDR{
			Mov: mov,
		}
	case InstMovRD:
		return helpErr()
	case InstMovDRO1:
		p.currentInst.Data = InstXCHGDRO1{
			Mov: mov,
		}
	case InstMovDRO2:
		p.currentInst.Data = InstXCHGDRO2{
			Mov: mov,
		}
	case InstMovRDO1:
		return helpErr()
	case InstMovRDO2:
		return helpErr()
	default:
		return errors.MakeParserError(
			p.currentInst.Line,
			p.currentInst.Col,
			"Exchange not allowed for this type of data move.",
		)
	}
	return nil
}
func (p *Parser) parseLea() error {
	if err := p.parseMov(); err != nil {
		return err
	}
	switch mov := p.currentInst.Data.(type) {
	case InstMovRR:
		p.currentInst.Data = InstMovZXRR{
			Mov: mov,
		}
	case InstMovIR:
		p.currentInst.Data = InstMovZXIR{
			Mov: mov,
		}
	case InstMovDR:
		p.currentInst.Data = InstMovZXDR{
			Mov: mov,
		}
	case InstMovDRO1:
		p.currentInst.Data = InstMovZXDRO1{
			Mov: mov,
		}
	case InstMovDRO2:
		p.currentInst.Data = InstMovZXDRO2{
			Mov: mov,
		}
	default:
		return errors.MakeParserError(
			p.currentInst.Line,
			p.currentInst.Col,
			"Zero extend move not allowed for this type of data move.",
		)
	}
	return nil
}
