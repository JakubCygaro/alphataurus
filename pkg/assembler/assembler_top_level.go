package assembler

import (
	"fmt"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
)

func (a *Assembler) handleExport(data InstImportExportData) error {
	if _, ok := a.symbols.ByName[data.Name]; ok {
		return errors.MultipleSymbolDefinitions(data.Name, a.parser.lexer.line, a.parser.lexer.col)
	}
	sym := SymbolData{
		Ty:  SYM_TFUNC,
		Vis: SYM_VEXPORT,
		Loc: 0,
		Name: data.Name,
	}
	if _, ok := a.symbols.AddSymbol(sym); !ok {
		return fmt.Errorf("Failed to declare export symbol")
	}
	return nil
}
func (a *Assembler) handleImport(data InstImportExportData) error {
	if _, _, ok := a.symbols.GetByName(data.Name); ok {
		return errors.MultipleSymbolDefinitions(data.Name, a.parser.lexer.line, a.parser.lexer.col)
	}
	sym := SymbolData{
		Ty:  SYM_TFUNC,
		Vis: SYM_VIMPORTSTRONG,
		Loc: 0,
		Name: data.Name,
	}
	if data.Weak {
		sym.Vis = SYM_VIMPORTWEAK
	}
	a.symbols.AddSymbol(sym)
	return nil
}
