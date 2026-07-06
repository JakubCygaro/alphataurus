package assembler

import (
	"fmt"

	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
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
	} else if arth, err := GetConcreteArthInst(data); err != nil {
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
	} else if mov, err := GetConcreteMovInst(data); err != nil {
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
	} else if mov, err := GetConcreteLogicalInst(data); err != nil {
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
	} else if mov, err := GetConcreteCmpInst(data); err != nil {
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
	} else if mov, err := GetConcretePushInst(data); err != nil {
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
