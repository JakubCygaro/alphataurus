package assembler

import (
	// lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
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
func getAsScale(expr *Expr) (Scale, bool) {
	if expr == nil {
		return SCALE_0, true
	}
	ret := SCALE_0
	c, ok := IsConstexprType[ConstExprILit](expr)
	if !ok {
		return ret, ok
	}
	switch c.Signed() {
	case 2:
		ret = SCALE_2
	case 4:
		ret = SCALE_4
	case 8:
		ret = SCALE_8
	default:
		ok = false
	}
	return ret, ok
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
			Base:   innerSrc.Reg,
			DispOp: lx.TOKEN_TPLUS,
			Offset: 0,
			SF:     SCALE_0,
		}, nil
	case MemExpr:
		var scale Scale
		if s, ok := getAsScale(innerSrc.ScaleF); !ok {
			return nil, errors.
				MakeParserError(
					innerSrc.ScaleF.Line,
					innerSrc.ScaleF.Col,
					"Invalid scale factor")
		} else {
			scale = s
		}
		var offset int64
		if o, ok := getAsOffset(innerSrc.Disp); !ok {
			return nil, MakeParserErrorWithExpr(
				innerSrc.Disp,
				func(s string) errors.ParserError {
					return errors.
						MakeParserError(
							innerSrc.Disp.Line,
							innerSrc.Disp.Col,
							"Bad offset expression for mov instruction `%s`.",
							s,
						)
				},
			)
		} else {
			offset = o
		}
		if innerSrc.Index.HasVal() {
			return InstMovDRO2{
				Dest:   dest.Reg,
				Base:   innerSrc.Base,
				RegOp:  innerSrc.RegOp,
				Index:  innerSrc.Index.Get(),
				Disp:   offset,
				DispOp: innerSrc.DispOp,
				SF:     scale,
			}, nil
		} else {
			return InstMovDRO1{
				Dest:   dest.Reg,
				Base:   innerSrc.Base,
				DispOp: innerSrc.DispOp,
				Offset: offset,
				SF:     scale,
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
				Disp:     0,
				Base:     innerDest.Reg,
				DispOp:   vm.OP_TADD,
				DataSize: mov.DataSize.Size,
				NoOff:    true,
			}, nil
		}
	// mov [rx], rx
	case RegExpr:
		return InstMovRDO1{
			Src:    src.Reg,
			Disp:   0,
			Base:   innerDest.Reg,
			DispOp: vm.OP_TADD,
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

// func movRDO1(
// 	innerDest OneRegOffsetExpr,
// 	mov *InstGenericMov,
// 	outer *Instruction,
// ) (any, error) {
// 	var off int64
// 	if o, ok := getAsOffset(innerDest.Disp); !ok {
// 		return nil, MakeParserErrorWithExpr(
// 			innerDest.Disp,
// 			func(s string) errors.ParserError {
// 				return errors.
// 					MakeParserError(
// 						innerDest.Disp.Line,
// 						innerDest.Disp.Col,
// 						"Bad offset expresion in move into one register dereference"+
// 							" `%s`",
// 						s,
// 					)
// 			},
// 		)
// 	} else {
// 		off = o
// 	}
// 	switch src := mov.Src.Val.(type) {
// 	// mov [rx+1], 1
// 	case ConstExpr:
// 		if imm, ok := IsConstexprType[ConstExprILit](mov.Src); ok {
// 			if mov.DataSize == nil {
// 				return nil,
// 					errors.
// 						MissingSizeParameter(
// 							outer.Line,
// 							outer.Col,
// 						)
// 			}
// 			return InstMovIDO1{
// 				Imm:      imm.Integer,
// 				Disp:     off,
// 				Base:     innerDest.Reg,
// 				DispOp:   innerDest.OffsetOp,
// 				DataSize: mov.DataSize.Size,
// 				NoOff:    off == 0,
// 			}, nil
// 		}
// 	// mov [rx+1], rx
// 	case RegExpr:
// 		return InstMovRDO1{
// 			Src:    src.Reg,
// 			Disp:   off,
// 			Base:   innerDest.Reg,
// 			DispOp: innerDest.OffsetOp,
// 		}, nil
// 	}
// 	return nil, MakeParserErrorWithExpr(
// 		mov.Src,
// 		func(s string) errors.ParserError {
// 			return errors.
// 				MakeParserError(
// 					mov.Src.Line,
// 					mov.Src.Col,
// 					"Bad source expression for one register dereference move "+
// 						"`%s`.",
// 					s,
// 				)
// 		},
// 	)
// }
// func movRDO2(
// 	innerDest TwoRegOffsetExpr,
// 	mov *InstGenericMov,
// 	outer *Instruction,
// ) (any, error) {
// 	var off int64
// 	if o, ok := getAsOffset(innerDest.Disp); !ok {
// 		return nil, MakeParserErrorWithExpr(
// 			innerDest.Disp,
// 			func(s string) errors.ParserError {
// 				return errors.
// 					MakeParserError(
// 						innerDest.Disp.Line,
// 						innerDest.Disp.Col,
// 						"Bad offset expresion in move into two register dereference"+
// 							" `%s`",
// 						s,
// 					)
// 			},
// 		)
// 	} else {
// 		off = o
// 	}
// 	switch src := mov.Src.Val.(type) {
// 	// mov [rx+rx+1], 1
// 	case ConstExpr:
// 		if imm, ok := IsConstexprType[ConstExprILit](mov.Src); ok {
// 			if mov.DataSize == nil {
// 				return nil,
// 					errors.
// 						MissingSizeParameter(
// 							outer.Line,
// 							outer.Col,
// 						)
// 			}
// 			if off == 0 {
// 				innerDest.RegOp = lx.TOKEN_TPLUS
// 			}
// 			return InstMovIDO2{
// 				Imm:      imm.Integer,
// 				DataSize: mov.DataSize.Size,
// 				OReg1:    innerDest.Reg1,
// 				OReg2:    innerDest.Reg2,
// 				RegOp:    innerDest.RegOp,
// 				Offset:   off,
// 				NoOff:    off == 0,
// 			}, nil
// 		}
// 	// mov [rx+rx+1], rx
// 	case RegExpr:
// 		return InstMovRDO2{
// 			Src:    src.Reg,
// 			Disp:   off,
// 			Base:   innerDest.Reg1,
// 			Index:  innerDest.Reg2,
// 			RegOp:  innerDest.RegOp,
// 			DispOp: innerDest.OffsetOp,
// 		}, nil
// 	}
// 	return nil, MakeParserErrorWithExpr(
// 		mov.Src,
// 		func(s string) errors.ParserError {
// 			return errors.
// 				MakeParserError(
// 					mov.Src.Line,
// 					mov.Src.Col,
// 					"Bad source expression for move into two "+
// 						"register dereference with offset "+
// 						"`%s`.",
// 					s,
// 				)
// 		},
// 	)
// }

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
	case MemExpr:
		var scale Scale
		if s, ok := getAsScale(innerDest.ScaleF); !ok {
			return nil, errors.
				MakeParserError(
					innerDest.ScaleF.Line,
					innerDest.ScaleF.Col,
					"Invalid scale factor")
		} else {
			scale = s
		}
		var offset int64
		if o, ok := getAsOffset(innerDest.Disp); !ok {
			return nil, MakeParserErrorWithExpr(
				innerDest.Disp,
				func(s string) errors.ParserError {
					return errors.
						MakeParserError(
							innerDest.Disp.Line,
							innerDest.Disp.Col,
							"Bad offset expression for mov instruction `%s`.",
							s,
						)
				},
			)
		} else {
			offset = o
		}
		switch src := mov.Src.Val.(type) {
		case RegExpr:
			if innerDest.Index.HasVal() {
				return InstMovRDO2{
					Src:    src.Reg,
					Base:   innerDest.Base,
					RegOp:  innerDest.RegOp,
					Index:  innerDest.Index.Get(),
					Disp:   offset,
					DispOp: innerDest.DispOp,
					SF:     scale,
				}, nil
			} else {
				return InstMovRDO1{
					Src:    src.Reg,
					Base:   innerDest.Base,
					DispOp: innerDest.DispOp,
					Disp:   offset,
					SF:     scale,
				}, nil
			}
		case ConstExpr:
			var imm ConstExprILit
			if i, ok := src.Val.(ConstExprILit); !ok {
				return nil, errors.
					MakeParserError(
						mov.Src.Line,
						mov.Src.Col,
						"Bad source immediate value",
					)
			} else {
				imm = i
			}
			if mov.DataSize == nil {
				return nil,
					errors.
						MissingSizeParameter(
							outer.Line,
							outer.Col,
						)
			}
			if innerDest.Index.HasVal() {
				return InstMovIDO2{
					Imm:      imm.Integer,
					Disp:     offset,
					Base:     innerDest.Base,
					Index:    innerDest.Index.Get(),
					RegOp:    innerDest.RegOp,
					DispOp:   innerDest.DispOp,
					DataSize: mov.DataSize.Size,
					NoOff:    offset == 0,
					SF:       scale,
				}, nil
			} else {
				return InstMovIDO1{
					Imm:      imm.Integer,
					Disp:     offset,
					Base:     innerDest.Base,
					RegOp:    innerDest.RegOp,
					DispOp:   innerDest.DispOp,
					DataSize: mov.DataSize.Size,
					NoOff:    offset == 0,
					SF:       scale,
				}, nil
			}
		}
		// case OneRegOffsetExpr:
		// 	if innerDest.Disp == nil {
		// 		return movRDO1_NO(RegExpr{
		// 			Reg: innerDest.Reg,
		// 		}, mov, outer)
		// 	} else {
		// 		return movRDO1(innerDest, mov, outer)
		// 	}
		// case TwoRegOffsetExpr:
		// 	return movRDO2(innerDest, mov, outer)
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
