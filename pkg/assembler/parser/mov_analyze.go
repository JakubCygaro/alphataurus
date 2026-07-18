package assembler

import (
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
	"github.com/JakubCygaro/alphataurus/pkg/assembler/parser/errors"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func getAsOffset(expr *Expr) (int64, bool) {
	if expr == nil {
		return 0, true
	}
	c, ok := IsConstexprType[ConstExprILit](expr)
	return c.Signed(), ok
}

// mov rx, rx
func movRR(dest, src RegExpr, mov *InstGenericMov) (any, error) {
	if mov.DataSize != nil {
		return nil,
			errors.
				UnnecessarySizeParameter(
					mov.DataSize.Line,
					mov.DataSize.Col,
					mov.DataSize.Size,
				)
	}
	if !vm.IsMovFromRAllowed(byte(src.Reg.Reg)) {
		return nil,
			errors.
				DisallowedSrcReg(
					mov.Src.Line,
					mov.Src.Col,
					src.Reg,
				)
	}
	if src.Reg.Size > dest.Reg.Size {
		return nil,
			errors.
				MismatchedRegisterSizes(
					mov.Dest.Line, mov.Dest.Col,
					mov.Src.Line, mov.Src.Col,
					src.Reg,
					dest.Reg,
				)
	}
	return InstMovRR{
		Src:  src.Reg,
		Dest: dest.Reg,
	}, nil
}

// mov rx, int/float
func movRI(dest RegExpr, imm ConstExpr, mov *InstGenericMov) (any, error) {
	if mov.DataSize != nil {
		return nil,
			errors.
				UnnecessarySizeParameter(
					mov.DataSize.Line,
					mov.DataSize.Col,
					mov.DataSize.Size,
				)
	}
	switch cexpr := imm.Val.(type) {
	// mov rx, 1
	case ConstExprILit:
		return InstMovIR{
			Imm:      cexpr.Integer,
			Dest:     dest.Reg,
			DataSize: dest.Reg.Size,
		}, nil
	// mov rx, 1.0
	case ConstExprFLit:
		if dest.Reg.Size != vm.SZ_64 {
			return nil,
				errors.MakeParserError(
					mov.DataSize.Line,
					mov.DataSize.Col,
					"Bad data size parameter, expected 64-bits for float literal",
				)
		}
		return InstMovIR{
			Imm:      cexpr.Float,
			Dest:     dest.Reg,
			DataSize: dest.Reg.Size,
		}, nil
	}
	return nil, MakeParserErrorWithExpr(
		mov.Src,
		func(s string) errors.ParserError {
			return errors.
				MakeParserError(
					mov.Src.Line,
					mov.Src.Col,
					"Bad source expression in immediate value move into register `%s`.",
					s,
				)
		},
	)
}

// mov rx, [(inner)]
func movDR(
	dest RegExpr,
	deref DerefExpr,
	mov *InstGenericMov,
) (any, error) {
	switch innerSrc := deref.Inner.Val.(type) {
	// mov rx, [1]
	case ConstExpr:
		if ilit, ok := innerSrc.Val.(ConstExprILit); ok {
			return InstMovDR{
				Address: ilit.Signed(),
				Dest:    dest.Reg,
			}, nil
		}
	// mov rx, [rx]
	case RegExpr:
		return InstMovDRO1{
			Dest:   dest.Reg,
			Offset: 0,
			OReg1:  innerSrc.Reg,
			OffOp:  lx.TOKEN_TPLUS,
		}, nil
	// mov rx, [rx + 1]
	case OneRegOffsetExpr:
		if o, ok := getAsOffset(innerSrc.Offset); ok {
			return InstMovDRO1{
				Dest:   dest.Reg,
				Offset: o,
				OReg1:  innerSrc.Reg,
				OffOp:  innerSrc.OffsetOp,
			}, nil
		}
		return nil, MakeParserErrorWithExpr(
			innerSrc.Offset,
			func(s string) errors.ParserError {
				return errors.
					MakeParserError(
						innerSrc.Offset.Line,
						innerSrc.Offset.Col,
						"Bad offset expression for mov instruction `%s`.",
						s,
					)
			},
		)
	// mov rx, [rx + rx + 1]
	case TwoRegOffsetExpr:
		if o, ok := getAsOffset(innerSrc.Offset); ok {
			return InstMovDRO2{
				Dest:   dest.Reg,
				OReg1:  innerSrc.Reg1,
				RegOp:  innerSrc.RegOp,
				OReg2:  innerSrc.Reg2,
				Offset: o,
			}, nil
		}
		return nil, MakeParserErrorWithExpr(
			innerSrc.Offset,
			func(s string) errors.ParserError {
				return errors.
					MakeParserError(
						innerSrc.Offset.Line,
						innerSrc.Offset.Col,
						"Bad offset expression for mov instruction `%s`.",
						s,
					)
			},
		)
	}
	return nil, MakeParserErrorWithExpr(
		mov.Src,
		func(s string) errors.ParserError {
			return errors.
				MakeParserError(
					mov.Src.Line,
					mov.Src.Col,
					"Bad source expression for dereference move into register `%s`.",
					s,
				)
		},
	)
}

// mov rx, (expr)
func movIntoRegister(dest RegExpr, mov *InstGenericMov) (any, error) {
	if !vm.IsMovIntoRAllowed(byte(dest.Reg.Reg)) {
		return nil,
			errors.
				DisallowedDestReg(
					mov.Dest.Line,
					mov.Dest.Col,
					dest.Reg,
				)
	}
	switch src := mov.Src.Val.(type) {
	case RegExpr:
		return movRR(dest, src, mov)
	case ConstExpr:
		return movRI(dest, src, mov)
	case DerefExpr:
		return movDR(dest, src, mov)
	}
	return nil, MakeParserErrorWithExpr(
		mov.Dest,
		func(s string) errors.ParserError {
			return errors.
				DisallowedDest(
					mov.Dest.Line,
					mov.Dest.Col,
					s,
				)
		},
	)
}
func movRD(
	innerDest ConstExpr,
	mov *InstGenericMov,
	outer *Instruction,
) (any, error) {
	var off int64
	if cexpr, ok := innerDest.Val.(ConstExprILit); !ok {
		return nil, MakeParserErrorWithExpr(
			mov.Dest,
			func(s string) errors.ParserError {
				return errors.
					MakeParserError(
						mov.Dest.Line,
						mov.Dest.Line,
						"Bad address expression in dereference"+
							" destination for move instruction. In expression `%s`.",
						s,
					)
			},
		)
	} else {
		off = cexpr.Signed()
	}
	switch src := mov.Src.Val.(type) {
	// mov [1], rx
	case RegExpr:
		if mov.DataSize != nil {
			return nil,
				errors.
					UnnecessarySizeParameter(
						mov.DataSize.Line,
						mov.DataSize.Col,
						mov.DataSize.Size,
					)
		}
		return InstMovRD{
			Src:     src.Reg,
			Address: off,
		}, nil
		// mov [1], 1
	case ConstExpr:
		if cexpr, ok := src.Val.(ConstExprILit); ok {
			if mov.DataSize == nil {
				return nil,
					errors.
						MissingSizeParameter(
							outer.Line,
							outer.Col,
						)
			}
			return InstMovID{
				Imm:      cexpr.Integer,
				Address:  off,
				DataSize: mov.DataSize.Size,
			}, nil
		}
	}
	return nil, MakeParserErrorWithExpr(
		mov.Src,
		func(s string) errors.ParserError {
			return errors.
				MakeParserError(
					mov.Src.Line,
					mov.Src.Col,
					"Bad source expression for move into dereference `%s`.",
					s,
				)
		},
	)
}
func movRDO1_NO(
	innerDest RegExpr,
	mov *InstGenericMov,
	outer *Instruction,
) (any, error) {
	switch src := mov.Src.Val.(type) {
	// mov [rx], 1
	case ConstExpr:
		if imm, ok := IsConstexprType[ConstExprILit](mov.Src); ok {
			if mov.DataSize == nil {
				return nil,
					errors.
						MissingSizeParameter(
							outer.Line,
							outer.Col,
						)
			}
			return InstMovIDO1{
				Imm:      imm.Integer,
				Offset:   0,
				OReg1:    innerDest.Reg,
				OffOp:    vm.OP_TADD,
				DataSize: mov.DataSize.Size,
				NoOff:    true,
			}, nil
		}
	// mov [rx], rx
	case RegExpr:
		return InstMovRDO1{
			Src:    src.Reg,
			Offset: 0,
			OReg1:  innerDest.Reg,
			OffOp:  vm.OP_TADD,
		}, nil
	}
	return nil, MakeParserErrorWithExpr(
		mov.Src,
		func(s string) errors.ParserError {
			return errors.
				MakeParserError(
					mov.Src.Line,
					mov.Src.Col,
					"Bad source expression for move into register dereference"+
						"`%s`.",
					s,
				)
		},
	)
}
func movRDO1(
	innerDest OneRegOffsetExpr,
	mov *InstGenericMov,
	outer *Instruction,
) (any, error) {
	var off int64
	if o, ok := getAsOffset(innerDest.Offset); !ok {
		return nil, MakeParserErrorWithExpr(
			innerDest.Offset,
			func(s string) errors.ParserError {
				return errors.
					MakeParserError(
						innerDest.Offset.Line,
						innerDest.Offset.Col,
						"Bad offset expresion in move into one register dereference"+
							" `%s`",
						s,
					)
			},
		)
	} else {
		off = o
	}
	switch src := mov.Src.Val.(type) {
	// mov [rx+1], 1
	case ConstExpr:
		if imm, ok := IsConstexprType[ConstExprILit](mov.Src); ok {
			if mov.DataSize == nil {
				return nil,
					errors.
						MissingSizeParameter(
							outer.Line,
							outer.Col,
						)
			}
			return InstMovIDO1{
				Imm:      imm.Integer,
				Offset:   off,
				OReg1:    innerDest.Reg,
				OffOp:    innerDest.OffsetOp,
				DataSize: mov.DataSize.Size,
				NoOff:    off == 0,
			}, nil
		}
	// mov [rx+1], rx
	case RegExpr:
		return InstMovRDO1{
			Src:    src.Reg,
			Offset: off,
			OReg1:  innerDest.Reg,
			OffOp:  innerDest.OffsetOp,
		}, nil
	}
	return nil, MakeParserErrorWithExpr(
		mov.Src,
		func(s string) errors.ParserError {
			return errors.
				MakeParserError(
					mov.Src.Line,
					mov.Src.Col,
					"Bad source expression for one register dereference move "+
						"`%s`.",
					s,
				)
		},
	)
}
func movRDO2(
	innerDest TwoRegOffsetExpr,
	mov *InstGenericMov,
	outer *Instruction,
) (any, error) {
	var off int64
	if o, ok := getAsOffset(innerDest.Offset); !ok {
		return nil, MakeParserErrorWithExpr(
			innerDest.Offset,
			func(s string) errors.ParserError {
				return errors.
					MakeParserError(
						innerDest.Offset.Line,
						innerDest.Offset.Col,
						"Bad offset expresion in move into two register dereference"+
							" `%s`",
						s,
					)
			},
		)
	} else {
		off = o
	}
	switch src := mov.Src.Val.(type) {
	// mov [rx+rx+1], 1
	case ConstExpr:
		if imm, ok := IsConstexprType[ConstExprILit](mov.Src); ok {
			if mov.DataSize == nil {
				return nil,
					errors.
						MissingSizeParameter(
							outer.Line,
							outer.Col,
						)
			}
			if off == 0 {
				innerDest.RegOp = lx.TOKEN_TPLUS
			}
			return InstMovIDO2{
				Imm:      imm.Integer,
				DataSize: mov.DataSize.Size,
				OReg1:    innerDest.Reg1,
				OReg2:    innerDest.Reg2,
				RegOp:    innerDest.RegOp,
				Offset:   off,
				NoOff:    off == 0,
			}, nil
		}
	// mov [rx+rx+1], rx
	case RegExpr:
		return InstMovRDO2{
			Src:    src.Reg,
			Offset: off,
			OReg1:  innerDest.Reg1,
			OReg2:  innerDest.Reg2,
			RegOp:  innerDest.RegOp,
			OffOp:  innerDest.OffsetOp,
		}, nil
	}
	return nil, MakeParserErrorWithExpr(
		mov.Src,
		func(s string) errors.ParserError {
			return errors.
				MakeParserError(
					mov.Src.Line,
					mov.Src.Col,
					"Bad source expression for move into two "+
						"register dereference with offset "+
						"`%s`.",
					s,
				)
		},
	)
}

func movIntoDeref(
	dest DerefExpr,
	mov *InstGenericMov,
	outer *Instruction,
) (any, error) {
	switch innerDest := dest.Inner.Val.(type) {
	case ConstExpr:
		return movRD(innerDest, mov, outer)
	case RegExpr:
		return movRDO1_NO(innerDest, mov, outer)
	case OneRegOffsetExpr:
		if innerDest.Offset == nil {
			return movRDO1_NO(RegExpr{
				Reg: innerDest.Reg,
			}, mov, outer)
		} else {
			return movRDO1(innerDest, mov, outer)
		}
	case TwoRegOffsetExpr:
		return movRDO2(innerDest, mov, outer)
	}
	return nil, MakeParserErrorWithExpr(
		dest.Inner,
		func(s string) errors.ParserError {
			return errors.
				MakeParserError(
					outer.Line,
					outer.Line,
					"Bad destination expression for move into dereference.",
				)
		},
	)
}

func GetConcreteMovInst(
	genericMov InstGenericMov,
	outer *Instruction,
) (any, error) {
	switch dest := genericMov.Dest.Val.(type) {
	case RegExpr:
		return movIntoRegister(dest, &genericMov)
	case DerefExpr:
		return movIntoDeref(dest, &genericMov, outer)
	}
	return nil, MakeParserErrorWithExpr(
		genericMov.Dest,
		func(s string) errors.ParserError {
			return errors.MakeParserError(
				outer.Line,
				outer.Col,
				"Bad mov instruction, destination parameter of unsupported type."+
					" In expression `%s`.",
				s,
			)
		},
	)
}
