package assembler

import (
	"fmt"

	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

type fieldSelector[Inst any] func(*Inst) **pr.Expr

// evaluate a selected field of an instruction and set the field to the evaluated result
func evalInstField[Inst any](
	a *Assembler,
	inst *Inst,
	slc fieldSelector[Inst],
) (bool, error) {
	expr := slc(inst)
	eval, err := a.ev.TryEvaluateExpression(*expr)
	if eval != nil && err == nil {
		*expr = eval
	}
	return eval != nil, err
}

func evalAll[Inst any](
	a *Assembler, inst *Inst, slcs ...fieldSelector[Inst],
) (bool, error) {
	for _, s := range slcs {
		ok, err := evalInstField(a, inst, s)
		if !ok {
			return false, err
		} else if err != nil {
			return false, err
		}
	}
	return true, nil
}

func (a *Assembler) emitGenericArth(data pr.InstGenericArth, at int) error {
	if ok, err := evalAll(a, &data,
		func(i *pr.InstGenericArth) **pr.Expr {
			return &(i.Dest)
		},
		func(i *pr.InstGenericArth) **pr.Expr {
			return &(i.Src)
		},
	); err != nil {
		return err
	} else if !ok {
		a.unevalInsts[at] = pr.Instruction{
			Line: a.line,
			Col:  a.col,
			Data: data,
		}
		a.emitNop(at)
	} else if arth, err := pr.GetConcreteArthInst(data); err != nil {
		return err
	} else if arth == nil {
		return fmt.Errorf("TODO: bad arth instruction cannot be deduced to concrete arth")
	} else {
		return a.emitInst(pr.Instruction{
			Line: a.line,
			Col:  a.col,
			Data: arth,
		}, at)
	}
	return nil
}

func (a *Assembler) emitGenericMov(data pr.InstGenericMov, at int) error {
	if ok, err := evalAll(a, &data,
		func(i *pr.InstGenericMov) **pr.Expr {
			return &(i.Src)
		},
		func(i *pr.InstGenericMov) **pr.Expr {
			return &(i.Dest)
		},
	); err != nil {
		return err
	} else if !ok {
		a.unevalInsts[at] = pr.Instruction{
			Line: a.line,
			Col:  a.col,
			Data: data,
		}
		a.emitNop(at)
	} else if mov, err := pr.GetConcreteMovInst(data); err != nil {
		return err
	} else if mov == nil {
		return fmt.Errorf("TODO: bad mov instruction cannot be deduced to concrete mov")
	} else {
		return a.emitInst(pr.Instruction{
			Line: a.line,
			Col:  a.col,
			Data: mov,
		}, at)
	}

	return nil
}
func (a *Assembler) emitGenericLogical(data pr.InstGenericLogical, at int) error {
	if ok, err := evalAll(a, &data,
		func(i *pr.InstGenericLogical) **pr.Expr {
			return &(i.First)
		},
		func(i *pr.InstGenericLogical) **pr.Expr {
			return &(i.Second)
		},
	); err != nil {
		return err
	} else if !ok {
		a.unevalInsts[at] = pr.Instruction{
			Line: a.line,
			Col:  a.col,
			Data: data,
		}
		a.emitNop(at)
	} else if mov, err := pr.GetConcreteLogicalInst(data); err != nil {
		return err
	} else if mov == nil {
		return fmt.
			Errorf("TODO: bad logical instruction cannot be deduced to concrete logical")
	} else {
		return a.emitInst(pr.Instruction{
			Line: a.line,
			Col:  a.col,
			Data: mov,
		}, at)
	}

	return nil
}
func (a *Assembler) emitGenericCmp(data pr.InstGenericCmp, at int) error {
	if ok, err := evalAll(a, &data,
		func(i *pr.InstGenericCmp) **pr.Expr {
			return &(i.Min)
		},
		func(i *pr.InstGenericCmp) **pr.Expr {
			return &(i.Sub)
		},
	); err != nil {
		return err
	} else if !ok {
		a.unevalInsts[at] = pr.Instruction{
			Line: a.line,
			Col:  a.col,
			Data: data,
		}
		a.emitNop(at)
	} else if mov, err := pr.GetConcreteCmpInst(data); err != nil {
		return err
	} else if mov == nil {
		return fmt.Errorf("TODO: bad mov instruction cannot be deduced to concrete mov")
	} else {
		return a.emitInst(pr.Instruction{
			Line: a.line,
			Col:  a.col,
			Data: mov,
		}, at)
	}

	return nil
}
func (a *Assembler) emitGenericPush(data pr.InstGenericPush, at int) error {
	if ok, err := evalAll(a, &data,
		func(i *pr.InstGenericPush) **pr.Expr {
			return &(i.Expr)
		},
	); err != nil {
		return err
	} else if !ok {
		a.unevalInsts[at] = pr.Instruction{
			Line: a.line,
			Col:  a.col,
			Data: data,
		}
		a.emitNop(at)
	} else if mov, err := pr.GetConcretePushInst(data); err != nil {
		return err
	} else if mov == nil {
		return fmt.Errorf("TODO: bad push instruction cannot be deduced to concrete push")
	} else {
		return a.emitInst(pr.Instruction{
			Line: a.line,
			Col:  a.col,
			Data: mov,
		}, at)
	}

	return nil
}
//TODO: this needs to be adjusted for symbols that are not defined in the current source
func (a *Assembler) emitGenericJmp(data pr.InstGenericJmp, at int) error {
	if evaluated, err := evalAll(a, &data,
		func(i *pr.InstGenericJmp) **pr.Expr {
			return &(i.Expr)
		},
	); err != nil {
		return err
	} else if !evaluated {
		a.unresolvedJumps[at] = unresolvedJump{
			Expr: data.Expr,
			PatchTy: PatchJmp{
				Variant: data.Variant,
			},
			Absolute: data.Absolute,
		}
		a.emitNop(at)
	} else if jmp, err := pr.GetConcreteJmpInst(data); err != nil {
		return err
	} else if jmp == nil {
		return fmt.Errorf("TODO: bad jmp instruction cannot be deduced to concrete jmp")
	} else if jmpi, ok := jmp.(pr.InstJmpI); ok && !data.Absolute {
		// in case this is not an absolute direct jump, emit it as an IP relative jump
		// this will reduce the number of relocations needed
		posAsInstAddr := uint64(at + vm.ADDRESSDEADZONE_SIZE)
		diff := int64(jmpi.Address) - int64(posAsInstAddr)
		return a.emitJmpIP0R(
			pr.InstJmpIP0R{
				Offset: diff,
				OpTy: vm.OP_TADD,
				JmpTy: jmpi.JmpTy,
			},
			at,
		)
	} else {
		return a.emitInst(pr.Instruction{
			Line: a.line,
			Col:  a.col,
			Data: jmp,
		}, at)
	}
	return nil
}
//TODO: this needs to be adjusted for symbols that are not defined in the current source
func (a *Assembler) emitGenericCall(data pr.InstGenericCall, at int) error {
	// call := a.opCodes.GetBytes(vm.OP_CALL)
	if evaluated, err := evalAll(a, &data,
		func(i *pr.InstGenericCall) **pr.Expr {
			return &(i.Expr)
		},
	); err != nil {
		return err
	} else if !evaluated {
		a.unresolvedJumps[at] = unresolvedJump{
			Expr: data.Expr,
			PatchTy: PatchCall{},
			Absolute: false,
		}
		a.emitNop(at)
	} else if call, err := pr.GetConcreteCallInst(data); err != nil {
		return err
	} else if call == nil {
		return fmt.Errorf("TODO: bad call instruction cannot be deduced to concrete call")
	} else if calli, ok := call.(pr.InstCallI); ok {
		// same case as with InstJmpI
		posAsInstAddr := uint64(at + vm.ADDRESSDEADZONE_SIZE)
		diff := int64(calli.Address) - int64(posAsInstAddr)
		return a.emitCallIP0R(
			pr.InstCallIP0R{
				Offset: diff,
				OpTy: vm.OP_TADD,
			},
			at,
		)
	} else {
		return a.emitInst(pr.Instruction{
			Line: a.line,
			Col:  a.col,
			Data: call,
		}, at)
	}
	return nil
	//direct call case
	// evaluated, err := a.ev.TryEvaluateExpression(data.Expr)
	// if err != nil {
	// 	return err
	// } else if evaluated != nil {
	// 	if addr, ok := pr.IsConstexprType[pr.ConstExprILit](evaluated); !ok {
	// 		return errors.Expected(
	// 			"Valid address",
	// 			data.Expr.Line,
	// 			data.Expr.Col,
	// 		)
	// 	} else {
	// 		binary.BigEndian.PutUint32(a.bytecode[at:], uint32(call))
	// 		binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(addr.Integer))
	// 	}
	// } else {
	// 	position := at
	// 	a.unresolvedJumps[position] = unresolvedJump{
	// 		Expr:    data.Expr,
	// 		PatchTy: PatchCall{},
	// 	}
	// 	binary.BigEndian.AppendUint32(a.bytecode[at:], uint32(a.opCodes.GetBytes(vm.OP_NOP)))
	// 	binary.BigEndian.AppendUint64(a.bytecode[at+4:], 0)
	// }
	// return nil
}
