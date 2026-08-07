package assembler

import (
	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
	decls "github.com/JakubCygaro/alphataurus/pkg/vm/decls"
)

func (a *Assembler) EmitBytecode() error {
	var ok bool
	var err error = nil
	ok, err = a.parser.ParseNext()
	for ; ok && err == nil; ok, err = a.parser.ParseNext() {
		inst := a.parser.CurrentInst()
		a.cInst = &inst
		increment := false
		switch inst.Data.(type) {
		// in case this is a non-emit declaration, do not allocate instruction space
		case pr.InstLab:
		case pr.InstEntry:
		default:
			// allocate instruction space, the emitInst() function does not allocate it
			dummy := [decls.INSTRUCTION_SIZE]byte{}
			a.bytecode = append(a.bytecode, dummy[:]...)
			increment = true
		}
		if emitErr := a.emitInst(inst, a.pos); emitErr != nil {
			return emitErr
		}
		if increment {
			a.pos += decls.INSTRUCTION_SIZE
			a.instCount++
		}
		a.cInst = nil
	}
	return err
}

// emit instruction bytecode *at* specified point in the bytecode array
// the instruction space must be allocated, ie. *at* is a valid index into a.bytecode
// otherwise, you are fucked
func (a *Assembler) emitInst(inst pr.Instruction, at int) error {
	a.cInst = &inst
	defer func() { a.cInst = nil }()
	var err error
	switch i := inst.Data.(type) {
	case pr.InstGenericMov:
		err = a.emitGenericMov(i, a.cInst, at)
	case pr.InstMovIR:
		err = a.emitMovIR(i, at)
	case pr.InstMovZXIR:
		err = a.emitMovZXIR(i, at)
	case pr.InstMovRR:
		err = a.emitMovRR(i, at)
	case pr.InstMovSXRR:
		err = a.emitMovSXRR(i, at)
	case pr.InstGenericMovZX:
		err = a.emitGenericMovZX(i, a.cInst, at)
	case pr.InstMovZXRR:
		err = a.emitMovZXRR(i, at)
	case pr.InstMovDR:
		err = a.emitMovDR(i, at)
	case pr.InstMovZXDR:
		err = a.emitMovZXDR(i, at)
	case pr.InstMovDRO1:
		err = a.emitMovDRO1(i, at)
	case pr.InstMovZXDRO1:
		err = a.emitMovZXDRO1(i, at)
	case pr.InstMovDRO2:
		err = a.emitMovDRO2(i, at)
	case pr.InstMovZXDRO2:
		err = a.emitMovZXDRO2(i, at)
	case pr.InstMovID:
		err = a.emitMovID(i, at)
	case pr.InstMovRD:
		err = a.emitMovRD(i, at)
	case pr.InstMovIDO1:
		err = a.emitMovIDO1(i, at)
	case pr.InstMovRDO1:
		err = a.emitMovRDO1(i, at)
	case pr.InstMovIDO2:
		err = a.emitMovIDO2(i, at)
	case pr.InstMovRDO2:
		err = a.emitMovRDO2(i, at)
	case pr.InstGenericXCHG:
		err = a.emitGenericXCHG(i, a.cInst, at)
	case pr.InstXCHGRR:
		err = a.emitXCHGRR(i, at)
	case pr.InstXCHGDR:
		err = a.emitXCHGDR(i, at)
	case pr.InstXCHGDRO1:
		err = a.emitXCHGDRO1(i, at)
	case pr.InstXCHGDRO2:
		err = a.emitXCHGDRO2(i, at)
	case pr.InstGenericLea:
		err = a.emitGenericLea(i, a.cInst, at)
	case pr.InstLeaO1:
		err = a.emitLeaO1(i, at)
	case pr.InstLeaO2:
		err = a.emitLeaO2(i, at)
	case pr.InstMovSB:
		err = a.emitMovSB(i, at)
	case pr.InstMovSQ:
		err = a.emitMovSQ(i, at)
	case pr.InstMovSH:
		err = a.emitMovSH(i, at)
	case pr.InstMovSW:
		err = a.emitMovSW(i, at)
	case pr.InstGenericArth:
		err = a.emitGenericArth(i, a.cInst, at)
	case pr.InstArthRR:
		err = a.emitArthRR(i, at)
	case pr.InstArthIR:
		err = a.emitArthIR(i, at)
	case pr.InstNotR:
		err = a.emitNot(i, at)
	case pr.InstRotate:
		err = a.emitRotate(i, at)
	case pr.InstNegR:
		err = a.emitNegR(i, at)
	case pr.InstGenericLogical:
		err = a.emitGenericLogical(i, a.cInst, at)
	case pr.InstLogicalRR:
		err = a.emitLogRR(i, at)
	case pr.InstLogicalIR:
		err = a.emitLogIR(i, at)
	case pr.InstInc:
		err = a.emitInc(i, at)
	case pr.InstDec:
		err = a.emitDec(i, at)
	case pr.InstGenericCmp:
		err = a.emitGenericCmp(i, a.cInst, at)
	case pr.InstCmpRR:
		err = a.emitCmpRR(i, at)
	case pr.InstCmpIR:
		err = a.emitCmpIR(i, at)
	case pr.InstGenericJmp:
		err = a.emitGenericJmp(i, a.cInst, at)
	case pr.InstJmpI:
		err = a.emitJmpI(i, at)
	case pr.InstJmpIP0R:
		err = a.emitJmpIP0R(i, at)
	case pr.InstJmpIP1R:
		err = a.emitJmpIP1R(i, at)
	case pr.InstLab:
		err = a.declareLabel(i, at)
	case pr.InstGenericPush:
		err = a.emitGenericPush(i, a.cInst, at)
	case pr.InstPushR:
		err = a.emitPushR(i, at)
	case pr.InstPushI:
		err = a.emitPushI(i, at)
	case pr.InstPop:
		err = a.emitPop(i, at)
	case pr.InstPopR:
		err = a.emitPopR(i, at)
	case pr.InstNop:
		err = a.emitNop(at)
	case pr.InstGenericCall:
		err = a.emitGenericCall(i, a.cInst, at)
	case pr.InstCallI:
		err = a.emitCallI(i, at)
	case pr.InstCallIP0R:
		err = a.emitCallIP0R(i, at)
	case pr.InstCallIP1R:
		err = a.emitCallIP1R(i, at)
	case pr.InstRet:
		err = a.emitRet(at)
	case pr.InstExitI:
		err = a.emitExitI(i, at)
	case pr.InstExitR:
		err = a.emitExitR(i, at)
	case pr.InstEntry:
		if a.hasEntry {
			err = errors.
				MakeAssemblerError(
					inst.Line,
					inst.Col,
					"More than one entry point declaration is not allowed.",
				)

		} else {
			a.hasEntry = true
			_, ent := a.codePos(at)
			a.entry = ent
		}
	case pr.InstClr:
		err = a.emitClr(at)
	case pr.InstSDF:
		err = a.emitSDF(at)
	case pr.InstCDF:
		err = a.emitCDF(at)
	default:
		a.lastInst = inst
		return err
	}
	if err != nil {
		return err
	}
	return nil
}
