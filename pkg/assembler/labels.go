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
	if sym, _, ok := a.symbols.GetByName(data.LabelName); ok {
		switch sym.Vis {
		case aobj.SYM_VPRIVATE:
			fallthrough
		case aobj.SYM_VEXPORT:
			if sym.Loc != 0 {
				pos := a.definedAt[sym]
				return errors.
					LabelRedeclared(
						a.cInst.Line, a.cInst.Col,
						data.LabelName,
						pos.Line, pos.Col,
					)
			} else {
				(*sym).Loc = posAsInstAddr
			}
		case aobj.SYM_VIMPORTSTRONG:
			fallthrough
		case aobj.SYM_VIMPORTWEAK:
			return errors.
				MakeAssemblerError(
					a.cInst.Line, a.cInst.Col,
					"Attempted to delcare an import symbol `%s`.",
					sym.GetName(),
				)
		}
	} else {
		lab := aobj.DefineSymbol(
			aobj.SYM_TFUNC,
			aobj.SYM_VPRIVATE,
			posAsInstAddr,
			data.LabelName,
		)
		if s, err := a.symbols.AddSymbol(lab); err != nil {
			return err
		} else {
			a.definedAt[s] = Point{a.cInst.Line, a.cInst.Col}
		}
	}
	return nil
}
