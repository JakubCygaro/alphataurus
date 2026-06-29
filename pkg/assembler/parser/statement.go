package assembler

import (
	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
)

type Stmt struct {
	Val any
}

type RegStmt struct {
	Reg lx.RegisterData
}
type OneRegOffsetStmt struct {
	Reg      lx.RegisterData
	OffsetOp int
	Offset   *Expr
}
type TwoRegOffsetStmt struct {
	Reg1, Reg2      lx.RegisterData
	OffsetOp, RegOp int
	Offset          *Expr
}

func (p *Parser) ParseRegisterStmt(offset *Expr) (*Stmt, error) {
	var first lx.Token
	if t, err := p.lexer.ReadNextTokenReturn(); err != nil {
		return nil, err
	} else {
		first = t
	}

	switch {
	case first.Ty == lx.TOKEN_TREG:
		ors := OneRegOffsetStmt{
			Reg: first.Val.(lx.RegisterData),
			OffsetOp: PENDING,
		}
		return p.parseWithOneRegisterStmt(ors)
	default:
		p.lastBOpT = lx.Token{ Ty: INVALID }
		if e, err := p.ParseExpression(); err != nil {
			return nil, err
		} else if e == nil {
			return nil, errors.FailedToParse(
				"Register offset statement",
				first.Line, first.Col,
				"Invalid token at `%s`", first.ForceValAsString(),
			)
		} else if op, ok := p.lexer.ReadNextTokenReturn(); !IsAssoc(op) {
			return nil, errors.FailedToParse(
				"Register offset statement",
				first.Line, first.Col,
				"Disallowed offset operator `%s`", op.ForceValAsString(),
			)
		} else if reg, ok := p.lexer.ReadNextTokenReturn(); reg.Ty != TOKEN_TREG {
			return nil, errors.FailedToParse(
				"Register offset statement",
				first.Line, first.Col,
				"Expected register, got `%s`", reg.ForceValAsString(),
			)
		} else {
			ors := OneRegOffsetStmt{
				Reg: first.Val.(lx.RegisterData),
				OffsetOp: op.Ty,
				Offset: e,
			}
			return p.parseWithOneRegisterStmt(ors)
		}

	}
}
func (p *Parser) parseWithOneRegisterStmt(ors OneRegOffsetStmt) (*Stmt, error) {
	var first lx.Token
	if t, err := p.lexer.ReadNextTokenReturn(); err != nil {
		return nil, err
	} else {
		first = t
	}
	if !IsInfixOp(first.Ty) {
		return nil, errors.FailedToParse(
			"One Register offset statement",
			first.Line, first.Col,
			"Invalid statement token `%s`", first.ForceValAsString(),
		)
	}
	op := first
	var next lx.Token
	if n, err := p.lexer.ReadNextTokenReturn(); err != nil {
		return nil, err
	} else {
		next = n
	}
	switch next.Ty {
	case lx.TOKEN_TREG:
		trs := TwoRegOffsetStmt {
			Reg1: ,
		}
	}
}
