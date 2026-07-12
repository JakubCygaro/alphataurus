package assembler

import (
	"github.com/JakubCygaro/alphataurus/pkg/assembler/parser/errors"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func GetConcreteArthInst(genericArth InstGenericArth, outer *Instruction) (any, error) {
	switch dest := genericArth.Dest.Val.(type) {
	case RegExpr:
		if !vm.IsArthRAllowed(byte(dest.Reg.Reg)) {
			return nil,
				errors.
					DisallowedDestReg(
						genericArth.Dest.Line,
						genericArth.Dest.Col,
						dest.Reg,
					)
		}
		switch src := genericArth.Src.Val.(type) {
		case RegExpr:
			if !vm.IsArthRAllowed(byte(src.Reg.Reg)) {
				return nil,
					errors.
						DisallowedSrcReg(
							genericArth.Src.Line,
							genericArth.Src.Col,
							src.Reg,
						)
			}
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
		return nil, MakeParserErrorWithExpr(
			genericArth.Src,
			func(s string) errors.ParserError {
				return errors.DisallowedSrc(
					genericArth.Src.Line,
					genericArth.Src.Col,
					s,
				)
			},
		)
	}
	return nil, MakeParserErrorWithExpr(
		genericArth.Dest,
		func(s string) errors.ParserError {
			return errors.DisallowedDest(
				genericArth.Dest.Line,
				genericArth.Dest.Col,
				s,
			)
		},
	)
}
