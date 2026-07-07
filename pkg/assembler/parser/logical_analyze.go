package assembler

import (
	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
)

func GetConcreteLogicalInst(genericLogical InstGenericLogical) (any, error) {
	switch first := genericLogical.First.Val.(type) {
	case RegExpr:
		switch second := genericLogical.Second.Val.(type) {
		case RegExpr:
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
			return nil, errors.
				BadSource(genericLogical.Second.Line, genericLogical.Second.Col)
		}
	}
	return nil, errors.
		BadDestination(genericLogical.First.Line, genericLogical.First.Col)
}
func GetConcreteCmpInst(genericCmp InstGenericCmp) (any, error) {
	switch min := genericCmp.Min.Val.(type) {
	case RegExpr:
		switch sub := genericCmp.Sub.Val.(type) {
		case RegExpr:
			return InstCmpRR{
				Sub: sub.Reg,
				Min: min.Reg,
				Ty:  genericCmp.Ty,
			}, nil
		case ConstExpr:
			if imm, ok := IsConstexprType[ConstExprILit](genericCmp.Sub); ok {
				return InstCmpIR{
					Imm: imm.Integer,
					Min: min.Reg,
					Ty:  genericCmp.Ty,
				}, nil
			} else if imm, ok :=
				IsConstexprType[ConstExprFLit](genericCmp.Sub); ok {
				return InstCmpIR{
					Imm: imm.Float,
					Min: min.Reg,
					Ty:  genericCmp.Ty,
				}, nil
			}
		default:
			return nil, errors.
				BadSource(genericCmp.Sub.Line, genericCmp.Sub.Col)
		}
	}
	return nil, errors.
		BadDestination(genericCmp.Min.Line, genericCmp.Min.Col)
}
