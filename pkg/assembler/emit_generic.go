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

func (a *Assembler) emitGenericArth(
	data pr.InstGenericArth,
	outer *pr.Instruction,
	at int,
) error {
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
			Line: outer.Line,
			Col:  outer.Col,
			Data: data,
		}
		a.emitNop(at)
	} else if arth, err := pr.GetConcreteArthInst(data, outer); err != nil {
		return err
	} else if arth == nil {
		return fmt.Errorf("TODO: bad arth instruction cannot be deduced to concrete arth")
	} else {
		return a.emitInst(pr.Instruction{
			Line: outer.Line,
			Col:  outer.Col,
			Data: arth,
		}, at)
	}
	return nil
}

func (a *Assembler) emitGenericMov(
	data pr.InstGenericMov,
	outer *pr.Instruction,
	at int,
) error {
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
			Line: outer.Line,
			Col:  outer.Col,
			Data: data,
		}
		a.emitNop(at)
	} else if mov, err := pr.GetConcreteMovInst(data, outer); err != nil {
		return err
	} else if mov == nil {
		return fmt.Errorf("TODO: bad mov instruction cannot be deduced to concrete mov")
	} else {
		return a.emitInst(pr.Instruction{
			Line: outer.Line,
			Col:  outer.Col,
			Data: mov,
		}, at)
	}

	return nil
}
func (a *Assembler) emitGenericLogical(
	data pr.InstGenericLogical,
	outer *pr.Instruction,
	at int,
) error {
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
			Line: outer.Line,
			Col:  outer.Col,
			Data: data,
		}
		a.emitNop(at)
	} else if mov, err := pr.GetConcreteLogicalInst(data, outer); err != nil {
		return err
	} else if mov == nil {
		return fmt.
			Errorf("TODO: bad logical instruction cannot be deduced to concrete logical")
	} else {
		return a.emitInst(pr.Instruction{
			Line: outer.Line,
			Col:  outer.Col,
			Data: mov,
		}, at)
	}

	return nil
}
func (a *Assembler) emitGenericCmp(
	data pr.InstGenericCmp,
	outer *pr.Instruction,
	at int,
) error {
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
			Line: outer.Line,
			Col:  outer.Col,
			Data: data,
		}
		a.emitNop(at)
	} else if mov, err := pr.GetConcreteCmpInst(data); err != nil {
		return err
	} else if mov == nil {
		return fmt.Errorf("TODO: bad mov instruction cannot be deduced to concrete mov")
	} else {
		return a.emitInst(pr.Instruction{
			Line: outer.Line,
			Col:  outer.Col,
			Data: mov,
		}, at)
	}

	return nil
}
func (a *Assembler) emitGenericPush(
	data pr.InstGenericPush,
	outer *pr.Instruction,
	at int,
) error {
	if ok, err := evalAll(a, &data,
		func(i *pr.InstGenericPush) **pr.Expr {
			return &(i.Expr)
		},
	); err != nil {
		return err
	} else if !ok {
		a.unevalInsts[at] = pr.Instruction{
			Line: outer.Line,
			Col:  outer.Col,
			Data: data,
		}
		a.emitNop(at)
	} else if mov, err := pr.GetConcretePushInst(data); err != nil {
		return err
	} else if mov == nil {
		return fmt.Errorf("TODO: bad push instruction cannot be deduced to concrete push")
	} else {
		return a.emitInst(pr.Instruction{
			Line: outer.Line,
			Col:  outer.Col,
			Data: mov,
		}, at)
	}

	return nil
}

func (a *Assembler) emitGenericJmp(
	data pr.InstGenericJmp,
	outer *pr.Instruction,
	at int,
) error {
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
			At:       Point{outer.Line, outer.Col},
		}
		a.emitNop(at)
	} else if jmp, err := pr.GetConcreteJmpInst(data, outer); err != nil {
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
				OpTy:   vm.OP_TADD,
				JmpTy:  jmpi.JmpTy,
			},
			at,
		)
	} else {
		return a.emitInst(pr.Instruction{
			Line: outer.Line,
			Col:  outer.Col,
			Data: jmp,
		}, at)
	}
	return nil
}

