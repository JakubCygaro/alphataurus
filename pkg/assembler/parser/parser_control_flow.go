package assembler

import (
	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func (p *Parser) parseCmp() error {
	genericCmp := InstGenericCmp{}
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	op1 := p.lexer.CurrentToken()
	ty := vm.TY_UINT
	switch op1.Ty {
	case lx.TOKEN_TFLOAT:
		ty = vm.TY_FLOAT
	case lx.TOKEN_TSIGNED:
		ty = vm.TY_SINT
	case lx.TOKEN_TUNSIGNED:
		ty = vm.TY_UINT
	default:
		p.lexer.UnreadCurrentToken()
	}
	genericCmp.Ty = ty
	if expr, err := p.ParseExpression(); err != nil {
		return err
	} else {
		genericCmp.Min = expr
	}
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	if p.lexer.CurrentToken().Ty != lx.TOKEN_TCOMMA {
		t := p.lexer.CurrentToken()
		return errors.FailedToParse(p.currentIdent,
			t.Line, t.Col,
			"Instruction missing a comma, got `%s`",
			t.ForceValAsString(),
		)
	}
	if expr, err := p.ParseExpression(); err != nil {
		return err
	} else {
		genericCmp.Sub = expr
	}
	if concrete, err := GetConcreteCmpInst(genericCmp); concrete != nil && err == nil {
		p.currentInst = Instruction{
			Data: concrete,
		}
	} else {
		p.currentInst = Instruction{
			Data: genericCmp,
		}
	}
	return nil
}
func (p *Parser) parseJmp(ty JmpVariant) error {
	genericJmp := InstGenericJmp{}
	var absolute bool
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	} else if p.lexer.CurrentToken().Ty == lx.TOKEN_TABSOLUTE {
		absolute = true
	} else {
		p.lexer.UnreadCurrentToken()
		absolute = false
	}
	if expr, err := p.ParseExpression(); err != nil {
		return err
	} else {
		genericJmp.Expr = expr
	}
	genericJmp.Absolute = absolute
	genericJmp.Variant = ty
	p.currentInst = Instruction{
		Data: genericJmp,
	}
	return nil
}
func (p *Parser) parseCall() error {
	genericCall := InstGenericCall{}
	if param, err := p.ParseExpression(); err != nil {
		return err
	} else {
		genericCall.Expr = param
	}
	p.currentInst = Instruction{
		Data: genericCall,
	}
	return nil
}
