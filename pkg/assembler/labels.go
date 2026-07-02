package assembler

import (
	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
	decls "github.com/JakubCygaro/alphataurus/pkg/vm/decls"
	aobj "github.com/JakubCygaro/alphataurus/pkg/vm/obj"
)

func (a *Assembler) declareLabel(data pr.InstLab, at int) error {
	// this needs to be the address of the function in the virtual address space
	_, posAsInstAddr := a.codePos(at)
	posAsInstAddr -= decls.INSTRUCTION_SIZE
	if sym, _, ok := a.symbols.GetByName(data.Label); ok {
		switch sym.Vis {
		case aobj.SYM_VEXPORT:
			if sym.Loc != 0 {
				return errors.RedeclaredLabel(data.Label, data.DeclaredAt,
					a.line, a.col)
			} else {
				(*sym).Loc = posAsInstAddr
			}
		case aobj.SYM_VPRIVATE:
			return errors.RedeclaredLabel(data.Label, data.DeclaredAt,
				a.line, a.col)
		default:
			return errors.ImportedSymbolDeclared(data.Label, a.line, a.col)
		}
	} else {
		lab := aobj.SymbolData{
			Ty:   aobj.SYM_TFUNC,
			Vis:  aobj.SYM_VPRIVATE,
			Loc:  posAsInstAddr,
			Name: data.Label,
		}
		a.symbols.AddSymbol(lab)
	}
	return nil
}
