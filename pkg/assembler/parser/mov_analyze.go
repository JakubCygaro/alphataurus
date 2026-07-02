package assembler

import (
	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

// mov rx, rx
func gcmMovRR(dest, src RegExpr, mov *InstMov) (any, error) {
	if mov.DataSize != nil {
		return nil, errors.UnnecessarySizeParameter("", mov.DataSize.Line, mov.Dest.Col)
	}
	return InstMovRR{
		Src:  src.Reg,
		Dest: dest.Reg,
	}, nil
}

// mov rx, int/float
func gcmMovRI(dest RegExpr, imm ConstExpr, mov *InstMov) (any, error) {
	if mov.DataSize != nil {
		return nil, errors.
			UnnecessarySizeParameter("", mov.Src.Line, mov.Src.Col)
	}
	switch cexpr := imm.Val.(type) {
	// mov rx, 1
	case ConstExprILit:
		return InstMovIR{
			Imm:      cexpr.Integer,
			Dest:     dest.Reg,
			DataSize: dest.Reg.Size,
		}, nil
	// mov rx, 1.0
	case ConstExprFLit:
		if dest.Reg.Size != vm.SZ_64 {
			return nil, errors.BadSizeArgument("TODO", "TODO",
				mov.DataSize.Line, mov.Dest.Col)
		}
		return InstMovIR{
			Imm:      cexpr.Float,
			Dest:     dest.Reg,
			DataSize: dest.Reg.Size,
		}, nil
	}
	return nil, nil
}

// mov rx, [(inner)]
func gcmMovDR(dest RegExpr, deref DerefExpr, mov *InstMov) (any, error) {
	switch inner := deref.Inner.Val.(type) {
	// mov rx, [1]
	case ConstExpr:
		switch ilit := inner.Val.(type) {
		case ConstExprILit:
			return InstMovDR{
				Address: ilit.Signed(),
				Dest:    dest.Reg,
			}, nil
		}
	// mov rx, [rx + 1]
	case OneRegOffsetExpr:
		if IsConstexprType[ConstExprILit](inner.Offset) {
			return InstMovDRO1{
				Dest:   dest.Reg,
				Offset: inner.Offset.Val.(ConstExprILit).Signed(),
				OReg1:  inner.Reg,
				OffOp:  inner.OffsetOp,
			}, nil
		}
	// mov rx, [rx + rx + 1]
	case TwoRegOffsetExpr:
		if IsConstexprType[ConstExprILit](inner.Offset) {
			return InstMovDRO2{
				Dest:   dest.Reg,
				OReg1:  inner.Reg1,
				RegOp:  inner.RegOp,
				OReg2:  inner.Reg2,
				OffOp:  inner.OffsetOp,
				Offset: inner.Offset.Val.(ConstExprILit).Signed(),
			}, nil
		}
	}
	return nil, nil
}

// mov rx, (expr)
func gcmIntoRegister(dest RegExpr, mov *InstMov) (any, error) {
	switch src := mov.Src.Val.(type) {
	case RegExpr:
		return gcmMovRR(dest, src, mov)
	case ConstExpr:
		return gcmMovRI(dest, src, mov)
	// mov rx, [(inner)]
	case DerefExpr:
		return gcmMovDR(dest, src, mov)
	}
	return nil, nil
}
func gcmRD(inner ConstExpr, mov *InstMov) (any, error) {
	var off int64
	if cexpr, ok := inner.Val.(ConstExprILit); !ok {
		return nil, nil
	} else {
		off = cexpr.Signed()
	}
	switch src := mov.Src.Val.(type) {
	// mov [1], rx
	case RegExpr:
		if mov.DataSize != nil {
			return nil, errors.UnnecessarySizeParameter("TODO",
				mov.Dest.Line, mov.Dest.Col)
		}
		return InstMovRD{
			Src:     src.Reg,
			Address: off,
		}, nil
		// mov [1], 1
	case ConstExpr:
		if cexpr, ok := src.Val.(ConstExprILit); ok {
			if mov.DataSize == nil {
				return nil, errors.MissingDataSize(mov.Dest.Line, mov.Dest.Col)
			}

			return InstMovID{
				Imm:      cexpr.Integer,
				Address:  off,
				DataSize: mov.DataSize.Size,
			}, nil
		}
	}
	return nil, nil
}
func gcmRDO1_NO(inner RegExpr, mov *InstMov) (any, error) {
	switch src := mov.Src.Val.(type) {
	// mov [rx], 1
	case ConstExpr:
		if IsConstexprType[ConstExprILit](mov.Src) {
			if mov.DataSize == nil {
				return nil, errors.MissingDataSize(mov.DataSize.Line, mov.Dest.Col)
			}
			return InstMovIDO1{
				Imm:      src.Val.(ConstExprILit).Integer,
				Offset:   0,
				OReg1:    inner.Reg,
				OffsetOp: vm.OP_TADD,
				DataSize: mov.DataSize.Size,
				NoOff:    true,
			}, nil
		}
	// mov [rx], rx
	case RegExpr:
		return InstMovRDO1{
			Src:      src.Reg,
			Offset:   0,
			OReg1:    inner.Reg,
			OffsetOp: vm.OP_TADD,
		}, nil
	}
	return nil, nil
}
func getAsOffset(expr *Expr) (int64, bool) {
	if !IsConstexprType[ConstExprILit](expr) {
		return 0, false
	} else {
		return expr.Val.(ConstExpr).Val.(ConstExprILit).Signed(), true
	}
}
func gcmRDO1(inner OneRegOffsetExpr, mov *InstMov) (any, error) {
	var off int64
	if o, ok := getAsOffset(inner.Offset); !ok {
		return nil, nil
	} else {
		off = o
	}
	switch src := mov.Src.Val.(type) {
	// mov [rx+1], 1
	case ConstExpr:
		if IsConstexprType[ConstExprILit](mov.Src) {
			if mov.DataSize == nil {
				return nil, errors.MissingDataSize(mov.Dest.Line, mov.Dest.Col)
			}
			return InstMovIDO1{
				Imm:      src.Val.(ConstExprILit).Integer,
				Offset:   off,
				OReg1:    inner.Reg,
				OffsetOp: inner.OffsetOp,
				DataSize: mov.DataSize.Size,
				NoOff:    off == 0,
			}, nil
		}
	// mov [rx+1], rx
	case RegExpr:
		return InstMovRDO1{
			Src:      src.Reg,
			Offset:   off,
			OReg1:    inner.Reg,
			OffsetOp: inner.OffsetOp,
		}, nil
	}
	return nil, nil
}
func gcmRDO2(inner TwoRegOffsetExpr, mov *InstMov) (any, error) {
	var off int64
	if inner.Offset == nil {
		off = 0
	} else if o, ok := getAsOffset(inner.Offset); !ok {
		return nil, nil
	} else {
		off = o
	}
	switch src := mov.Src.Val.(type) {
	// mov [rx+rx+1], 1
	case ConstExpr:
		if IsConstexprType[ConstExprILit](mov.Src) {
			if mov.DataSize == nil {
				return nil, errors.MissingDataSize(mov.Src.Line, mov.Src.Col)
			}
			return InstMovIDO2{
				Imm:      src.Val.(ConstExprILit).Integer,
				DataSize: mov.DataSize.Size,
				OReg1:    inner.Reg1,
				OReg2:    inner.Reg2,
				RegOp:    inner.RegOp,
				Offset:   off,
				OffsetOp: inner.OffsetOp,
				NoOff:    off == 0,
			}, nil
		}
	// mov [rx+rx+1], rx
	case RegExpr:
		return InstMovRDO2{
			Src:      src.Reg,
			Offset:   off,
			OReg1:    inner.Reg1,
			OReg2:    inner.Reg2,
			RegOp:    inner.RegOp,
			OffsetOp: inner.OffsetOp,
		}, nil
	}

	return nil, nil
}

func gcmIntoDeref(dest DerefExpr, mov *InstMov) (any, error) {
	switch inner := mov.Dest.Val.(type) {
	case ConstExpr:
		return gcmRD(inner, mov)
	// mov [rx], (expr)
	case RegExpr:
		return gcmRDO1_NO(inner, mov)
	// mov [rx+1], (expr)
	case OneRegOffsetExpr:
		if inner.Offset == nil {
			return gcmRDO1_NO(RegExpr{
				Reg: inner.Reg,
			}, mov)
		} else {
			return gcmRDO1(inner, mov)
		}
	// mov [rx+rx+1], (expr)
	case TwoRegOffsetExpr:
		return gcmRDO2(inner, mov)
	}
	return nil, nil
}

func GetConcreteMovInst(genericMov InstMov) (any, error) {
	switch dest := genericMov.Dest.Val.(type) {
	// mov rx, (src)
	case RegExpr:
		return gcmIntoRegister(dest, &genericMov)
	// mov [(inner)], (src)
	case DerefExpr:
		return gcmIntoDeref(dest, &genericMov)
	}
	return nil, nil
}
