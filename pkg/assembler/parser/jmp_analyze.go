package assembler

import (
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
	"github.com/JakubCygaro/alphataurus/pkg/assembler/parser/errors"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func GetConcreteJmpInst(genericCmp InstGenericJmp, outer *Instruction) (any, error) {
	switch addr := genericCmp.Expr.Val.(type) {
	case ConstExpr:
		if i, ok := addr.Val.(ConstExprILit); !ok {
			return nil, MakeParserErrorWithExpr(
				genericCmp.Expr,
				func(s string) errors.ParserError {
					return errors.
						MakeParserError(
							genericCmp.Expr.Line,
							genericCmp.Expr.Col,
							"Disallowed expression used as jump address `%s`. "+
								"Value is not an integer.",
							s,
						)

				},
			)
		} else {
			return InstJmpI{
				Address: i.Signed(),
				JmpTy:   genericCmp.Variant,
			}, nil
		}
	case DerefExpr:
		switch deref := addr.Inner.Val.(type) {
		case RegExpr:
			if deref.Reg.Reg != vm.IP_IDX {
				return nil,
					errors.
						MakeParserError(
							addr.Inner.Line,
							addr.Inner.Col,
							"Expected IP register in IP relative jump instruction. "+
								"Got `%s` instead.",
							deref.Reg.String(),
						)
			}
			return InstJmpIP0R{
				Offset: 0,
				OpTy:   vm.OP_TADD,
				JmpTy:  genericCmp.Variant,
			}, nil
		case OneRegOffsetExpr:
			if deref.Reg.Reg != vm.IP_IDX {
				return nil,
					errors.
						MakeParserError(
							addr.Inner.Line,
							addr.Inner.Col,
							"Expected IP register in IP relative jump instruction. "+
								"Got `%s` instead.",
							deref.Reg.String(),
						)
			}
			if off, ok := getAsOffset(deref.Offset); !ok {
				return nil, MakeParserErrorWithExpr(
					deref.Offset,
					func(s string) errors.ParserError {
						return errors.
							MakeParserError(
								addr.Inner.Line,
								addr.Inner.Col,
								"Bad offset value in IP relative jump instruction. "+
									"Got expressiom `%s` which does not evaluate "+
									"to a valid offset.",
								s,
							)

					},
				)
			} else {
				return InstJmpIP0R{
					Offset: off,
					OpTy:   vm.OP_TADD,
					JmpTy:  genericCmp.Variant,
				}, nil
			}
		case TwoRegOffsetExpr:
			if deref.Reg1.Reg != vm.IP_IDX && deref.Reg2.Reg != vm.IP_IDX {
				return nil,
					errors.
						MakeParserError(
							addr.Inner.Line,
							addr.Inner.Col,
							"Expected IP register in IP relative jump instruction. "+
								"Got `%s` and `%s` instead.",
							deref.Reg1.String(),
							deref.Reg2.String(),
						)
			}
			if off, ok := getAsOffset(deref.Offset); !ok {
				return nil, MakeParserErrorWithExpr(
					deref.Offset,
					func(s string) errors.ParserError {
						return errors.
							MakeParserError(
								addr.Inner.Line,
								addr.Inner.Col,
								"Bad offset value in IP relative jump instruction. "+
									"Got expressiom `%s` which does not evaluate "+
									"to a valid offset.",
								s,
							)
					},
				)
			} else {
				return InstJmpIP1R{
					Reg:    deref.Reg2,
					Offset: off,
					OpTy:   deref.RegOp,
					JmpTy:  genericCmp.Variant,
				}, nil
			}
		}
	}
	return nil,
		errors.
			MakeParserError(
				genericCmp.Expr.Line,
				genericCmp.Expr.Col,
				"Invalid jump instruction address expression.",
			)
}
func GetConcreteCallInst(genericCall InstGenericCall, outer *Instruction) (any, error) {
	switch addr := genericCall.Expr.Val.(type) {
	case ConstExpr:
		if i, ok := addr.Val.(ConstExprILit); !ok {
			return nil, MakeParserErrorWithExpr(
				genericCall.Expr,
				func(s string) errors.ParserError {
					return errors.
						MakeParserError(
							genericCall.Expr.Line,
							genericCall.Expr.Col,
							"Disallowed expression used as call address `%s`. "+
								"Value is not an integer.",
							s,
						)
				},
			)
		} else {
			return InstCallI{
				Address: i.Signed(),
			}, nil
		}
	case DerefExpr:
		switch deref := addr.Inner.Val.(type) {
		case RegExpr:
			if deref.Reg.Reg != vm.IP_IDX {
				return nil,
					errors.
						MakeParserError(
							genericCall.Expr.Line,
							genericCall.Expr.Col,
							"Expected IP register in IP relative call instruction. "+
								"Got `%s` instead.",
							deref.Reg.String(),
						)
			}
			return InstCallIP0R{
				Offset: 0,
				OpTy:   vm.OP_TADD,
			}, nil
		case OneRegOffsetExpr:
			if deref.Reg.Reg != vm.IP_IDX {
				return nil,
					errors.
						MakeParserError(
							genericCall.Expr.Line,
							genericCall.Expr.Col,
							"Expected IP register in IP relative call instruction. "+
								"Got `%s` instead.",
							deref.Reg.String(),
						)
			}
			if off, ok := getAsOffset(deref.Offset); !ok {
				return nil, MakeParserErrorWithExpr(
					deref.Offset,
					func(s string) errors.ParserError {
						return errors.
							MakeParserError(
							genericCall.Expr.Line,
							genericCall.Expr.Col,
								"Bad offset value in IP relative call instruction. "+
									"Got expressiom `%s` which does not evaluate "+
									"to a valid offset.",
								s,
							)
					},
				)
			} else {
				return InstCallIP0R{
					Offset: off,
					OpTy:   vm.OP_TADD,
				}, nil
			}
		case TwoRegOffsetExpr:
			if deref.Reg1.Reg != vm.IP_IDX && deref.Reg2.Reg != vm.IP_IDX {
				return nil,
					errors.
						MakeParserError(
							genericCall.Expr.Line,
							genericCall.Expr.Col,
							"Expected IP register in IP relative call instruction. "+
								"Got `%s` and `%s` instead.",
							deref.Reg1.String(),
							deref.Reg2.String(),
						)
			}
			if off, ok := getAsOffset(deref.Offset); !ok {
				return nil, MakeParserErrorWithExpr(
					deref.Offset,
					func(s string) errors.ParserError {
						return errors.
							MakeParserError(
							genericCall.Expr.Line,
							genericCall.Expr.Col,
								"Bad offset value in IP relative call instruction. "+
									"Got expressiom `%s` which does not evaluate "+
									"to a valid offset.",
								s,
							)
					},
				)
			} else {
				var oReg lx.RegisterData
				if deref.Reg2.Reg != vm.IP_IDX {
					oReg = deref.Reg2
				} else {
					oReg = deref.Reg1
				}
				if oReg.Size != vm.SZ_64 {
					return nil, errors.
						MakeParserError(
							genericCall.Expr.Line,
							genericCall.Expr.Col,
							"Bad register size `%s`, expected a WORD sized register.",
							oReg.String(),
						)
				}
				return InstCallIP1R{
					Reg:    oReg,
					Offset: off,
					OpTy:   deref.RegOp,
				}, nil
			}
		}
	}
	return nil,
		errors.
			MakeParserError(
				genericCall.Expr.Line,
				genericCall.Expr.Col,
				"Invalid call instruction address expression.",
			)
}
