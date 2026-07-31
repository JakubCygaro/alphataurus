package assembler

import (
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
	"github.com/JakubCygaro/alphataurus/pkg/assembler/parser/errors"
)

var Void void = void{}

type pair [2]int
type precedenceMap map[int]pair
type assocMap map[int]void
type endParseMap map[int]void
type allowedBetweenRegsMap map[int]void

func IsAssoc(op int) bool {
	_, isAssoc := associativeBOps[op]
	return isAssoc
}

var associativeBOps = assocMap{
	lx.TOKEN_TPLUS:     Void,
	lx.TOKEN_TASTERISK: Void,
}
var inBMap = precedenceMap{
	lx.TOKEN_TPLUS:     pair{1, 2},
	lx.TOKEN_TMINUS:    pair{1, 2},
	lx.TOKEN_TASTERISK: pair{3, 4},
	lx.TOKEN_TSLASH:    pair{3, 4},
}

func IsInfixOp(op int) bool {
	_, ok := inBMap[op]
	return ok
}

var preBMap = precedenceMap{
	lx.TOKEN_TMINUS: pair{0, 5},
}

func IsPrefixOp(op int) bool {
	_, ok := preBMap[op]
	return ok
}

var endParse = endParseMap{
	lx.TOKEN_TCLOSEDBRACKET:   Void,
	lx.TOKEN_TEOF:             Void,
	lx.TOKEN_TNEWLINE:         Void,
	lx.TOKEN_TCLOSEDPAREN:     Void,
	lx.TOKEN_TDOUBLESEMICOLON: Void,
}

func shouldEndParseExpr(op int) bool {
	_, ok := endParse[op]
	return ok
}

var abrMap = allowedBetweenRegsMap{
	lx.TOKEN_TPLUS:  Void,
	lx.TOKEN_TMINUS: Void,
}

func allowedBetweenRegs(op int) bool {
	_, ok := abrMap[op]
	return ok
}
func (p *Parser) ParseExpression() (*Expr, error) {
	t, _ := p.lexer.ReadNextTokenReturn()
	p.lexer.UnreadCurrentToken()
	e, err := p.parseExpression(0)
	if err == nil {
		e.Col = t.Col
		e.Line = t.Line
	}
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
			return inner, errors.
				MakeParserError(
					t.Line, t.Col,
					"Unclosed bracket in dereference expression, got `%s` instead.",
					t.ForceValAsString(),
				)
		}
		deref := MakeDeref(inner)
		return deref, nil
	case lx.TOKEN_TREG:
		regData := lhsToken.Val.(lx.RegisterData)
		// lhs = MakeRegexpr(byte(regData.Reg), regData.Size)
		var err error
		lhs, err = p.parserRegExpr(regData)
		if err != nil {
			return nil, err
		}
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
				lhs = MakeNeg(
					rhs,
				)
			default:
				return lhs, errors.
					MakeParserError(
						lhsToken.Line, lhsToken.Col,
						"Prefix operator TODO `%s`.",
						lhsToken.ForceValAsString(),
					)
			}
		} else {
			return nil, errors.MakeParserError(
				lhsToken.Line, lhsToken.Col,
				"Disallowed token in expression `%s`.",
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
		if shouldEndParseExpr(op.Ty) {
			p.lexer.UnreadCurrentToken()
			break
		}
		binding, ok := inBMap[op.Ty]
		if !ok {
			p.lexer.UnreadCurrentToken()
			break
		}
		if binding[0] < minBp {
			p.lexer.UnreadCurrentToken()
			break
		}
		rhs, err := p.parseExpression(binding[1])
		if err != nil {
			return lhs, err
		}
		if e, err := makeBinop(lhs, rhs, op); err != nil {
			return nil, err
		} else {
			lhs = e
		}
	}
	return lhs, nil
}

