package assembler

import (
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
	"github.com/JakubCygaro/alphataurus/pkg/assembler/parser/errors"
)

func (p *Parser) parseLogical(logTy int) error {
	genericLogical := InstGenericLogical{}
	if expr, err := p.ParseExpression(); err != nil {
		return err
	} else {
		genericLogical.First = expr
	}
	if logTy == LOG_TNOT {
		if !genericLogical.First.IsRegexpr() {
			em, _ := genericLogical.First.Emit()
			return errors.MakeParserError(
				genericLogical.First.Line, genericLogical.First.Col,
				"First operand to instruction must be a valid register. "+
					"In expression `%s`.",
				em,
			)
		}
		p.currentInst = Instruction{
			Data: InstNotR{
				First: genericLogical.First.Val.(RegExpr).Reg,
			},
		}
		return nil
	}

	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	comma := p.lexer.CurrentToken()

	if comma.Ty != lx.TOKEN_TCOMMA {
		return errors.MakeParserError(
			comma.Line, comma.Col,
			"Instruction missing a comma, got `%s`.",
			comma.ForceValAsString(),
		)
	}
	if expr, err := p.ParseExpression(); err != nil {
		return err
	} else {
		genericLogical.Second = expr
	}
	var ty LogTy
	switch logTy {
	case LOG_TOR:
		ty = OR
	case LOG_TAND:
		ty = AND
	case LOG_TXOR:
		ty = XOR
	case LOG_TLSH:
		ty = LSH
	case LOG_TRSH:
		ty = RSH
	}
	genericLogical.LogTy = ty
	if concrete, err :=
		GetConcreteLogicalInst(genericLogical, &p.currentInst); concrete != nil &&
		err == nil {
		p.currentInst = Instruction{
			Data: concrete,
		}
	} else {
		p.currentInst = Instruction{
			Data: genericLogical,
		}
	}
	return nil
}
