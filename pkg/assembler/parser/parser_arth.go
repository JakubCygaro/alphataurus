package assembler

import (
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
	"github.com/JakubCygaro/alphataurus/pkg/assembler/parser/errors"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func (p *Parser) parseAddOrSub(arthTy int) error {
	genericArth := InstGenericArth{}
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	op1 := p.lexer.CurrentToken()
	valTy := ARTH_TUNSIGNED
	if op1.Ty == lx.TOKEN_TSIGNED || op1.Ty == lx.TOKEN_TFLOAT || op1.Ty == lx.TOKEN_TUNSIGNED {
		switch op1.Ty {
		case lx.TOKEN_TSIGNED:
			valTy = ARTH_TSIGNED
		case lx.TOKEN_TUNSIGNED:
			valTy = ARTH_TUNSIGNED
		case lx.TOKEN_TFLOAT:
			valTy = ARTH_TFLOAT
		}
	} else {
		p.lexer.UnreadCurrentToken()
	}
	if expr, err := p.ParseExpression(); err != nil {
		return err
	} else {
		genericArth.Dest = expr
	}
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	comma := p.lexer.CurrentToken()

	if comma.Ty != lx.TOKEN_TCOMMA {
		return errors.MissingComma(
			comma.Line, comma.Col,
		)
	}
	if expr, err := p.parseExpression(0); err != nil {
		return err
	} else {
		genericArth.Src = expr
	}
	var ty ArthTy
	switch arthTy {
	case ARTH_TADD:
		ty = ADD
	case ARTH_TSUB:
		ty = SUB
	}
	genericArth.ArthTy = ty
	genericArth.Ty = valTy
	if concrete, err :=
		GetConcreteArthInst(genericArth, &p.currentInst); concrete != nil && err == nil {
		p.currentInst = Instruction{
			Data: concrete,
		}
	} else {
		p.currentInst = Instruction{
			Data: genericArth,
		}
	}
	return nil
}
func (p *Parser) parseDivOrMul(arthTy int) error {
	err := p.lexer.ReadNextToken()
	if err != nil {
		return err
	}
	op1 := p.lexer.CurrentToken()
	valTy := ARTH_TUNSIGNED
	switch op1.Ty {
	case lx.TOKEN_TSIGNED:
		valTy = ARTH_TSIGNED
	case lx.TOKEN_TUNSIGNED:
		valTy = ARTH_TUNSIGNED
	case lx.TOKEN_TFLOAT:
		valTy = ARTH_TFLOAT
	default:
		valTy = -1
		p.lexer.UnreadCurrentToken()
	}
	err = p.lexer.ReadNextToken()
	if err != nil {
		return err
	}
	sizeT := p.lexer.CurrentToken()
	var size byte
	if sz, ok := lx.TokenAsSize(&sizeT); !ok {
		return errors.MakeParserError(
			op1.Line, op1.Col,
			"Missing data size parameter, got `%s`.",
			sizeT.ForceValAsString(),
		)
	} else {
		size = sz
	}
	var ty ArthTy
	switch arthTy {
	case ARTH_TDIV:
		ty = DIV
	case ARTH_TMUL:
		ty = MUL
		if valTy == -1 {
			valTy = ARTH_TUNSIGNED
		} else if valTy != ARTH_TFLOAT {
			return errors.MakeParserError(
				op1.Line, op1.Col,
				"Invalid data type specifier `%s`. "+
					"Either no specifier or only FLOAT is allowed.",
				op1.ForceValAsString(),
			)
		}
	}
	p.currentInst = Instruction{
		Data: InstArthRR{
			Ty:       valTy,
			DataSize: size,
			ArthTy:   ty,
		},
	}
	return nil
}
func (p *Parser) parseInc() error {
	err := p.lexer.ReadNextToken()
	if err != nil {
		return err
	}
	op1 := p.lexer.CurrentToken()
	if op1.Ty == lx.TOKEN_TEOF {
		return errors.PrematureEndOfInput(
			op1.Line,
			op1.Col,
		)
	}
	if op1.Ty != lx.TOKEN_TREG {
		return errors.MakeParserError(
			op1.Line, op1.Col,
			"The instruction operand must be a valid register, got `%s`.",
			op1.ForceValAsString(),
		)
	}
	switch op1.Val.(lx.RegisterData).Reg {
	case vm.IP_IDX:
		return errors.MakeParserError(
			op1.Line, op1.Col,
			"Disallowed operand register `%s`.",
			op1.ForceValAsString(),
		)
	}
	p.currentInst = Instruction{
		Data: InstInc{
			Reg: op1.Val.(lx.RegisterData),
		},
	}
	return nil
}
func (p *Parser) parseDec() error {
	err := p.lexer.ReadNextToken()
	if err != nil {
		return err
	}
	op1 := p.lexer.CurrentToken()
	if op1.Ty == lx.TOKEN_TEOF {
		return errors.PrematureEndOfInput(op1.Line, op1.Col)
	}
	if op1.Ty != lx.TOKEN_TREG {
		return errors.MakeParserError(
			op1.Line, op1.Col,
			"The instruction operand must be a valid register, got `%s`.",
			op1.ForceValAsString(),
		)
	}
	switch op1.Val.(lx.RegisterData).Reg {
	case vm.IP_IDX:
		return errors.MakeParserError(
			op1.Line, op1.Col,
			"Disallowed operand register `%s`.",
			op1.ForceValAsString(),
		)
	}
	p.currentInst = Instruction{
		Data: InstDec{
			Reg: op1.Val.(lx.RegisterData),
		},
	}
	return nil
}
