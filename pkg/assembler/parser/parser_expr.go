package assembler

import (
	"fmt"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
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
			return inner, errors.FailedToParse("dereference expression",
				t.Line, t.Col,
				"Unclosed bracket in dereference expression, got `%s` instead",
				t.ForceValAsString(),
			)
		}
		deref := MakeDeref(inner)
		return deref, nil
	// case lx.TOKEN_TREG:
	// 	// this wont be an expression, so stop parsing here
	// 	p.lexer.UnreadCurrentToken()
	// 	return nil, nil
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
		// else if rhs == nil {
		// 	// in case a register was hit, unread the operator
		// 	p.lexer.UnreadToken(op)
		// 	return lhs, nil
		// }
		if e, err := makeBinop(lhs, rhs, op); err != nil {
			return nil, err
		} else {
			lhs = e
		}
		// lhs = &Expr{
		// 	Val: ArthExpr{
		// 		Val: opTyToArthExpr(lhs, rhs, op.Ty),
		// 	},
		// }
	}
	return lhs, nil
}

func makeBinop(lhs, rhs *Expr, opToken lx.Token) (*Expr, error) {
	if lhs.IsRegexpr() || rhs.IsRegexpr() {
		if e, err := opTyToRegOffExpr(lhs, rhs, opToken.Ty); err != nil {
			return nil, err
		} else {
			return &Expr{
				Val: e,
			}, nil
		}
	}
	if lhs.IsOneRegOffsetExpr() || rhs.IsOneRegOffsetExpr() {
		if e, err := opTyWithOneOffRegExpr(lhs, rhs, opToken.Ty); err != nil {
			return nil, err
		} else {
			return &Expr{
				Val: e,
			}, nil
		}
	}
	if lhs.IsTwoRegOffsetExpr() {
		if e, err := opTyWithTwoOffRegExpr(lhs, rhs, opToken.Ty); err != nil {
			return nil, err
		} else {
			return &Expr{
				Val: e,
			}, nil
		}
	} else if rhs.IsTwoRegOffsetExpr() {
		if e, err := opTyWithTwoOffRegExpr(lhs, rhs, opToken.Ty); err != nil {
			return nil, err
		} else {
			return &Expr{
				Val: e,
			}, nil
		}
	}
	return &Expr{
		Val: ArthExpr{
			opTyToArthExpr(lhs, rhs, opToken.Ty),
		},
	}, nil
}
func arthOrPass(lhs, rhs *Expr, ty int) *Expr {
	if lhs == nil {
		return rhs
	} else if rhs == nil {
		return lhs
	} else {
		return &Expr{
			Val: ArthExpr{
				Val: opTyToArthExpr(lhs, rhs, ty),
			},
		}
	}
}
func opTyWithTwoOffRegExpr(lhs, rhs *Expr, ty int) (any, error) {
	switch a := lhs.Val.(type) {
	case TwoRegOffsetExpr:
		switch rhs.Val.(type) {
		case ConstExpr:
			return TwoRegOffsetExpr{
				Reg1:     a.Reg1,
				Reg2:     a.Reg2,
				RegOp:    a.RegOp,
				OffsetOp: ty,
				Offset:   arthOrPass(a.Offset, rhs, ty),
			}, nil
		case ArthExpr:
			return TwoRegOffsetExpr{
				Reg1:     a.Reg1,
				Reg2:     a.Reg2,
				RegOp:    a.RegOp,
				OffsetOp: ty,
				Offset:   arthOrPass(a.Offset, rhs, ty),
			}, nil
		default:
			return nil, errors.FailedToParse(
				"two register offset expression",
				rhs.Line, rhs.Col,
				"Disallowed operation",
			)
		}
	case ConstExpr:
		switch b := rhs.Val.(type) {
		case TwoRegOffsetExpr:
			if IsAssoc(ty) && ty == b.OffsetOp {
				return TwoRegOffsetExpr{
					Reg1:     b.Reg1,
					Reg2:     b.Reg2,
					RegOp:    b.RegOp,
					OffsetOp: b.OffsetOp,
					Offset:   arthOrPass(lhs, b.Offset, ty),
				}, nil
			}
			return nil, errors.FailedToParse(
				"two register offset expression",
				rhs.Line, rhs.Col,
				"Disallowed operation",
			)
		default:
			return nil, errors.FailedToParse(
				"two register offset expression",
				rhs.Line, rhs.Col,
				"Disallowed operation",
			)
		}
	case ArthExpr:
		switch b := rhs.Val.(type) {
		case TwoRegOffsetExpr:
			if IsAssoc(ty) && ty == b.OffsetOp {
				return TwoRegOffsetExpr{
					Reg1:     b.Reg1,
					Reg2:     b.Reg2,
					RegOp:    b.RegOp,
					OffsetOp: b.OffsetOp,
					Offset:   arthOrPass(lhs, b.Offset, ty),
				}, nil
			}
			return nil, errors.FailedToParse(
				"two register offset expression",
				rhs.Line, rhs.Col,
				"Disallowed operation",
			)
		default:
			return nil, errors.FailedToParse(
				"two register offset expression",
				rhs.Line, rhs.Col,
				"Disallowed operation",
			)
		}
	}
	return nil, errors.FailedToParse(
		"two register offset expression",
		rhs.Line, rhs.Col,
		"Disallowed operation",
	)
}

