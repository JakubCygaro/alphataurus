package assembler

import "github.com/JakubCygaro/alphataurus/pkg/assembler/errors"

func GetConcreteMovInst(genericMov InstMov) (any, error) {
	sized := genericMov.DataSize
	switch dest := genericMov.Src.Val.(type) {
	// mov rx, (expr)
	case RegExpr:
		switch src := genericMov.Dest.Val.(type) {
		// mov rx, rx
		case RegExpr:
			if sized == 0xff {
				return errors.MissingDataSize(sizedL, sizedC)
			}
			p.currentInst = Instruction{
				Data: InstMovRR{
					Src:      src.Reg,
					Dest:     dest.Reg,
					DataSize: sized,
				},
			}
		case ConstExpr:
			switch cexpr := src.Val.(type) {
			// mov rx, 1
			case ConstExprILit:
				if sized == 0xff {
					return errors.MissingDataSize(sizedL, sizedC)
				}
				p.currentInst = Instruction{
					Data: InstMovIR{
						Imm:      cexpr.Integer,
						Dest:     dest.Reg,
						DataSize: sized,
					},
				}
			// mov rx, 1.0
			case ConstExprFLit:
				if sized == 0xff {
					return errors.MissingDataSize(sizedL, sizedC)
				}
				p.currentInst = Instruction{
					Data: InstMovIR{
						Imm:      cexpr.Float,
						Dest:     dest.Reg,
						DataSize: sized,
					},
				}
			}
		// mov rx, [(inner)]
		case DerefExpr:
			if sized != 0xff {
				s, _ := lx.GetSizeKeyword(sized)
				return errors.UnnecessarySizeParameter(s, sizedL, sizedC)
			}
			switch inner := src.Inner.Val.(type) {
			// mov rx, [1]
			case ConstExpr:
				switch ilit := inner.Val.(type) {
				case ConstExprILit:
					p.currentInst = Instruction{
						Data: InstMovDR{
							Address: ilit.Signed(),
							Dest:    dest.Reg,
						},
					}
				}
			// mov rx, [rx + 1]
			case OneRegOffsetExpr:
				if IsConstexprType[ConstExprILit](inner.Offset) {
					p.currentInst = Instruction{
						Data: InstMovDRO1{
							Dest:   dest.Reg,
							Offset: inner.Offset.Val.(ConstExprILit).Signed(),
							OReg1:  inner.Reg,
							OpTy:   inner.OffsetOp,
						},
					}
				}
			// mov rx, [rx + rx + 1]
			case TwoRegOffsetExpr:
				if IsConstexprType[ConstExprILit](inner.Offset) {
					p.currentInst = Instruction{
						Data: InstMovDRO2{
							Dest:   dest.Reg,
							Offset: inner.Offset.Val.(ConstExprILit).Signed(),
							OReg1:  inner.Reg1,
							OReg2:  inner.Reg2,
							OpTy:   inner.OffsetOp,
						},
					}
				}
			}
		}
	// mov [(inner)], (dest)
	case DerefExpr:
		switch inner := genericMov.Dest.Val.(type) {
		case ConstExpr:
			var off int64
			if cexpr, ok := inner.Val.(ConstExprILit); !ok {
				return nil
			} else {
				off = cexpr.Signed()
			}
			switch src := genericMov.Src.Val.(type) {
			// mov [1], rx
			case RegExpr:
				if sized != 0xff {
					s, _ := lx.GetSizeKeyword(sized)
					return errors.UnnecessarySizeParameter(s, sizedL, sizedC)
				}
				p.currentInst = Instruction{
					Data: InstMovRD{
						Src:    src.Reg,
						Offset: off,
					},
				}
				// mov [1], 1
			case ConstExprILit:
				if sized == 0xff {
					return errors.MissingDataSize(sizedL, sizedC)
				}
				p.currentInst = Instruction{
					Data: InstMovID{
						Imm:      src.Integer,
						Offset:   off,
						DataSize: sized,
					},
				}
			}
		// mov [rx], (expr)
		case RegExpr:
			switch src := genericMov.Src.Val.(type) {
			// mov [rx], 1
			case ConstExpr:
				if IsConstexprType[ConstExprILit](genericMov.Src) {
					if sized == 0xff {
						return errors.MissingDataSize(sizedL, sizedC)
					}
					p.currentInst = Instruction{
						Data: InstMovIDO1{
							Imm:      src.Val.(ConstExprILit).Integer,
							Offset:   0,
							OReg1:    inner.Reg,
							OpTy:     vm.OP_TADD,
							DataSize: sized,
							NoOff:    true,
						},
					}
				}
			// mov [rx], rx
			case RegExpr:
				p.currentInst = Instruction{
					Data: InstMovRDO1{
						Src:    src.Reg,
						Offset: 0,
						OReg1:  inner.Reg,
						OpTy:   vm.OP_TADD,
					},
				}
			}
		// mov [rx+1], (expr)
		case OneRegOffsetExpr:
			var off int64
			if inner.Offset == nil {
				off = 0
			} else if !IsConstexprType[ConstExprILit](inner.Offset) {
				return nil
			} else {
				off = inner.Offset.Val.(ConstExpr).Val.(ConstExprILit).Signed()
			}
			switch src := genericMov.Src.Val.(type) {
			// mov [rx+1], 1
			case ConstExpr:
				if IsConstexprType[ConstExprILit](genericMov.Src) {
					if sized == 0xff {
						return errors.MissingDataSize(sizedL, sizedC)
					}
					p.currentInst = Instruction{
						Data: InstMovIDO1{
							Imm:      src.Val.(ConstExprILit).Integer,
							Offset:   off,
							OReg1:    inner.Reg,
							OpTy:     inner.OffsetOp,
							DataSize: sized,
							NoOff:    off == 0,
						},
					}
				}
			// mov [rx+1], rx
			case RegExpr:
				p.currentInst = Instruction{
					Data: InstMovRDO1{
						Src:    src.Reg,
						Offset: off,
						OReg1:  inner.Reg,
						OpTy:   inner.OffsetOp,
					},
				}
			}
		// mov [rx+rx+1], (expr)
		case TwoRegOffsetExpr:
			var off int64
			if inner.Offset == nil {
				off = 0
			} else if !IsConstexprType[ConstExprILit](inner.Offset) {
				return nil
			} else {
				off = inner.Offset.Val.(ConstExpr).Val.(ConstExprILit).Signed()
			}
			switch src := genericMov.Src.Val.(type) {
			// mov [rx+1], 1
			case ConstExpr:
				if IsConstexprType[ConstExprILit](genericMov.Src) {
					if sized == 0xff {
						return errors.MissingDataSize(sizedL, sizedC)
					}
					p.currentInst = Instruction{
						Data: InstMovIDO2{
							Imm:    src.Val.(ConstExprILit).Integer,
							Offset: off,
							OReg1:  inner.Reg1,
							OReg2:  inner.Reg2,
							OpTy:   inner.RegOp,
							DataSize: sized,
							NoOff: off == 0,
						},
					}
				}
			// mov [rx+1], rx
			case RegExpr:
				p.currentInst = Instruction{
					Data: InstMovRDO2{
						Src:    src.Reg,
						Offset: off,
						OReg1:  inner.Reg1,
						OReg2:  inner.Reg2,
						OpTy:   inner.RegOp,
					},
				}
			}
		}
	}
}
