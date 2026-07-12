package assembler

import (
	"github.com/JakubCygaro/alphataurus/pkg/assembler/parser/errors"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func GetConcreteLogicalInst(
	genericLogical InstGenericLogical,
	outer *Instruction,
) (any, error) {
	switch first := genericLogical.First.Val.(type) {
	case RegExpr:
		if !vm.IsLogRAllowed(byte(first.Reg.Reg)) {
			return nil,
				errors.
					DisallowedSrcReg(
						genericLogical.Second.Line,
						genericLogical.Second.Col,
						first.Reg,
					)
		}
		switch second := genericLogical.Second.Val.(type) {
		case RegExpr:
			if !vm.IsLogRAllowed(byte(second.Reg.Reg)) {
				return nil,
					errors.
						DisallowedSrcReg(
							genericLogical.Second.Line,
							genericLogical.Second.Col,
							second.Reg,
						)
			}
			return InstLogicalRR{
				Second: second.Reg,
				First:  first.Reg,
				LogTy:  genericLogical.LogTy,
			}, nil
		case ConstExpr:
			if imm, ok :=
				IsConstexprType[ConstExprILit](genericLogical.Second); ok {
				return InstLogicalIR{
					Imm:   imm.Integer,
					First: first.Reg,
					LogTy: genericLogical.LogTy,
				}, nil
			} else if imm, ok :=
				IsConstexprType[ConstExprFLit](genericLogical.Second); ok {
				return InstLogicalIR{
					Imm:   imm.Float,
					First: first.Reg,
					LogTy: genericLogical.LogTy,
				}, nil
			}
		default:
			return nil, MakeParserErrorWithExpr(
				genericLogical.Second,
				func(s string) errors.ParserError {
					return errors.
						DisallowedSrc(
							genericLogical.Second.Line,
							genericLogical.Second.Col,
							s,
						)
				},
			)
		}
	}
	return nil, MakeParserErrorWithExpr(
		genericLogical.First,
		func(s string) errors.ParserError {
			return errors.
				DisallowedDest(
					genericLogical.First.Line,
					genericLogical.First.Col,
					s,
				)
		},
	)
}
func GetConcreteCmpInst(genericCmp InstGenericCmp) (any, error) {
	switch minu := genericCmp.Min.Val.(type) {
	case RegExpr:
		switch sub := genericCmp.Sub.Val.(type) {
		case RegExpr:
			return InstCmpRR{
				Sub: sub.Reg,
				Min: minu.Reg,
				Ty:  genericCmp.Ty,
			}, nil
		case ConstExpr:
			if imm, ok := IsConstexprType[ConstExprILit](genericCmp.Sub); ok {
				return InstCmpIR{
					Imm: imm.Integer,
					Min: minu.Reg,
					Ty:  genericCmp.Ty,
				}, nil
			} else if imm, ok :=
				IsConstexprType[ConstExprFLit](genericCmp.Sub); ok {
				return InstCmpIR{
					Imm: imm.Float,
					Min: minu.Reg,
					Ty:  genericCmp.Ty,
				}, nil
			}
		default:
			return nil, MakeParserErrorWithExpr(
				genericCmp.Sub,
				func(s string) errors.ParserError {
					return errors.
						DisallowedSrc(
							genericCmp.Sub.Line,
							genericCmp.Sub.Col,
							s,
						)
				},
			)
		}
	}
	return nil, MakeParserErrorWithExpr(
		genericCmp.Min,
		func(s string) errors.ParserError {
			return errors.
				DisallowedDest(
					genericCmp.Min.Line,
					genericCmp.Min.Col,
					s,
				)
		},
	)
}
