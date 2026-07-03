package assembler

import (
	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
)

func GetConcreteLogicalInst(genericLogical pr.InstGenericLogical) (any, error) {
	switch first := genericLogical.First.Val.(type) {
	case pr.RegExpr:
		switch second := genericLogical.Second.Val.(type) {
		case pr.RegExpr:
			return pr.InstLogicalRR{
				Second: second.Reg,
				First:  first.Reg,
				LogTy:  genericLogical.LogTy,
			}, nil
		case pr.ConstExpr:
			if pr.IsConstexprType[pr.ConstExprILit](genericLogical.Second) {
				return pr.InstLogicalIR{
					Imm:   second.Val.(pr.ConstExprILit).Integer,
					First: first.Reg,
					LogTy: genericLogical.LogTy,
				}, nil
			} else if pr.IsConstexprType[pr.ConstExprFLit](genericLogical.Second) {
				return pr.InstLogicalIR{
					Imm:   second.Val.(pr.ConstExprFLit).Float,
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
func GetConcreteCmpInst(genericCmp pr.InstGenericCmp) (any, error) {
	switch min := genericCmp.Min.Val.(type) {
	case pr.RegExpr:
		switch sub := genericCmp.Sub.Val.(type) {
		case pr.RegExpr:
			return pr.InstCmpRR{
				Sub: sub.Reg,
				Min:  min.Reg,
				Ty:  genericCmp.Ty,
			}, nil
		case pr.ConstExpr:
			if pr.IsConstexprType[pr.ConstExprILit](genericCmp.Sub) {
				return pr.InstCmpIR{
					Imm:   sub.Val.(pr.ConstExprILit).Integer,
					Min: min.Reg,
					Ty: genericCmp.Ty,
				}, nil
			} else if pr.IsConstexprType[pr.ConstExprFLit](genericCmp.Sub) {
				return pr.InstCmpIR{
					Imm:   sub.Val.(pr.ConstExprFLit).Float,
					Min: min.Reg,
					Ty: genericCmp.Ty,
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
