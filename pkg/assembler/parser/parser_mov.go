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
func (p *Parser) parseZXMov() error {
	if err := p.parseMov(); err != nil {
		return err
	}
	switch mov := p.currentInst.Data.(type) {
	case InstMovRR:
		p.currentInst.Data = InstMovZXRR{
			Src:  mov.Src,
			Dest: mov.Dest,
		}
	case InstMovIR:
		p.currentInst.Data = InstMovZXIR{
			Dest: mov.Dest,
			Imm:  mov.Imm,
		}
	case InstMovDR:
		p.currentInst.Data = InstMovZXDR{
			Dest:    mov.Dest,
			Address: mov.Address,
		}
	case InstMovDRO1:
		p.currentInst.Data = InstMovZXDRO1{
			Dest:   mov.Dest,
			Offset: mov.Offset,
			OReg1:  mov.OReg1,
			OffOp:  mov.OffOp,
		}
	case InstMovDRO2:
		p.currentInst.Data = InstMovZXDRO2{
			Dest:   mov.Dest,
			Offset: mov.Offset,
			OReg1:  mov.OReg1,
			OReg2:  mov.OReg2,
			RegOp:  mov.RegOp,
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
	switch mov := p.currentInst.Data.(type) {
	case InstMovRR:
		p.currentInst.Data = InstXCHGRR{
			Src:  mov.Src,
			Dest: mov.Dest,
		}
	case InstMovDR:
		p.currentInst.Data = InstXCHGDR{
			Dest:    mov.Dest,
			Address: mov.Address,
		}
	case InstMovRD:
		p.currentInst.Data = InstXCHGRD{
			Src:     mov.Src,
			Address: mov.Address,
		}
	case InstMovDRO1:
		p.currentInst.Data = InstXCHGDRO1{
			Dest:   mov.Dest,
			Offset: mov.Offset,
			OReg1:  mov.OReg1,
			OffOp:  mov.OffOp,
		}
	case InstMovDRO2:
		p.currentInst.Data = InstXCHGDRO2{
			Dest:   mov.Dest,
			Offset: mov.Offset,
			OReg1:  mov.OReg1,
			OReg2:  mov.OReg2,
			RegOp:  mov.RegOp,
		}
	case InstMovRDO1:
		p.currentInst.Data = InstXCHGRDO1{
			Src:    mov.Src,
			Offset: mov.Offset,
			OReg1:  mov.OReg1,
			OffOp:  mov.OffOp,
		}
	case InstMovRDO2:
		p.currentInst.Data = InstXCHGRDO2{
			Src:    mov.Src,
			Offset: mov.Offset,
			OReg1:  mov.OReg1,
			OReg2:  mov.OReg2,
			RegOp:  mov.RegOp,
		}
	default:
		return errors.MakeParserError(
			p.currentInst.Line,
			p.currentInst.Col,
			"Exchange not allowed for this type of data move.",
		)
	}
	return nil
}
