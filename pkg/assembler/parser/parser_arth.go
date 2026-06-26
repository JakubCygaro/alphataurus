package assembler

import (
	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func (p *Parser) parseAddOrSub(arthTy int) error {
	genericArth := InstArth{}
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
		p.lexer.UnreadToken()
	}
	if expr, err := p.ParseExpression(); err != nil {
		return err
	} else {
		genericArth.Src = expr
	}
	// else if eval, ok :=
	// 	TryEvaluateExpression(expr); !IsConstexprType(eval, CONSTEXPR_TREG) || !ok {
	// 	em, _ := eval.Emit()
	// 	return errors.FailedToParse(p.currentIdent,
	// 		eval.Line, eval.Col,
	// 		"First operand to instruction must be a valid register, got `%s`",
	// 		em)
	// } else {
	// 	op1.Ty = lx.TOKEN_TREG
	// 	op1.Val = expr.Val.(ConstExpr).UnpackAsRegisterData()
	// 	op1.Col = eval.Col
	// 	op1.Line = eval.Line
	// }
	// switch op1.Val.(lx.RegisterData).Reg {
	// case vm.IP_IDX:
	// 	return errors.FailedToParse(p.currentIdent,
	// 		op1.Line, op1.Col,
	// 		"Disallowed destination register `%s`",
	// 		op1.Val.(lx.RegisterData).String(),
	// 	)
	// }

	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	comma := p.lexer.CurrentToken()

	if comma.Ty != lx.TOKEN_TCOMMA {
		return errors.FailedToParse(p.currentIdent,
			comma.Line, comma.Col,
			"Instruction missing a comma, got `%s` instead",
			comma.ForceValAsString())
	}
	// var op2 ConstExpr
	// var op2Line, op2Col int
	if expr, err := p.parseExpression(0); err != nil {
		return err
	} else {
		genericArth.Dest = expr
	}
	// else {
	// 	eval, ok := TryConstEvaluateExpression(expr)
	// 	if eval.Ty != CONSTEXPR_TREG && !ok {
	// 		t := p.lexer.CurrentToken()
	// 		// em, _ := expr.Emit()
	// 		return errors.FailedToParse(p.currentIdent,
	// 			expr.Line, expr.Col,
	// 			"Second operand to instruction has to be a valid register "+
	// 				"or a compile time expression, got `%s`", t.ForceValAsString())
	// 	}
	// 	op2 = eval
	// 	op2Line = expr.Line
	// 	op2Col = expr.Col
	// }
	var ty ArthTy
	switch arthTy {
	case ARTH_TADD:
		ty = ADD
	case ARTH_TSUB:
		ty = SUB
	}
	if IsConstexprType[ConstExprReg](genericArth.Dest) {
		if IsConstexprType[ConstExprReg](genericArth.Src) {
			p.currentInst = Instruction{
				Data: InstArthRR{
					Src:    genericArth.Src.Val(ConstExpr).Val(ConstExprReg).Reg,
					Dest:   genericArth.Dest.Val(ConstExpr).Val(ConstExprReg).Reg,
					Ty:     valTy,
					ArthTy: ty,
				},
			}
		} else if IsConstexprType[ConstExprILit](genericArth.Src) {
			p.currentInst = Instruction{
				Data: InstArthIR{
					Imm:    genericArth.Src.Val(ConstExpr).Val(ConstExprILit).Integer,
					Dest:   genericArth.Dest.Val(ConstExpr).Val(ConstExprReg).Reg,
					Ty:     valTy,
					ArthTy: ty,
				},
			}
		} else if IsConstexprType[ConstExprFLit](genericArth.Src) {
			p.currentInst = Instruction{
				Data: InstArthIR{
					Imm:    genericArth.Src.Val(ConstExpr).Val(ConstExprFLit).Float,
					Dest:   genericArth.Dest.Val(ConstExpr).Val(ConstExprReg).Reg,
					Ty:     valTy,
					ArthTy: ty,
				},
			}
		}
	} else {
		p.currentInst = Instruction{
			Data: genericArth,
		}
	}
	// switch op2.Ty {
	// case CONSTEXPR_TREG:
	// 	switch int(op2.Val) {
	// 	case vm.IP_IDX:
	// 		em, _ := op2.Emit()
	// 		return errors.FailedToParse(p.currentIdent,
	// 			op2Line, op2Col,
	// 			"Disallowed source register `%s`", em)
	// 	}
	// 	var ty ArthTy
	// 	switch arthTy {
	// 	case ARTH_TADD:
	// 		ty = ADD
	// 	case ARTH_TSUB:
	// 		ty = SUB
	// 	}
	// 	p.currentInst = Instruction{
	// 		Data: InstArthRR{
	// 			Src:    op2.UnpackAsRegisterData(),
	// 			Dest:   op1.Val.(lx.RegisterData),
	// 			Ty:     valTy,
	// 			ArthTy: ty,
	// 		},
	// 	}
	// case CONSTEXPR_TILIT:
	// 	var ty ArthTy
	// 	switch arthTy {
	// 	case ARTH_TADD:
	// 		ty = ADD
	// 	case ARTH_TSUB:
	// 		ty = SUB
	// 	}
	// 	p.currentInst = Instruction{
	// 		Data: InstArthIR{
	// 			Imm:    op2.Val,
	// 			Dest:   op1.Val.(lx.RegisterData),
	// 			Ty:     valTy,
	// 			ArthTy: ty,
	// 		},
	// 	}
	// case CONSTEXPR_TFLIT:
	// 	var ty ArthTy
	// 	switch arthTy {
	// 	case ARTH_TADD:
	// 		ty = ADD
	// 	case ARTH_TSUB:
	// 		ty = SUB
	// 	}
	// 	p.currentInst = Instruction{
	// 		Data: InstArthIR{
	// 			Imm:    op2.Val,
	// 			Dest:   op1.Val.(lx.RegisterData),
	// 			Ty:     valTy,
	// 			ArthTy: ty,
	// 		},
	// 	}
	// default:
	// 	em, _ := op2.Emit()
	// 	return errors.FailedToParse(p.currentIdent,
	// 		op2Line, op2Col,
	// 		"Second operand to instruction has to be a valid register "+
	// 			"or a compile time expression got `%s`", em)
	// }
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
		p.lexer.UnreadToken()
	}
	err = p.lexer.ReadNextToken()
	if err != nil {
		return err
	}
	sizeT := p.lexer.CurrentToken()
	var size byte
	if sz, ok := lx.TokenAsSize(&sizeT); !ok {
		return errors.FailedToParse(p.currentIdent, op1.Line, op1.Col,
			"Missing data size parameter, got `%s`", sizeT.ForceValAsString())
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
			return errors.FailedToParse(p.currentIdent,
				p.currentStartToken.Line, p.currentStartToken.Col,
				"Invalid data type specifier `%s`. "+
					"Either no specifier or FLOAT are allowed.",
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
		return errors.PrematureEndOfInput(p.lexer.CurrentToken().Line,
			p.lexer.CurrentToken().Col)
	}
	if op1.Ty != lx.TOKEN_TREG {
		return errors.FailedToParse(p.currentIdent,
			op1.Line, op1.Col,
			"The instruction operand must be a valid register, got `%s`",
			op1.ForceValAsString())
	}
	switch op1.Val.(lx.RegisterData).Reg {
	case vm.IP_IDX:
		return errors.FailedToParse(p.currentIdent,
			op1.Line, op1.Col,
			"Disallowed operand register `%s`",
			op1.ForceValAsString())
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
		return errors.FailedToParse(p.currentIdent,
			op1.Line, op1.Col,
			"The instruction operand must be a valid register, got `%s`",
			op1.ForceValAsString())
	}
	switch op1.Val.(lx.RegisterData).Reg {
	case vm.IP_IDX:
		return errors.FailedToParse(p.currentIdent,
			op1.Line, op1.Col,
			"Disallowed operand register `%s`",
			op1.ForceValAsString())
	}
	p.currentInst = Instruction{
		Data: InstDec{
			Reg: op1.Val.(lx.RegisterData),
		},
	}
	return nil
}
