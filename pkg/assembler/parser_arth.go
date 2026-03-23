package assembler

import (
	"fmt"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func (p *Parser) parseAddOrSub(arthTy int) error {
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	op1 := p.lexer.CurrentToken()
	valTy := ARTH_TUNSIGNED
	if op1.Ty == TOKEN_TSIGNED || op1.Ty == TOKEN_TFLOAT || op1.Ty == TOKEN_TUNSIGNED {
		switch op1.Ty {
		case TOKEN_TSIGNED:
			valTy = ARTH_TSIGNED
		case TOKEN_TUNSIGNED:
			valTy = ARTH_TUNSIGNED
		case TOKEN_TFLOAT:
			valTy = ARTH_TFLOAT
		}
	} else {
		p.lexer.UnreadToken()
	}
	if expr, err := p.parseExpression(0); err != nil {
		return err
	} else if eval, _ := TryEvaluateExpression(expr); eval.Ty != CONSTEXPR_TREG {
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"First operand to instruction must be a valid register",
			p.lexer.line, p.lexer.col)
	} else {
		op1.Ty = TOKEN_TREG
		op1.Val = int(expr.Val.(ConstExpr).Val)
	}
	switch op1.Val.(int) {
	case vm.IP_IDX:
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"Disallowed source register",
			p.lexer.line, p.lexer.col)
	}

	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	comma := p.lexer.CurrentToken()

	if comma.Ty != TOKEN_TCOMMA {
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"Instruction missing a comma",
			p.lexer.line, p.lexer.col)
	}
	var op2 ConstExpr
	if expr, err := p.parseExpression(0); err != nil {
		return err
	} else {
		eval, ok := TryConstEvaluateExpression(expr)
		if eval.Ty != CONSTEXPR_TREG && !ok {
			return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
				"Second operand to instruction has to be a valid register or a compile time expression",
				p.lexer.line, p.lexer.col)
		}
		op2 = eval
	}
	switch op2.Ty {
	case CONSTEXPR_TREG:
		switch int(op2.Val) {
		case vm.IP_IDX:
			return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
				"Disallowed destination register",
				p.lexer.line, p.lexer.col)
		}
		var ty int
		switch arthTy {
		case ARTH_TADD:
			ty = INST_TADDRR
		case ARTH_TSUB:
			ty = INST_TSUBRR
		}
		p.currentInst = Instruction{
			Ty: ty,
			Data: InstArthData{
				Source:  int(op2.Val),
				Dest: op1.Val.(int),
				Ty:   valTy,
			},
		}
	case CONSTEXPR_TILIT:
		var ty int
		switch arthTy {
		case ARTH_TADD:
			ty = INST_TADDIR
		case ARTH_TSUB:
			ty = INST_TSUBIR
		}
		p.currentInst = Instruction{
			Ty: ty,
			Data: InstArthData{
				Imm:  op2.Val,
				Dest: op1.Val.(int),
				Ty:   valTy,
			},
		}
	case CONSTEXPR_TFLIT:
		var ty int
		switch arthTy {
		case ARTH_TADD:
			ty = INST_TADDIR
		case ARTH_TSUB:
			ty = INST_TSUBIR
		}
		p.currentInst = Instruction{
			Ty: ty,
			Data: InstArthData{
				Imm:  op2.Val,
				Dest: op1.Val.(int),
				Ty:   valTy,
			},
		}
	default:
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"Second operand to instruction has to be a valid register or a compile time expression",
			p.lexer.line, p.lexer.col)
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
	case TOKEN_TSIGNED:
		valTy = ARTH_TSIGNED
	case TOKEN_TUNSIGNED:
		valTy = ARTH_TUNSIGNED
	case TOKEN_TFLOAT:
		valTy = ARTH_TFLOAT
	case TOKEN_TNEWLINE:
		p.lexer.UnreadToken()
	case TOKEN_TEOF:
		p.lexer.UnreadToken()
	default:
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"Invalid instruction",
			p.lexer.line, p.lexer.col)
	}
	var ty int
	switch arthTy {
	case ARTH_TDIV:
		ty = INST_TDIVRR
	case ARTH_TMUL:
		ty = INST_TMULRR
	}
	p.currentInst = Instruction{
		Ty: ty,
		Data: InstArthData{
			Ty: valTy,
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
	if op1.Ty == TOKEN_TEOF {
		return errors.PrematureEndOfInput(p.lexer.line, p.lexer.col)
	}
	if op1.Ty != TOKEN_TREG {
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"The instruction operand must be a valid register",
			p.lexer.line, p.lexer.col)
	}
	switch op1.Val.(int) {
	case vm.IP_IDX:
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"Disallowed operand register",
			p.lexer.line, p.lexer.col)
	}
	p.currentInst = Instruction{
		Ty: INST_TINCR,
		Data: InstIncDecData{
			Reg: op1.Val.(int),
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
	if op1.Ty == TOKEN_TEOF {
		return errors.PrematureEndOfInput(p.lexer.line, p.lexer.col)
	}
	if op1.Ty != TOKEN_TREG {
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"The instruction operand must be a valid register",
			p.lexer.line, p.lexer.col)
	}
	switch op1.Val.(int) {
	case vm.IP_IDX:
		return errors.FailedToParse(fmt.Sprintf("%s instruction", p.currentIdent),
			"Disallowed operand register",
			p.lexer.line, p.lexer.col)
	}
	p.currentInst = Instruction{
		Ty: INST_TDECR,
		Data: InstIncDecData{
			Reg: op1.Val.(int),
		},
	}
	return nil
}