func opTyWithOneOffRegExpr(lhs, rhs *Expr, ty int) (any, error) {
	switch a := lhs.Val.(type) {
	case OneRegOffsetExpr:
		switch b := rhs.Val.(type) {
		case RegExpr:
			if allowedBetweenRegs(ty) && allowedBetweenRegs(a.OffsetOp) {
				return TwoRegOffsetExpr{
					Reg1:     a.Reg,
					Reg2:     b.Reg,
					RegOp:    ty,
					OffsetOp: a.OffsetOp,
					Offset:   arthOrPass(a.Offset, rhs, ty),
				}, nil
			} else {
				return nil, errors.FailedToParse(
					"two register offset expression",
					rhs.Line, rhs.Col,
					"Disallowed operation between registers",
				)
			}
		default:
			return OneRegOffsetExpr{
				Reg:      a.Reg,
				OffsetOp: a.OffsetOp,
				Offset:   arthOrPass(a.Offset, rhs, ty),
			}, nil
		}
	case RegExpr:
		switch b := rhs.Val.(type) {
		case OneRegOffsetExpr:
			if allowedBetweenRegs(ty) && allowedBetweenRegs(b.OffsetOp) {
				return TwoRegOffsetExpr{
					Reg1:     a.Reg,
					Reg2:     b.Reg,
					RegOp:    ty,
					OffsetOp: b.OffsetOp,
					Offset:   arthOrPass(lhs, b.Offset, ty),
				}, nil
			} else {
				return nil, errors.FailedToParse(
					"two register offset expression",
					rhs.Line, rhs.Col,
					"Disallowed operation between registers",
				)
			}
		default:
			return nil, errors.FailedToParse(
				"two register offset expression",
				rhs.Line, rhs.Col,
				"Disallowed expression operands",
			)
		}
	case ConstExpr:
		switch b := rhs.Val.(type) {
		case OneRegOffsetExpr:
			if IsAssoc(ty) && ty == b.OffsetOp {
				return OneRegOffsetExpr{
					Reg:      b.Reg,
					OffsetOp: b.OffsetOp,
					Offset:   arthOrPass(lhs, b.Offset, ty),
				}, nil
			} else {
				return nil, errors.FailedToParse(
					"two register offset expression",
					rhs.Line, rhs.Col,
					"Disallowed operation between registers",
				)
			}
		}
	}
	return nil, errors.FailedToParse(
		"two register offset expression",
		rhs.Line, rhs.Col,
		"Bad expression",
	)
}

func opTyToRegOffExpr(lhs, rhs *Expr, ty int) (any, error) {
	switch a := lhs.Val.(type) {
	case RegExpr:
		if rhs.IsRegexpr() {
			if ty != lx.TOKEN_TMINUS && ty != lx.TOKEN_TPLUS {
				return nil, errors.FailedToParse(
					"two register offset expression",
					rhs.Line, rhs.Col,
					"Disallowed operation between registers",
				)
			}
			return TwoRegOffsetExpr{
				Reg1:   a.Reg,
				Reg2:   rhs.Val.(RegExpr).Reg,
				RegOp:  ty,
				Offset: nil,
			}, nil
		} else {
			return OneRegOffsetExpr{
				Reg:      a.Reg,
				OffsetOp: ty,
				Offset:   rhs,
			}, nil
		}
	case ConstExpr:
		if !IsAssoc(ty) {
			return nil, errors.FailedToParse(
				"single register offset expression",
				lhs.Line, lhs.Col,
				"Disallowed operator between register and offset",
			)
		}
		return OneRegOffsetExpr{
			Reg:      rhs.Val.(RegExpr).Reg,
			OffsetOp: ty,
			Offset:   lhs,
		}, nil
	}
	return nil, nil
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
