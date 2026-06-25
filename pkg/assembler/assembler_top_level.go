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
	sym := vm.SymbolData{
		Ty:   vm.SYM_TFUNC,
		Vis:  vm.SYM_VEXPORT,
		Loc:  0,
		Name: data.Name,
	}
	if _, ok := a.symbols.AddSymbol(sym); !ok {
		return fmt.Errorf("Failed to declare export symbol")
	}
	return nil
}
func (a *Assembler) handleImport(data pr.InstImport) error {
	if _, _, ok := a.symbols.GetByName(data.Name); ok {
		return errors.MultipleSymbolDefinitions(data.Name, a.line, a.col)
	}
	sym := vm.SymbolData{
		Ty:   vm.SYM_TFUNC,
		Vis:  vm.SYM_VIMPORTSTRONG,
		Loc:  0,
		Name: data.Name,
	}
	if data.Weak {
		sym.Vis = vm.SYM_VIMPORTWEAK
	}
	a.symbols.AddSymbol(sym)
	return nil
}
