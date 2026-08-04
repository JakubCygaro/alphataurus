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
		return errors.MissingComma(
			comma.Line, comma.Col,
			comma,
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
	} else if err != nil && p.ForceCoalesceGenerics {
		return err
	} else {
		p.currentInst = Instruction{
			Data: genericLogical,
		}
	}
	return nil
}
func (p *Parser) parseNeg() error {
	float := false
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	if p.lexer.CurrentToken().Ty != lx.TOKEN_TFLOAT {
		p.lexer.UnreadCurrentToken()
	} else {
		float = true
	}
	if r, ok := p.lexer.Expect(lx.TOKEN_TREG); !ok {
		return errors.MakeParserError(
			r.Line,
			r.Col,
			"Expected register, got '%s'",
			r.ForceValAsString(),
		)
	} else {
		p.currentInst = Instruction{
			Data: InstNegR{
				Arg:   r.Val.(lx.RegisterData),
				Float: float,
			},
		}
	}
	return nil
}
func (p *Parser) parseRotate(left bool) error {
	var reg lx.RegisterData
	var rotateBy uint64
	if r, ok := p.lexer.Expect(lx.TOKEN_TREG); !ok {
		return errors.MakeParserError(
			r.Line,
			r.Col,
			"Expected register, got '%s'",
			r.ForceValAsString(),
		)
	} else {
		reg = r.Val.(lx.RegisterData)
	}
	if r, ok := p.lexer.Expect(lx.TOKEN_TCOMMA); !ok {
		return errors.MissingComma(
			r.Line,
			r.Col,
			r,
		)
	}
	if r, ok := p.lexer.Expect(lx.TOKEN_TINTEGER_LIT); !ok {
		return errors.MakeParserError(
			r.Line,
			r.Col,
			"Expected integer literal, got '%s'",
			r.ForceValAsString(),
		)
	} else {
		rotateBy = r.Val.(uint64)
	}

	p.currentInst = Instruction{
		Data: InstRotate{
			Arg: reg,
			Left: left,
			RotateBy: rotateBy,
		},
	}

	return nil
}
