package assembler

import (
	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func (a *Assembler) jmpInstToOpCode(ty pr.JmpVariant) uint32 {
	var opcode uint32
	switch ty {
	case pr.JMP:
		opcode = a.opCodes.GetBytes(vm.OP_JMP)
	case pr.JMPE:
		opcode = a.opCodes.GetBytes(vm.OP_JMPE)
	case pr.JMPNE:
		opcode = a.opCodes.GetBytes(vm.OP_JMPNE)
	case pr.JMPZ:
		opcode = a.opCodes.GetBytes(vm.OP_JMPZ)
	case pr.JMPNZ:
		opcode = a.opCodes.GetBytes(vm.OP_JMPNZ)
	case pr.JMPG:
		opcode = a.opCodes.GetBytes(vm.OP_JMPG)
	case pr.JMPGE:
		opcode = a.opCodes.GetBytes(vm.OP_JMPGE)
	case pr.JMPL:
		opcode = a.opCodes.GetBytes(vm.OP_JMPL)
	case pr.JMPLE:
		opcode = a.opCodes.GetBytes(vm.OP_JMPLE)
	case pr.JMPS:
		opcode = a.opCodes.GetBytes(vm.OP_JMPS)
	case pr.JMPNS:
		opcode = a.opCodes.GetBytes(vm.OP_JMPNS)
	case pr.JMPC:
		opcode = a.opCodes.GetBytes(vm.OP_JMPC)
	case pr.JMPNC:
		opcode = a.opCodes.GetBytes(vm.OP_JMPNC)
	case pr.JMPO:
		opcode = a.opCodes.GetBytes(vm.OP_JMPO)
	case pr.JMPNO:
		opcode = a.opCodes.GetBytes(vm.OP_JMPNO)
	}
	return opcode
}

func (a *Assembler) absoluteJmpToIPJmp(instTy pr.JmpVariant) (opcode uint32) {
	switch instTy {
	case pr.JMP:
		opcode = a.opCodes.GetBytes(vm.OP_JMPIP)
	case pr.JMPE:
		opcode = a.opCodes.GetBytes(vm.OP_JMPEIP)
	case pr.JMPNE:
		opcode = a.opCodes.GetBytes(vm.OP_JMPNEIP)
	case pr.JMPZ:
		opcode = a.opCodes.GetBytes(vm.OP_JMPZIP)
	case pr.JMPNZ:
		opcode = a.opCodes.GetBytes(vm.OP_JMPNZIP)
	case pr.JMPG:
		opcode = a.opCodes.GetBytes(vm.OP_JMPGIP)
	case pr.JMPGE:
		opcode = a.opCodes.GetBytes(vm.OP_JMPGEIP)
	case pr.JMPL:
		opcode = a.opCodes.GetBytes(vm.OP_JMPLIP)
	case pr.JMPLE:
		opcode = a.opCodes.GetBytes(vm.OP_JMPLEIP)
	case pr.JMPS:
		opcode = a.opCodes.GetBytes(vm.OP_JMPSIP)
	case pr.JMPNS:
		opcode = a.opCodes.GetBytes(vm.OP_JMPNSIP)
	case pr.JMPC:
		opcode = a.opCodes.GetBytes(vm.OP_JMPCIP)
	case pr.JMPNC:
		opcode = a.opCodes.GetBytes(vm.OP_JMPNCIP)
	case pr.JMPO:
		opcode = a.opCodes.GetBytes(vm.OP_JMPOIP)
	case pr.JMPNO:
		opcode = a.opCodes.GetBytes(vm.OP_JMPNOIP)
	}
	return opcode
}
