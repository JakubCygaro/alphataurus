package assembler

import (
	"fmt"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
	"github.com/JakubCygaro/alphataurus/pkg/vm/obj"
)

func (a *Assembler) handleExport(data pr.InstExport) error {
	if _, ok := a.symbols.ByName[data.Name]; ok {
		return errors.MultipleSymbolDefinitions(data.Name, a.line,
			a.col)
	}
	sym := vm.DefineSymbol(
		vm.SYM_TFUNC,
		vm.SYM_VEXPORT,
		0,
		data.Name,
	)
	if _, err := a.symbols.AddSymbol(sym); err != nil {
		return fmt.Errorf("Failed to declare export symbol: %s", err.Error())
	}
	return nil
}
func (a *Assembler) handleImport(data pr.InstImport) error {
	if _, _, ok := a.symbols.GetByName(data.Name); ok {
		return errors.MultipleSymbolDefinitions(data.Name, a.line, a.col)
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
	_, err := a.symbols.AddSymbol(sym)
	return err
}
