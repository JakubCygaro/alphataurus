package assembler

import (
	"bufio"
	"fmt"

	"github.com/JakubCygaro/alphataurus/assembler/errors"
	"github.com/JakubCygaro/alphataurus/internal/vm"
)

const (
	INST_TMOVRR = iota
	INST_TMOVIR
	INST_TMOVDRI
	INST_TMOVDRO1
	INST_TADDRR
	INST_TSUBRR
	INST_TDIVRR
	INST_TMULRR
	INST_TADDIR
	INST_TSUBIR
	INST_TINCR
	INST_TDECR
	INST_TCMPRR
	INST_TCMPIR
	INST_TJMP
	INST_TJMPE
	INST_TJMPZ
	INST_TJMPNE
	INST_TJMPNZ
	INST_TJMPG
	INST_TJMPGE
	INST_TJMPL
	INST_TJMPLE
	INST_TLABEL
	INST_TPUSHR
	INST_TPUSHI
	INST_TPOP
)

type InstMovData struct {
	Src, Dest int
	Imm       uint64
}

const (
	ARTH_TUNSIGNED = vm.TY_UINT64
	ARTH_TSIGNED   = vm.TY_INT64
	ARTH_TFLOAT    = vm.TY_FLOAT64
)

const (
	ARTH_TADD = iota
	ARTH_TSUB
	ARTH_TDIV
	ARTH_TMUL
)

type InstIncDecData struct {
	Reg int
}
type InstArthData struct {
	Src, Dest int
	Ty        int
	Imm       uint64
}
type InstCmpData struct {
	Ty       int
	Sub, Min int
	Imm      uint64
}
type InstJmpData struct {
	Address any
}
type InstLabData struct {
	Label      string
	DeclaredAt string
}
type PushPopData struct {
	Reg uint64
	Imm uint64
}
type InstDerefMovData struct {
	Dest   int
	Offset int64
	OReg1  int
	Label  string
	OpTy   int
}
type Instruction struct {
	Ty   int
	Data any
}
type Parser struct {
	lexer        Lexer
	currentInst  Instruction
	currentIdent string
}

func NewParser(reader bufio.Reader) Parser {
	return Parser{
		lexer: NewLexer(reader),
	}
}

func (p *Parser) CurrentInst() Instruction {
	return p.currentInst
}
func (p *Parser) ParseNext() (bool, error) {
	var err error = nil
	var start Token
	for {
		err = p.lexer.ReadNextToken()
		if err != nil {
			return false, err
		}
		start = p.lexer.CurrentToken()
		if start.Ty == TOKEN_TEOF {
			return false, nil
		}
		if start.Ty != TOKEN_TNEWLINE {
			break
		}
	}
	switch start.Ty {
	case TOKEN_TIDENT:
		err = p.parseStartIdent(start)
		if err != nil {
			return false, err
		}
	default:
		return false, fmt.Errorf("Unimplemented instruction %s", p.lexer.CurrentPosition())
	}
	err = p.lexer.ReadNextToken()
	if p.lexer.CurrentToken().Ty != TOKEN_TNEWLINE && p.lexer.CurrentToken().Ty != TOKEN_TEOF {
		return false, errors.ExtraTokensOnLine(p.lexer.line, p.lexer.col)
	}
	return true, err
}
func (p *Parser) parseStartIdent(t Token) error {
	ident := t.val.(string)
	p.currentIdent = ident
	switch ident {
	case "mov":
		return p.parseMov()
	case "add":
		return p.parseAddOrSub(ARTH_TADD)
	case "sub":
		return p.parseAddOrSub(ARTH_TSUB)
	case "div":
		return p.parseDivOrMul(ARTH_TDIV)
	case "mul":
		return p.parseDivOrMul(ARTH_TMUL)
	case "inc":
		return p.parseInc()
	case "dec":
		return p.parseDec()
	case "cmp":
		return p.parseCmp()
	case "jmp":
		return p.parseJmp(INST_TJMP)
	case "je":
		return p.parseJmp(INST_TJMPE)
	case "jne":
		return p.parseJmp(INST_TJMPNE)
	case "jz":
		return p.parseJmp(INST_TJMPZ)
	case "jnz":
		return p.parseJmp(INST_TJMPNZ)
	case "jg":
		return p.parseJmp(INST_TJMPG)
	case "jge":
		return p.parseJmp(INST_TJMPGE)
	case "jl":
		return p.parseJmp(INST_TJMPL)
	case "jle":
		return p.parseJmp(INST_TJMPLE)
	case "push":
		return p.parsePush()
	case "pop":
		return p.parsePop()
	default:
		pos := p.lexer.CurrentPosition()
		if _, ok := p.lexer.Expect(TOKEN_TCOLON); ok {
			p.currentInst = Instruction{
				Ty: INST_TLABEL,
				Data: InstLabData{
					Label:      ident,
					DeclaredAt: pos,
				},
			}
			return nil
		}
	}
	p.currentIdent = ""
	return errors.UnknownIdentifier(ident, p.lexer.line, p.lexer.col)
}
func (p *Parser) parseMov() error {
	var op1 Token
	if expr, err := p.parseExpression(0); err != nil {
		return err
	} else if eval, _ := TryConstEvaluateExpression(&expr); eval.Ty != CONSTEXPR_TREG {
		return errors.FailedToParse("mov instruction",
			"First operand to instruction must be a valid register",
			p.lexer.line, p.lexer.col)
	} else {
		op1.val = int(expr.Val.(ConstExpr).Val)
	}
	if err := p.lexer.ReadNextToken(); err != nil {
		return err
	}
	comma := p.lexer.CurrentToken()
	if comma.Ty != TOKEN_TCOMMA {
		return errors.FailedToParse("mov instruction",
			"Instruction missing a comma",
			p.lexer.line, p.lexer.col)
	}

	var op2 ConstExpr
	if expr, err := p.parseExpression(0); err != nil {
		return err
	} else {
		eval, _ := TryEvaluateExpression(&expr)
		switch eval.Ty {
		case EXPR_TCONST:
			op2 = eval.Val.(ConstExpr)
		case EXPR_TDEREF:
			return p.parseDerefMov(op1.val.(int), eval.Val.(DerefExpr).Inner)
		default:
			return errors.FailedToParse("mov instruction",
				"Second operand to instruction has to be a valid register, dereference, or a compile time expression",
				p.lexer.line, p.lexer.col)
		}
	}
	switch op2.Ty {
	case CONSTEXPR_TREG:
		p.currentInst = Instruction{
			Ty: INST_TMOVRR,
			Data: InstMovData{
				Src:  int(op2.Val),
				Dest: op1.val.(int),
			},
		}
	case CONSTEXPR_TILIT:
		p.currentInst = Instruction{
			Ty: INST_TMOVIR,
			Data: InstMovData{
				Imm:  op2.Val,
				Dest: op1.val.(int),
			},
		}
	case CONSTEXPR_TFLIT:
		p.currentInst = Instruction{
			Ty: INST_TMOVIR,
			Data: InstMovData{
				Imm:  op2.Val,
				Dest: op1.val.(int),
			},
		}
	default:
		return errors.FailedToParse("mov instruction",
			"Second operand to instruction has to be a valid register, dereference or a compile time expression",
			p.lexer.line, p.lexer.col)
	}
	return nil
}