func (a *Assembler) emitGenericCall(
	data pr.InstGenericCall,
	outer *pr.Instruction,
	at int,
) error {
	if evaluated, err := evalAll(a, &data,
		func(i *pr.InstGenericCall) **pr.Expr {
			return &(i.Expr)
		},
	); err != nil {
		return err
	} else if !evaluated {
		a.unresolvedJumps[at] = unresolvedJump{
			Expr:     data.Expr,
			PatchTy:  PatchCall{},
			Absolute: false,
			At:       Point{outer.Line, outer.Col},
		}
		a.emitNop(at)
	} else if call, err := pr.GetConcreteCallInst(data, outer); err != nil {
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
				OpTy:   vm.OP_TADD,
			},
			at,
		)
	} else {
		return a.emitInst(pr.Instruction{
			Line: outer.Line,
			Col:  outer.Col,
			Data: call,
		}, at)
	}
	return nil
}
func (a *Assembler) emitGenericMovZX(
	data pr.InstGenericMovZX,
	outer *pr.Instruction,
	at int,
) error {
	if ok, err := evalAll(a, &data,
		func(i *pr.InstGenericMovZX) **pr.Expr {
			return &(i.Mov.Src)
		},
		func(i *pr.InstGenericMovZX) **pr.Expr {
			return &(i.Mov.Dest)
		},
	); err != nil {
		return err
	} else if !ok {
		a.unevalInsts[at] = pr.Instruction{
			Line: outer.Line,
			Col:  outer.Col,
			Data: data,
		}
		a.emitNop(at)
	} else if mov, err := pr.GetConcreteMovZXInst(data, outer); err != nil {
		return err
	} else if mov == nil {
		return fmt.Errorf(
			"TODO: bad movzx instruction cannot be deduced to concrete movzx")
	} else {
		return a.emitInst(pr.Instruction{
			Line: outer.Line,
			Col:  outer.Col,
			Data: mov,
		}, at)
	}

	return nil
}
func (a *Assembler) emitGenericXCHG(
	data pr.InstGenericXCHG,
	outer *pr.Instruction,
	at int,
) error {
	if ok, err := evalAll(a, &data,
		func(i *pr.InstGenericXCHG) **pr.Expr {
			return &(i.Mov.Src)
		},
		func(i *pr.InstGenericXCHG) **pr.Expr {
			return &(i.Mov.Dest)
		},
	); err != nil {
		return err
	} else if !ok {
		a.unevalInsts[at] = pr.Instruction{
			Line: outer.Line,
			Col:  outer.Col,
			Data: data,
		}
		a.emitNop(at)
	} else if mov, err := pr.GetConcreteXChgInst(data, outer); err != nil {
		return err
	} else if mov == nil {
		return fmt.Errorf(
			"TODO: bad xchg instruction cannot be deduced to concrete xchg")
	} else {
		return a.emitInst(pr.Instruction{
			Line: outer.Line,
			Col:  outer.Col,
			Data: mov,
		}, at)
	}

	return nil
}
func (a *Assembler) emitGenericLea(
	data pr.InstGenericLea,
	outer *pr.Instruction,
	at int,
) error {
	if ok, err := evalAll(a, &data,
		func(i *pr.InstGenericLea) **pr.Expr {
			return &(i.Mov.Src)
		},
		func(i *pr.InstGenericLea) **pr.Expr {
			return &(i.Mov.Dest)
		},
	); err != nil {
		return err
	} else if !ok {
		a.unevalInsts[at] = pr.Instruction{
			Line: outer.Line,
			Col:  outer.Col,
			Data: data,
		}
		a.emitNop(at)
	} else if mov, err := pr.GetConcreteLeaInst(data, outer); err != nil {
		return err
	} else if mov == nil {
		return fmt.Errorf(
			"TODO: bad lea instruction cannot be deduced to concrete lea")
	} else {
		return a.emitInst(pr.Instruction{
			Line: outer.Line,
			Col:  outer.Col,
			Data: mov,
		}, at)
	}

	return nil
}
