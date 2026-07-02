package assembler

import (
	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

// mov rx, rx
func gcmMovRR(dest, src pr.RegExpr, mov *pr.InstGenericMov) (any, error) {
	if mov.DataSize != nil {
		return nil, errors.UnnecessarySizeParameter("", mov.DataSize.Line, mov.Dest.Col)
	}
	return pr.InstMovRR{
		Src:  src.Reg,
		Dest: dest.Reg,
	}, nil
}

// mov rx, int/float
func gcmMovRI(dest pr.RegExpr, imm pr.ConstExpr, mov *pr.InstGenericMov) (any, error) {
	if mov.DataSize != nil {
		return nil, errors.
			UnnecessarySizeParameter("", mov.Src.Line, mov.Src.Col)
	}
	switch cexpr := imm.Val.(type) {
	// mov rx, 1
	case pr.ConstExprILit:
		return pr.InstMovIR{
			Imm:      cexpr.Integer,
			Dest:     dest.Reg,
			DataSize: dest.Reg.Size,
		}, nil
	// mov rx, 1.0
	case pr.ConstExprFLit:
		if dest.Reg.Size != vm.SZ_64 {
			return nil, errors.BadSizeArgument("TODO", "TODO",
				mov.DataSize.Line, mov.Dest.Col)
		}
		return pr.InstMovIR{
			Imm:      cexpr.Float,
			Dest:     dest.Reg,
			DataSize: dest.Reg.Size,
		}, nil
	}
	return nil, nil
}

// mov rx, [(inner)]
func gcmMovDR(dest pr.RegExpr, deref pr.DerefExpr, mov *pr.InstGenericMov) (any, error) {
	switch inner := deref.Inner.Val.(type) {
	// mov rx, [1]
	case pr.ConstExpr:
		switch ilit := inner.Val.(type) {
		case pr.ConstExprILit:
			return pr.InstMovDR{
				Address: ilit.Signed(),
				Dest:    dest.Reg,
			}, nil
		}
	// mov rx, [rx + 1]
	case pr.OneRegOffsetExpr:
		if pr.IsConstexprType[pr.ConstExprILit](inner.Offset) {
			return pr.InstMovDRO1{
				Dest:   dest.Reg,
				Offset: inner.Offset.Val.(pr.ConstExprILit).Signed(),
				OReg1:  inner.Reg,
				OffOp:  inner.OffsetOp,
			}, nil
		}
	// mov rx, [rx + rx + 1]
	case pr.TwoRegOffsetExpr:
		if pr.IsConstexprType[pr.ConstExprILit](inner.Offset) {
			return pr.InstMovDRO2{
				Dest:   dest.Reg,
				OReg1:  inner.Reg1,
				RegOp:  inner.RegOp,
				OReg2:  inner.Reg2,
				OffOp:  inner.OffsetOp,
				Offset: inner.Offset.Val.(pr.ConstExprILit).Signed(),
			}, nil
		}
	}
	return nil, nil
}

// mov rx, (expr)
func gcmIntoRegister(dest pr.RegExpr, mov *pr.InstGenericMov) (any, error) {
	switch src := mov.Src.Val.(type) {
	case pr.RegExpr:
		return gcmMovRR(dest, src, mov)
	case pr.ConstExpr:
		return gcmMovRI(dest, src, mov)
	// mov rx, [(inner)]
	case pr.DerefExpr:
		return gcmMovDR(dest, src, mov)
	}
	return nil, nil
}
func gcmRD(inner pr.ConstExpr, mov *pr.InstGenericMov) (any, error) {
	var off int64
	if cexpr, ok := inner.Val.(pr.ConstExprILit); !ok {
		return nil, nil
	} else {
		off = cexpr.Signed()
	}
	switch src := mov.Src.Val.(type) {
	// mov [1], rx
	case pr.RegExpr:
		if mov.DataSize != nil {
			return nil, errors.UnnecessarySizeParameter("TODO",
				mov.Dest.Line, mov.Dest.Col)
		}
		return pr.InstMovRD{
			Src:     src.Reg,
			Address: off,
		}, nil
		// mov [1], 1
	case pr.ConstExpr:
		if cexpr, ok := src.Val.(pr.ConstExprILit); ok {
			if mov.DataSize == nil {
				return nil, errors.MissingDataSize(mov.Dest.Line, mov.Dest.Col)
			}

			return pr.InstMovID{
				Imm:      cexpr.Integer,
				Address:  off,
				DataSize: mov.DataSize.Size,
			}, nil
		}
	}
	return nil, nil
}
func gcmRDO1_NO(inner pr.RegExpr, mov *pr.InstGenericMov) (any, error) {
	switch src := mov.Src.Val.(type) {
	// mov [rx], 1
	case pr.ConstExpr:
		if pr.IsConstexprType[pr.ConstExprILit](mov.Src) {
			if mov.DataSize == nil {
				return nil, errors.MissingDataSize(mov.DataSize.Line, mov.Dest.Col)
			}
			return pr.InstMovIDO1{
				Imm:      src.Val.(pr.ConstExprILit).Integer,
				Offset:   0,
				OReg1:    inner.Reg,
				OffsetOp: vm.OP_TADD,
				DataSize: mov.DataSize.Size,
				NoOff:    true,
			}, nil
		}
	// mov [rx], rx
	case pr.RegExpr:
		return pr.InstMovRDO1{
			Src:      src.Reg,
			Offset:   0,
			OReg1:    inner.Reg,
			OffsetOp: vm.OP_TADD,
		}, nil
	}
	return nil, nil
}
func getAsOffset(expr *pr.Expr) (int64, bool) {
	if !pr.IsConstexprType[pr.ConstExprILit](expr) {
		return 0, false
	} else {
		return expr.Val.(pr.ConstExpr).Val.(pr.ConstExprILit).Signed(), true
	}
}
func gcmRDO1(inner pr.OneRegOffsetExpr, mov *pr.InstGenericMov) (any, error) {
	var off int64
	if o, ok := getAsOffset(inner.Offset); !ok {
		return nil, nil
	} else {
		off = o
	}
	switch src := mov.Src.Val.(type) {
	// mov [rx+1], 1
	case pr.ConstExpr:
		if pr.IsConstexprType[pr.ConstExprILit](mov.Src) {
			if mov.DataSize == nil {
				return nil, errors.MissingDataSize(mov.Dest.Line, mov.Dest.Col)
			}
			return pr.InstMovIDO1{
				Imm:      src.Val.(pr.ConstExprILit).Integer,
				Offset:   off,
				OReg1:    inner.Reg,
				OffsetOp: inner.OffsetOp,
				DataSize: mov.DataSize.Size,
				NoOff:    off == 0,
			}, nil
		}
	// mov [rx+1], rx
	case pr.RegExpr:
		return pr.InstMovRDO1{
			Src:      src.Reg,
			Offset:   off,
			OReg1:    inner.Reg,
			OffsetOp: inner.OffsetOp,
		}, nil
	}
	return nil, nil
}
func gcmRDO2(inner pr.TwoRegOffsetExpr, mov *pr.InstGenericMov) (any, error) {
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
	case pr.ConstExpr:
		if pr.IsConstexprType[pr.ConstExprILit](mov.Src) {
			if mov.DataSize == nil {
				return nil, errors.MissingDataSize(mov.Src.Line, mov.Src.Col)
			}
			return pr.InstMovIDO2{
				Imm:      src.Val.(pr.ConstExprILit).Integer,
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
	case pr.RegExpr:
		return pr.InstMovRDO2{
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

func gcmIntoDeref(dest pr.DerefExpr, mov *pr.InstGenericMov) (any, error) {
	switch inner := mov.Dest.Val.(type) {
	case pr.ConstExpr:
		return gcmRD(inner, mov)
	// mov [rx], (expr)
	case pr.RegExpr:
		return gcmRDO1_NO(inner, mov)
	// mov [rx+1], (expr)
	case pr.OneRegOffsetExpr:
		if inner.Offset == nil {
			return gcmRDO1_NO(pr.RegExpr{
				Reg: inner.Reg,
			}, mov)
		} else {
			return gcmRDO1(inner, mov)
		}
	// mov [rx+rx+1], (expr)
	case pr.TwoRegOffsetExpr:
		return gcmRDO2(inner, mov)
	}
	return nil, nil
}

func GetConcreteMovInst(genericMov pr.InstGenericMov) (any, error) {
	switch dest := genericMov.Dest.Val.(type) {
	// mov rx, (src)
	case pr.RegExpr:
		return gcmIntoRegister(dest, &genericMov)
	// mov [(inner)], (src)
	case pr.DerefExpr:
		return gcmIntoDeref(dest, &genericMov)
	}
	return nil, nil
}
