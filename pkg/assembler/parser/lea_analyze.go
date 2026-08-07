package assembler

import (
	"github.com/JakubCygaro/alphataurus/pkg/assembler/parser/errors"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func GetConcreteLeaInst(
	genericLea InstGenericLea,
	outer *Instruction,
) (any, error) {
	innerMov, err := GetConcreteMovInst(genericLea.Mov, outer)
	if err != nil {
		return nil, err
	}
	switch mov := innerMov.(type) {
	case InstMovDRO1:
		if mov.Dest.Size != vm.SZ_64 {
			return nil, errors.
				MakeParserError(
					outer.Line,
					outer.Col,
					"Destination registed size must be 64-bits wide",
				)
		}
		return InstLeaO1{
			Mov: mov,
		}, nil
	case InstMovDRO2:
		if mov.Dest.Size != vm.SZ_64 {
			return nil, errors.
				MakeParserError(
					outer.Line,
					outer.Col,
					"Destination registed size must be 64-bits wide",
				)
		}
		return InstLeaO2{
			Mov: mov,
		}, nil
	default:
		return nil, errors.MakeParserError(
			outer.Line,
			outer.Col,
			"Load effective address not allowed for this type of data move.",
		)
	}
}
