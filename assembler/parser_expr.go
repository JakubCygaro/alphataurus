package assembler

import "fmt"

type pair [2]int
type precedenceMap map[int]pair

var inBMap = precedenceMap{
	TOKEN_TPLUS:     pair{1, 2},
	TOKEN_TMINUS:    pair{1, 2},
	TOKEN_TASTERISK: pair{3, 4},
	TOKEN_TSLASH:    pair{3, 4},
}
var preBMap = precedenceMap{
	TOKEN_TMINUS:    pair{0, 5},
}

func (p *Parser) parseExpression(minBp int) (Expr, error) {
	if err := p.lexer.ReadNextToken(); err != nil {
		return Expr{}, err
	}
	lhsToken := p.lexer.CurrentToken()
	var lhs Expr
	switch lhsToken.Ty {
	case TOKEN_TEOF:
		return Expr{}, fmt.Errorf("Premature end of input while parsing expression %s", p.lexer.CurrentPosition())
	case TOKEN_TNEWLINE:
		return Expr{}, fmt.Errorf("Premature end of input while parsing expression %s", p.lexer.CurrentPosition())
	case TOKEN_TOPENBRACKET:
		inner, err := p.parseExpression(0)
		if err != nil {
			return inner, err
		}
		if inner.Ty != EXPR_TARTH {
			return inner, fmt.Errorf("Invalid inner expression of deref expression %s", p.lexer.CurrentPosition())
		}
		if err := p.lexer.ReadNextToken(); err != nil {
			return inner, err
		}
		if p.lexer.CurrentToken().Ty != TOKEN_TCLOSEDBRACKET {
			return inner, fmt.Errorf("Unclosed deref expression bracket %s", p.lexer.CurrentPosition())
		}
		deref := Expr{
			Ty: EXPR_TDEREF,
			Val: DerefExpr{
				Inner: inner.Val.(ArthExpr),
			},
		}
		return deref, nil
	case TOKEN_TREG:
		lhs = Expr{
			Ty: EXPR_TCONST,
			Val: ConstExpr{
				Ty:  CONSTEXPR_TREG,
				Val: uint64(lhsToken.val.(int)),
			},
		}
	case TOKEN_TINTEGER_LIT:
		lhs = Expr{
			Ty: EXPR_TCONST,
			Val: ConstExpr{
				Ty:  CONSTEXPR_TILIT,
				Val: lhsToken.val.(uint64),
			},
		}
	case TOKEN_TFLOAT_LIT:
		lhs = Expr{
			Ty: EXPR_TCONST,
			Val: ConstExpr{
				Ty:  CONSTEXPR_TFLIT,
				Val: lhsToken.val.(uint64),
			},
		}
	default:
		if binding, ok := preBMap[lhsToken.Ty]; ok{
			rhs, err := p.parseExpression(binding[1])
			if err != nil {
				return rhs, err
			}
			lhs = Expr {
				Ty: ARTHEXPR_TSUB,
				Val: ArthExpr {
					A: &Expr{
						Ty: CONSTEXPR_TILIT,
						Val: 0,
					},
					B: &rhs,
				},
			}
		} else {
			return Expr{}, fmt.Errorf("Bad expression %s", p.lexer.CurrentPosition())
		}
	}
	var retExpr Expr
	for {
		if err := p.lexer.ReadNextToken(); err != nil {
			return Expr{}, err
		}
		op := p.lexer.CurrentToken()
		if op.Ty == TOKEN_TCLOSEDBRACKET || op.Ty == TOKEN_TEOF || op.Ty == TOKEN_TNEWLINE {
			p.lexer.UnreadToken()
			return lhs, nil
		}
		if op.Ty < TOKEN_OPSTART || op.Ty > TOKEN_OPEND {
			p.lexer.UnreadToken()
			return lhs, nil
		}
		binding := inBMap[op.Ty]
		if binding[0] < minBp {
			break
		}
		rhs, err := p.parseExpression(binding[1])
		if err != nil {
			return lhs, err
		}
		retExpr = Expr{
			Ty: EXPR_TARTH,
			Val: ArthExpr{
				Ty: op.Ty,
				A:  &lhs,
				B:  &rhs,
			},
		}
	}
	return retExpr, nil
}
