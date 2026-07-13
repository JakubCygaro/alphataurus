package assembler

import (
	"github.com/JakubCygaro/alphataurus/pkg/assembler/parser/errors"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func GetConcretePushInst(push InstGenericPush) (any, error) {
	switch val := push.Expr.Val.(type) {
	case RegExpr:
		if push.DataSz != nil {
			return nil,
				errors.
					UnnecessarySizeParameter(
						push.DataSz.Line,
						push.DataSz.Col,
						push.DataSz.Size,
					)
		}
		return InstPushR{
			Reg: val.Reg,
		}, nil
	case ConstExpr:
		var dataSz byte
		if push.DataSz == nil {
			dataSz = vm.SZ_64
		} else {
			dataSz = push.DataSz.Size
		}
		switch cexpr := val.Val.(type) {
		case ConstExprILit:
			return InstPushI{
				Imm:    cexpr.Integer,
				DataSz: dataSz,
			}, nil
		case ConstExprFLit:
			if dataSz != vm.SZ_64 {
				return nil, errors.
					MakeParserError(
						push.DataSz.Line,
						push.DataSz.Col,
						"Bad data size parameter with 64-bit"+
							" floating point immediate value",
					)
			}
			return InstPushI{
				Imm:    cexpr.Float,
				DataSz: dataSz,
			}, nil
		}
	}
	return nil, MakeParserErrorWithExpr(
		push.Expr,
		func(s string) errors.ParserError {
			return errors.
				MakeParserError(
					push.Expr.Line,
					push.Expr.Col,
					"Bad expression as parameter to push instruction "+
						"`%s`.",
					s,
				)
		},
	)
}
