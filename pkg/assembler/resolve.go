package assembler

import (
	"fmt"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
	decls "github.com/JakubCygaro/alphataurus/pkg/vm/decls"
	aobj "github.com/JakubCygaro/alphataurus/pkg/vm/obj"
)

func (a *Assembler) resolveWithSymbols(
	at int, unr unresolvedJump, syms []*aobj.SymbolData,
) error {
	if len(syms) > 1 {
		return fmt.
			Errorf("TODO: too many symbols to resolve")
	}
	sym := syms[0]
	_, symIdx, ok := a.symbols.GetByName(sym.GetName())
	if !ok {
		return errors.
			MakeAssemblerError(
				unr.At.Line,
				unr.Expr.Col,
				"Symbol `%s` could not be resolved.",
				sym.GetName(),
			)
	}
	switch sym.Vis {
	case aobj.SYM_VPRIVATE:
		fallthrough
	case aobj.SYM_VEXPORT:
		var err error
		switch p := unr.PatchTy.(type) {
		case PatchCall:
			err = a.patchCallIP(sym, at)
		case PatchJmp:
			if unr.Absolute {
				err = a.patchJmp(p, at, sym.Loc)
				reloc := aobj.RelocData{
					Loc:       uint64(at) + decls.OPCODE_SIZE,
					Ref:       uint64(symIdx),
					PatchSize: 8,
				}
				a.relocations = append(a.relocations, reloc)
			} else {
				err = a.patchJmpIP(p, sym, at)
			}
		}
		if err != nil {
			return err
		}
	default:
		var err error
		switch p := unr.PatchTy.(type) {
		case PatchCall:
			err = a.patchCall(at, 0)
		case PatchJmp:
			err = a.patchJmp(p, at, 0)
		}
		if err != nil {
			return err
		}
		reloc := aobj.RelocData{
			Loc:       uint64(at) + decls.OPCODE_SIZE,
			Ref:       uint64(symIdx),
			PatchSize: 8,
		}
		a.relocations = append(a.relocations, reloc)
	}
	return nil
}

func (a *Assembler) resolveJumpInsturctions() error {
	for codePos, unresolved := range a.unresolvedJumps {
		var addr uint64
		if eval, ok, err :=
			a.ev.TryConstEvaluateExpression(unresolved.Expr); err != nil {
			return err
		} else if _, isIden :=
			pr.IsConstexprType[pr.ExprIden](unresolved.Expr); isIden &&
			!ok && len(a.prov.LastFailedSymbols) == 1 {
			if err :=
				a.resolveWithSymbols(
					codePos, unresolved, a.prov.LastFailedSymbols); err != nil {
				return err
			}
			continue
		} else if !ok && len(a.prov.LastFailedSymbols) == 0 {
			return errors.
				MakeAssemblerError(
					unresolved.At.Line,
					unresolved.Expr.Col,
					"Unresolved symbol 1.",
				)
		} else if cepxr, ok := eval.Val.(pr.ConstExprILit); !ok {
			return errors.
				MakeAssemblerError(
					unresolved.At.Line,
					unresolved.Expr.Col,
					"Unresolved symbol 2.",
				)
		} else {
			addr = cepxr.Integer
		}
		sym, ok := a.symbols.GetByLocation(addr)
		if !ok {
			return errors.
				MakeAssemblerError(
					unresolved.At.Line,
					unresolved.Expr.Col,
					"Unresolved symbol 3.",
				)
		}
		if err := a.resolveWithSymbols(
			codePos, unresolved, []*aobj.SymbolData{sym}); err != nil {
			return err
		}
	}
	return nil
}
func (a *Assembler) resolveSymbols() error {
	for sname, idx := range a.symbols.ByName {
		symbol := a.symbols.InOrder[idx]
		switch symbol.Vis {
		case aobj.SYM_VPRIVATE:
			fallthrough
		case aobj.SYM_VEXPORT:
			if symbol.Loc == 0 {
				return errors.
					MakeAssemblerError(
						0,
						0,
						"Unresolved symbol `%s`.",
						sname,
					)
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
