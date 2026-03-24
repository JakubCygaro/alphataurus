package assembler

import "github.com/JakubCygaro/alphataurus/pkg/assembler/errors"

func (a *Assembler) handleExport(data InstImportExportData) error {
	if _, ok := a.symbols[data.Name]; ok {
		return errors.MultipleSymbolDefinitions(data.Name, a.parser.lexer.line, a.parser.lexer.col)
	}
	sym := SymbolData{
		Ty:  SYM_TFUNC,
		Vis: SYM_VEXPORT,
		Loc: 0,
	}
	a.symbols[data.Name] = sym
	return nil
}
func (a *Assembler) handleImport(data InstImportExportData) error {
	if _, ok := a.symbols[data.Name]; ok {
		return errors.MultipleSymbolDefinitions(data.Name, a.parser.lexer.line, a.parser.lexer.col)
	}
	sym := SymbolData{
		Ty:  SYM_TFUNC,
		Vis: SYM_VIMPORTSTRONG,
		Loc: 0,
	}
	a.symbols[data.Name] = sym
	return nil
}
