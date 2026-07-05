package assembler

import (
	"fmt"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
	decls "github.com/JakubCygaro/alphataurus/pkg/vm/decls"
	aobj "github.com/JakubCygaro/alphataurus/pkg/vm/obj"
)

func (a *Assembler) resolveJumpInsturctions() error {
	for codePos, unresolved := range a.unresolvedJumps {
		var addr uint64
		if eval, ok :=
			a.ev.TryConstEvaluateExpression(unresolved.Expr); !ok {
			return errors.UnresolvedSymbol(fmt.Sprint(codePos))
		} else if cepxr, ok := eval.Val.(pr.ConstExprILit); !ok {
			return errors.UnresolvedSymbol(fmt.Sprint(codePos))
		} else {
			addr = cepxr.Integer
		}
		sym, ok := a.symbols.GetByLocation(addr)
		_, symIdx, _ := a.symbols.GetByName(sym.GetName())
		if !ok {
			return errors.UnresolvedSymbol(fmt.Sprint(codePos))
		}
		switch sym.Vis {
		case aobj.SYM_VPRIVATE:
			fallthrough
		case aobj.SYM_VEXPORT:
			var err error
			switch p := unresolved.PatchTy.(type) {
			case PatchCall:
				err = a.patchCallIP(sym, codePos)
			case PatchJmp:
				if unresolved.Absolute {
					err = a.patchJmp(p, codePos, sym.Loc)
					reloc := aobj.RelocData{
						Loc:       uint64(codePos) + decls.OPCODE_SIZE,
						Ref:       uint64(symIdx),
						PatchSize: 8,
					}
					a.relocations = append(a.relocations, reloc)
				} else {
					err = a.patchJmpIP(p, sym, codePos)
				}
			}
			// if unresolved.PatchTy == PATCH_CALL {
			// 	err = a.patchCallIP(sym, codePos)
			// } else if unresolved.Absolute {
			// 	err = a.patchJmp(unresolved, codePos, sym.Loc)
			// 	reloc := aobj.RelocData{
			// 		Loc:       uint64(codePos) + decls.OPCODE_SIZE,
			// 		Ref:       uint64(symIdx),
			// 		PatchSize: 8,
			// 	}
			// 	a.relocations = append(a.relocations, reloc)
			// } else {
			// 	err = a.patchJmpIP(unresolved, sym, codePos)
			// }
			if err != nil {
				return err
			}
		default:
			var err error
			switch p := unresolved.PatchTy.(type) {
			case PatchCall:
				err = a.patchCall(codePos, 0)
			case PatchJmp:
				err = a.patchJmp(p, codePos, 0)
			}
			// if unresolved.PatchTy == PATCH_CALL {
			// 	err = a.patchCall(codePos, 0)
			// } else {
			// 	err = a.patchJmp(unresolved, codePos, 0)
			// }
			if err != nil {
				return err
			}
			reloc := aobj.RelocData{
				Loc:       uint64(codePos) + decls.OPCODE_SIZE,
				Ref:       uint64(symIdx),
				PatchSize: 8,
			}
			a.relocations = append(a.relocations, reloc)
		}
	}
	return nil
}
func (a *Assembler) resolveSymbols() error {
	for sname, idx := range a.symbols.ByName {
		symbol := a.symbols.InOrder[idx]
		switch symbol.Vis {
		case aobj.SYM_VPRIVATE:
			if symbol.Loc == 0 {
				return errors.UnresolvedSymbol(sname)
			}
		case aobj.SYM_VEXPORT:
			if symbol.Loc == 0 {
				return errors.UnresolvedSymbol(sname)
			}
		}
	}
	return nil
}
func (a *Assembler) resolveUnevaluated() error {
	startingUnev := len(a.unevalInsts)
	// change in unevaluted instructions
	diff := -1
	for diff < 0 {
		for pos := range a.unevalInsts {
			inst := a.unevalInsts[pos]
			delete(a.unevalInsts, pos)
			a.emitInst(inst, pos)
		}
		diff = len(a.unevalInsts) - startingUnev
	}
	if len(a.unevalInsts) != 0 {
		// last := a.prov.LastFailedAccess

		return fmt.
			Errorf("TODO: Unresolved expressions, unable to finish assembling")
	}
	return nil
}
