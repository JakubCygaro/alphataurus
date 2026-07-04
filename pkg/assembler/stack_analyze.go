package assembler

import (
	"fmt"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func GetConcretePushInst(push pr.InstGenericPush) (any, error) {
	switch val := push.Expr.Val.(type) {
	case pr.RegExpr:
		if push.DataSz != nil {
			return nil, errors.
				UnnecessarySizeParameter("TODO", push.DataSz.Line, push.DataSz.Col)
		}
		return pr.InstPushR{
			Reg: val.Reg,
		}, nil
	case pr.ConstExpr:
		var dataSz byte
		if push.DataSz == nil {
			dataSz = vm.SZ_64
		} else {
			dataSz = push.DataSz.Size
		}
		switch cexpr := val.Val.(type) {
		case pr.ConstExprILit:
			return pr.InstPushI{
				Imm:    cexpr.Integer,
				DataSz: dataSz,
			}, nil
		case pr.ConstExprFLit:
			if dataSz != vm.SZ_64 {
				return errors.BadSizeArgument(
					"TODO",
					"TODO WORD",
					push.DataSz.Line,
					push.DataSz.Col,
				), nil
			}
			return pr.InstPushI{
				Imm:    cexpr.Float,
				DataSz: dataSz,
			}, nil
		}
	}
	return nil, fmt.
		Errorf("TODO: bad push parameter expression type")
}
