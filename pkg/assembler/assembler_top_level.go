package assembler

import (
	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
	"github.com/JakubCygaro/alphataurus/pkg/vm/obj"
)

func (a *Assembler) handleExport(data pr.InstExport) error {
	if s, ok := a.symbols.ByName[data.Name]; ok {
		definedAt := a.definedAt[s]
		return errors.
			MakeAssemblerError(
				a.cInst.Line,
				a.cInst.Col,
				"Multiple symbol `%s` definitions."+
					"\nFirst defined at (%v:%v).\n",
				definedAt.Line, definedAt.Col,
			)
	}
	sym := vm.DefineSymbol(
		vm.SYM_TFUNC,
		vm.SYM_VEXPORT,
		0,
		data.Name,
	)
	if s, err := a.symbols.AddSymbol(sym); err != nil {
		return errors.MakeAssemblerError(
				a.cInst.Line,
				a.cInst.Col,
			"Failed to declare export symbol, %s.",
			err.Error(),
		)
	} else {
		a.definedAt[s] = Point{a.cInst.Line, a.cInst.Col}
	}
	return nil
}
func (a *Assembler) handleImport(data pr.InstImport) error {
	if s, _, ok := a.symbols.GetByName(data.Name); ok {
		definedAt := a.definedAt[s]
		return errors.
			MakeAssemblerError(
				a.cInst.Line,
				a.cInst.Col,
				"Multiple symbol `%s` definitions."+
					"\nFirst defined at (%v:%v).\n",
				definedAt.Line, definedAt.Col,
			)
	}
	sym := vm.DefineSymbol(
		vm.SYM_TFUNC,
		vm.SYM_VIMPORTSTRONG,
		0,
		data.Name,
	)
	if data.Weak {
		sym.Vis = vm.SYM_VIMPORTWEAK
	}
	if s, err := a.symbols.AddSymbol(sym); err != nil {
		return errors.MakeAssemblerError(
				a.cInst.Line,
				a.cInst.Col,
			"Failed to declare import symbol, %s.",
			err.Error(),
		)
	} else {
		a.definedAt[s] = Point{a.cInst.Line, a.cInst.Col}
	}
	return nil
}
