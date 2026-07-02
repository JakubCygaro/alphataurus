package assembler

import (
	// "github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
	// "github.com/JakubCygaro/alphataurus/pkg/vm"
)

func GetConcreteArth(genericArth pr.InstGenericArth) (any, error) {
	switch dest := genericArth.Dest.Val.(type) {
	case pr.RegExpr:
		switch src := genericArth.Src.Val.(type) {
		case pr.RegExpr:
			return pr.InstArthRR{
				Src:      src.Reg,
				Dest:     dest.Reg,
				Ty:       genericArth.Ty,
				DataSize: genericArth.DataSize,
				ArthTy:   genericArth.ArthTy,
			}, nil
		case pr.ConstExpr:
			switch imm := src.Val.(type) {
			case pr.ConstExprILit:
				return pr.InstArthIR{
					Imm:      imm.Integer,
					Dest:     dest.Reg,
					Ty:       genericArth.Ty,
					DataSize: genericArth.DataSize,
					ArthTy:   genericArth.ArthTy,
				}, nil
			case pr.ConstExprFLit:
				return pr.InstArthIR{
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
