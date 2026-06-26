package assembler

import (
	"fmt"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
)

type pair [2]int
type precedenceMap map[int]pair

var inBMap = precedenceMap{
	lx.TOKEN_TPLUS:     pair{1, 2},
	lx.TOKEN_TMINUS:    pair{1, 2},
	lx.TOKEN_TASTERISK: pair{3, 4},
	lx.TOKEN_TSLASH:    pair{3, 4},
}
var preBMap = precedenceMap{
	lx.TOKEN_TMINUS: pair{0, 5},
}

func (p *Parser) ParseExpression() (*Expr, error) {
	t, _ := p.lexer.ReadNextTokenReturn()
	p.lexer.UnreadToken()
	e, err := p.parseExpression(0)
	e.Col = t.Col
	e.Line = t.Line
	return e, err
}
func (p *Parser) parseExpression(minBp int) (*Expr, error) {
	if err := p.lexer.ReadNextToken(); err != nil {
		return nil, err
	}
	lhsToken := p.lexer.CurrentToken()
	var lhs *Expr
	switch lhsToken.Ty {
	case lx.TOKEN_TEOF:
		return nil, errors.PrematureEndOfInput(p.lexer.GetPos())
	case lx.TOKEN_TNEWLINE:
		return nil, errors.PrematureEndOfInput(p.lexer.GetPos())
	case lx.TOKEN_TOPENPAREN:
		inner, err := p.parseExpression(0)
		if err != nil {
			return inner, err
		}
		if err := p.lexer.ReadNextToken(); err != nil {
			return inner, err
		}
		if p.lexer.CurrentToken().Ty != lx.TOKEN_TCLOSEDPAREN {
			t := p.lexer.CurrentToken()
			return inner, errors.UnclosedParen(t.Line, t.Col)
		}
		lhs = inner
	case lx.TOKEN_TOPENBRACKET:
		inner, err := p.parseExpression(0)
		if err != nil {
			return inner, err
		}
		if err := p.lexer.ReadNextToken(); err != nil {
			return inner, err
		}
		if p.lexer.CurrentToken().Ty != lx.TOKEN_TCLOSEDBRACKET {
			t := p.lexer.CurrentToken()
			return inner, errors.FailedToParse("dereference expression",
				t.Line, t.Col,
				"Unclosed bracket in dereference expression, got `%s` instead",
				t.ForceValAsString(),
			)
		}
		deref := MakeDeref(inner)
		return deref, nil
	case lx.TOKEN_TREG:
		regData := lhsToken.Val.(lx.RegisterData)
		lhs = MakeRegexpr(byte(regData.Reg), regData.Size)
	case lx.TOKEN_TINTEGER_LIT:
		lhs = MakeConstexprU64(lhsToken.Val.(uint64))
	case lx.TOKEN_TFLOAT_LIT:
		lhs = MakeConstexprF64Bits(lhsToken.Val.(uint64))
	case lx.TOKEN_TIDENT:
		lhs = MakeConstexprIdent(lhsToken.Val.(string))
	default:
		if binding, ok := preBMap[lhsToken.Ty]; ok {
			rhs, err := p.parseExpression(binding[1])
			if err != nil {
				return lhs, err
			}
			switch lhsToken.Ty {
			case lx.TOKEN_TMINUS:
				lhs = MakeArth(
					ArthExprSub{
						A: MakeConstexprU64(0),
						B: rhs,
					},
				)
			default:
				return lhs, fmt.Errorf("Prefix operator TODO %s", p.lexer.CurrentPosition())
			}
		} else {
			return nil, errors.FailedToParse("expression",
				lhsToken.Line, lhsToken.Col,
				"Disallowed token in expression `%s`",
				lhsToken.ForceValAsString(),
			)
		}
	}
	lhs.Line = lhsToken.Line
	lhs.Col = lhsToken.Col
	for {
		if err := p.lexer.ReadNextToken(); err != nil {
			return nil, err
		}
		op := p.lexer.CurrentToken()
		if op.Ty == lx.TOKEN_TCLOSEDBRACKET ||
			op.Ty == lx.TOKEN_TEOF ||
			op.Ty == lx.TOKEN_TNEWLINE ||
			op.Ty == lx.TOKEN_TCLOSEDPAREN ||
			op.Ty == lx.TOKEN_TDOUBLESEMICOLON {
			p.lexer.UnreadToken()
			break
		}
		binding, ok := inBMap[op.Ty]
		if !ok {
			p.lexer.UnreadToken()
			break
		}
		if binding[0] < minBp {
			p.lexer.UnreadToken()
			break
		}
		rhs, err := p.parseExpression(binding[1])
		if err != nil {
			return lhs, err
		}
		lhs = &Expr{
			Val: ArthExpr{
				Val: opTyToArthExpr(lhs, rhs, op.Ty),
			},
		}
	}
	return lhs, nil
}

func opTyToArthExpr(lhs, rhs *Expr, ty int) any {
	switch ty {
	case lx.TOKEN_TPLUS:
		return ArthExprAdd{
			A: lhs,
			B: rhs,
		}
	case lx.TOKEN_TMINUS:
		return ArthExprSub{
			A: lhs,
			B: rhs,
		}
	case lx.TOKEN_TASTERISK:
		return ArthExprMul{
			A: lhs,
			B: rhs,
		}
	case lx.TOKEN_TSLASH:
		return ArthExprDiv{
			A: lhs,
			B: rhs,
		}
	}
	return nil
}
