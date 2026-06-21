package assembler

import (
	"fmt"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
)

type pair [2]int
type precedenceMap map[int]pair

var inBMap = precedenceMap{
	TOKEN_TPLUS:     pair{1, 2},
	TOKEN_TMINUS:    pair{1, 2},
	TOKEN_TASTERISK: pair{3, 4},
	TOKEN_TSLASH:    pair{3, 4},
}
var preBMap = precedenceMap{
	TOKEN_TMINUS: pair{0, 5},
}

func (p *Parser) ParseExpression() (*Expr, error) {
	return p.parseExpression(0)
}
func (p *Parser) parseExpression(minBp int) (*Expr, error) {
	if err := p.lexer.ReadNextToken(); err != nil {
		return nil, err
	}
	lhsToken := p.lexer.CurrentToken()
	var lhs *Expr
	switch lhsToken.Ty {
	case TOKEN_TEOF:
		return nil, errors.PrematureEndOfInput(p.lexer.line, p.lexer.col)
	case TOKEN_TNEWLINE:
		return nil, errors.PrematureEndOfInput(p.lexer.line, p.lexer.col)
	case TOKEN_TOPENPAREN:
		inner, err := p.parseExpression(0)
		if err != nil {
			return inner, err
		}
		if err := p.lexer.ReadNextToken(); err != nil {
			return inner, err
		}
		if p.lexer.CurrentToken().Ty != TOKEN_TCLOSEDPAREN {
			return inner, errors.UnclosedParen(p.lexer.line, p.lexer.col)
		}
		lhs = inner
	case TOKEN_TOPENBRACKET:
		inner, err := p.parseExpression(0)
		if err != nil {
			return inner, err
		}
		if err := p.lexer.ReadNextToken(); err != nil {
			return inner, err
		}
		if p.lexer.CurrentToken().Ty != TOKEN_TCLOSEDBRACKET {
			t := p.lexer.CurrentToken()
			return inner, errors.FailedToParse("dereference expression",
				t.Line, t.Col,
				"Unclosed bracket in dereference expression, got `%s` instead",
				t.ForceValAsString(),
			)
		}
		deref := MakeDeref(inner)
		return deref, nil
	case TOKEN_TREG:
		regData := lhsToken.Val.(RegisterData)
		lhs = MakeConstexprR(byte(regData.Reg), regData.Size)
	case TOKEN_TINTEGER_LIT:
		lhs = MakeConstexprU64(lhsToken.Val.(uint64))
	case TOKEN_TFLOAT_LIT:
		lhs = MakeConstexprF64Bits(lhsToken.Val.(uint64))
	case TOKEN_TIDENT:
		lhs = MakeConstexprIdent(lhsToken.Val.(string))
	default:
		if binding, ok := preBMap[lhsToken.Ty]; ok {
			rhs, err := p.parseExpression(binding[1])
			if err != nil {
				return lhs, err
			}
			switch lhsToken.Ty {
			case TOKEN_TMINUS:
				lhs = MakeArth(
					MakeConstexprU64(0),
					rhs,
					lhsToken.Ty,
				)
			default:
				return lhs, fmt.Errorf("Prefix operator TODO %s", p.lexer.CurrentPosition())
			}
		} else {
			em, _ := lhs.Emit()
			return nil, errors.FailedToParse("expression",
				lhsToken.Line, lhsToken.Col,
				"Bad expression `%s`",
				em,
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
		if op.Ty == TOKEN_TCLOSEDBRACKET ||
			op.Ty == TOKEN_TEOF ||
			op.Ty == TOKEN_TNEWLINE ||
			op.Ty == TOKEN_TCLOSEDPAREN ||
			op.Ty == TOKEN_TDOUBLESEMICOLON {
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
			Ty: EXPR_TARTH,
			Val: ArthExpr{
				Ty: op.Ty,
				A:  lhs,
				B:  rhs,
			},
		}
	}
	return lhs, nil
}
