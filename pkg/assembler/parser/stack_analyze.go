package assembler

import (
	"fmt"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func GetConcretePushInst(push InstGenericPush) (any, error) {
	switch val := push.Expr.Val.(type) {
	case RegExpr:
		if push.DataSz != nil {
			return nil, errors.
				UnnecessarySizeParameter("TODO", push.DataSz.Line, push.DataSz.Col)
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
				return errors.BadSizeArgument(
					"TODO",
					"TODO WORD",
					push.DataSz.Line,
					push.DataSz.Col,
				), nil
			}
			return InstPushI{
				Imm:    cexpr.Float,
				DataSz: dataSz,
			}, nil
		}
	}
	return nil, fmt.
		Errorf("TODO: bad push parameter expression type")
}