const (
	DEREF_T0RO = iota
	DEREF_T1RO
	DEREF_T2RO
)

type DerefData struct {
	Ty         int
	Reg1, Reg2 int
	// plus or minus
	OffsetOp int
	Offset   int64
	// in case there are labels to resolve
	OffsetExpr *Expr
}

func (p *Parser) processDeref(inner Expr) (DerefData, error) {
	ret := DerefData{
		Reg1:       INVALID,
		Reg2:       INVALID,
		OffsetOp:   INVALID,
		Offset:     INVALID,
		OffsetExpr: nil,
	}
	switch inner.Ty {
	case EXPR_TCONST:
		innerConst := inner.Val.(ConstExpr)
		switch innerConst.Ty {
		case CONSTEXPR_TILIT:
			ret.Ty = DEREF_T0RO
			ret.Offset = int64(innerConst.Val)
		case CONSTEXPR_TREG:
			ret.Ty = DEREF_T1RO
			ret.Reg1 = int(innerConst.Val)
			ret.Offset = int64(0)
			ret.OffsetOp = vm.OP_TADD
		// TODO: label dereference support
		default:
			return ret, errors.FailedToParse("dereference expression",
				"Invalid dereference expression parameter",
				p.lexer.line, p.lexer.col)
		}
	case EXPR_TARTH:
		arthExpr := inner.Val.(ArthExpr)
		switch {
		case IsConstexpr(&arthExpr.A, CONSTEXPR_TREG) && IsConstexpr(&arthExpr.B, CONSTEXPR_TILIT):
			ret.Ty = DEREF_T1RO
			ret.Reg1 = int(arthExpr.A.Val.(ConstExpr).Val)
			ret.Offset = int64(arthExpr.B.Val.(ConstExpr).Val)
			ret.OffsetOp = arthExpr.GetVMOpType()
		case IsConstexpr(&arthExpr.A, CONSTEXPR_TILIT) && IsConstexpr(&arthExpr.B, CONSTEXPR_TREG) &&
			(arthExpr.Ty == ARTHEXPR_TADD):
			ret.Ty = DEREF_T1RO
			ret.Reg1 = int(arthExpr.A.Val.(ConstExpr).Val)
			ret.Offset = int64(arthExpr.B.Val.(ConstExpr).Val)
			ret.OffsetOp = arthExpr.GetVMOpType()
		default:
			return ret, errors.FailedToParse("mov instruction",
				"Invalid dereference expression parameter",
				p.lexer.line, p.lexer.col)
			// TODO: label dereference support
		}
	default:
		return ret, errors.FailedToParse("mov instruction",
			"Invalid dereference expression parameter",
			p.lexer.line, p.lexer.col)
	}
	return ret, nil
}
func (p *Parser) parseDerefMov(reg int, inner Expr) error {
	dData, err := p.processDeref(inner)
	if err != nil {
		return err
	}
	switch dData.Ty {
	case DEREF_T0RO:
		p.currentInst = Instruction{
			Ty: INST_TMOVDRI,
			Data: InstDerefMovData{
				Dest:   reg,
				Offset: dData.Offset,
			},
		}
	case DEREF_T1RO:
		p.currentInst = Instruction{
			Ty: INST_TMOVDRO1,
			Data: InstDerefMovData{
				Dest:   reg,
				OReg1:  dData.Reg1,
				Offset: dData.Offset,
				OpTy:   dData.OffsetOp,
			},
		}
	default:
		return errors.FailedToParse("mov instruction",
			"Invalid dereference expression parameter",
			p.lexer.line, p.lexer.col)
	}
	return nil
}
