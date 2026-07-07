package assembler

import (
	// "github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	// "github.com/JakubCygaro/alphataurus/pkg/vm"
)

func GetConcreteArthInst(genericArth InstGenericArth) (any, error) {
	switch dest := genericArth.Dest.Val.(type) {
	case RegExpr:
		switch src := genericArth.Src.Val.(type) {
		case RegExpr:
			return InstArthRR{
				Src:      src.Reg,
				Dest:     dest.Reg,
				Ty:       genericArth.Ty,
				DataSize: genericArth.DataSize,
				ArthTy:   genericArth.ArthTy,
			}, nil
		case ConstExpr:
			switch imm := src.Val.(type) {
			case ConstExprILit:
				return InstArthIR{
					Imm:      imm.Integer,
					Dest:     dest.Reg,
					Ty:       genericArth.Ty,
					DataSize: genericArth.DataSize,
					ArthTy:   genericArth.ArthTy,
				}, nil
			case ConstExprFLit:
				return InstArthIR{
					Imm:      imm.Float,
					Dest:     dest.Reg,
					Ty:       genericArth.Ty,
					DataSize: genericArth.DataSize,
					ArthTy:   genericArth.ArthTy,
				}, nil
			}
		}
		return nil,
			errors.BadSource(genericArth.Dest.Line, genericArth.Dest.Col)
	}
	return nil,
		errors.BadDestination(genericArth.Dest.Line, genericArth.Dest.Col)
}
