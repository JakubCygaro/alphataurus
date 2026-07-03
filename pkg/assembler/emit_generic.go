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
) bool {
	expr := slc(inst)
	if eval, ok := a.ev.TryEvaluateExpression(*expr); ok {
		*expr = eval
		return true
	} else {
		return false
	}
}

func (a *Assembler) emitGenericArth(data pr.InstGenericArth, at int) error {
	evaluated := true
	evaluated = evaluated &&
		evalInstField(a, &data, func(i *pr.InstGenericArth) **pr.Expr {
			return &(i.Dest)
		})
	evaluated = evaluated &&
		evalInstField(a, &data, func(i *pr.InstGenericArth) **pr.Expr {
			return &(i.Src)
		})
	if !evaluated {
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
	evaluated := true
	evaluated = evaluated &&
		evalInstField(a, &data, func(i *pr.InstGenericMov) **pr.Expr {
			return &(i.Src)
		})
	evaluated = evaluated &&
		evalInstField(a, &data, func(i *pr.InstGenericMov) **pr.Expr {
			return &(i.Dest)
		})
	if !evaluated {
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
	evaluated := true
	evaluated = evaluated &&
		evalInstField(a, &data, func(i *pr.InstGenericLogical) **pr.Expr {
			return &(i.First)
		})
	evaluated = evaluated &&
		evalInstField(a, &data, func(i *pr.InstGenericLogical) **pr.Expr {
			return &(i.Second)
		})
	if !evaluated {
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
	evaluated := true
	evaluated = evaluated &&
		evalInstField(a, &data, func(i *pr.InstGenericCmp) **pr.Expr {
			return &(i.Min)
		})
	evaluated = evaluated &&
		evalInstField(a, &data, func(i *pr.InstGenericCmp) **pr.Expr {
			return &(i.Sub)
		})
	if !evaluated {
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