func (p *Parser) parserRegExpr(start lx.RegisterData) (*Expr, error) {
	var ret MemExpr
	n, e := p.lexer.ReadNextTokenReturn()
	if e != nil {
		return nil, e
	}
	// addition by default, so that the downstream code does not shit itself
	ret.DispOp = lx.TOKEN_TPLUS
	opReg := allowedBetweenRegs(n.Ty)
	sf := n.Ty == lx.TOKEN_TASTERISK
	ret.Base = start
	// plain register
	if !opReg && !sf {
		p.lexer.UnreadCurrentToken()
		return &Expr{
			Val: RegExpr{
				Reg: start,
			},
		}, nil
	}
	if sf {
		// check if sf is a simple integer literal or an ident expression
		n, e = p.lexer.ReadNextTokenReturn()
		if e != nil {
			return nil, e
		}
		if n.Ty == lx.TOKEN_TINTEGER_LIT {
			ret.ScaleF = MakeConstexprU64(n.Val.(uint64))
		} else if n.Ty == lx.TOKEN_TIDENT {
			ret.ScaleF = MakeConstexprIdent(n.Val.(string))
		} else {
			// if not, then unread parse this as an expression
			p.lexer.UnreadCurrentToken()
			if e, err := p.parseExpression(0); err != nil {
				return nil, err
			} else {
				ret.ScaleF = e
			}
		}
		// the next register has to be an operand for either the displacement
		// or another register
		n, e = p.lexer.ReadNextTokenReturn()
		if e != nil {
			return nil, e
		}
		opReg = allowedBetweenRegs(n.Ty)
		if !opReg {
			p.lexer.UnreadCurrentToken()
			return &Expr{
				Val: ret,
			}, nil
		}
	}
	op := n.Ty
	n, e = p.lexer.ReadNextTokenReturn()
	if e != nil {
		return nil, e
	}
	//check for index
	if n.Ty == lx.TOKEN_TREG {
		ret.RegOp = op
		ret.Index.Set(n.Val.(lx.RegisterData))
		n, e = p.lexer.ReadNextTokenReturn()
		if e != nil {
			return nil, e
		}
	} else {
		// // if minus then also unread it so the first expression value is negative
		if op == lx.TOKEN_TMINUS {
			p.lexer.UnreadToken(lx.Token{ Ty: lx.TOKEN_TMINUS })
		}
		// expression start
		p.lexer.UnreadCurrentToken()
		if e, err := p.parseExpression(0); err != nil {
			return nil, err
		} else {
			ret.Disp = e
			// ret.DispOp = op
			return &Expr{
				Val: ret,
			}, nil
		}
	}
	opReg = allowedBetweenRegs(n.Ty)
	sf = n.Ty == lx.TOKEN_TASTERISK
	// end of memory addressing expression
	if !opReg && !sf {
		p.lexer.UnreadCurrentToken()
		return &Expr{
			Val: ret,
		}, nil
	}
	if sf && ret.ScaleF == nil {
		// check if sf is a simple integer literal or an ident expression
		n, e = p.lexer.ReadNextTokenReturn()
		if e != nil {
			return nil, e
		}
		if n.Ty == lx.TOKEN_TINTEGER_LIT {
			ret.ScaleF = MakeConstexprU64(n.Val.(uint64))
		} else if n.Ty == lx.TOKEN_TIDENT {
			ret.ScaleF = MakeConstexprIdent(n.Val.(string))
		} else {
			// if not, then unread parse this as an expression
			p.lexer.UnreadCurrentToken()
			if e, err := p.parseExpression(0); err != nil {
				return nil, err
			} else {
				ret.ScaleF = e
			}
		}
		// the next expression has to be an operand for the displacement
		n, e = p.lexer.ReadNextTokenReturn()
		if e != nil {
			return nil, e
		}
		opReg = allowedBetweenRegs(n.Ty)
		if !opReg {
			p.lexer.UnreadCurrentToken()
			return &Expr{
				Val: ret,
			}, nil
		}
	} else if ret.ScaleF != nil {
		return nil, errors.
			MakeParserError(
				n.Line,
				n.Col,
				"Too many scale factors "+
					"",
			)
	}
	op = n.Ty
	// if minus then also unread it so the first expression value is negative
	if op == lx.TOKEN_TMINUS {
		p.lexer.UnreadToken(lx.Token{ Ty: lx.TOKEN_TMINUS })
	}
	if e, err := p.parseExpression(0); err != nil {
		return nil, err
	} else {
		ret.Disp = e
		// ret.DispOp = op
		return &Expr{
			Val: ret,
		}, nil
	}
}

func makeBinop(lhs, rhs *Expr, opToken lx.Token) (*Expr, error) {
	if lhs.IsRegexpr() ||
		rhs.IsRegexpr() ||
		lhs.IsMemexpr() ||
		rhs.IsMemexpr() {
		return nil, errors.
			MakeParserError(
				lhs.Line,
				lhs.Col,
				"Disallowed binary expression",
			)

	}
	return &Expr{
		Val: ArthExpr{
			opTyToArthExpr(lhs, rhs, opToken.Ty),
		},
	}, nil
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
