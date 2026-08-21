package assembler

import (
	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func (a *Assembler) jmpInstToOpCode(ty pr.JmpVariant) uint32 {
	var opcode uint32
	switch ty {
	case pr.JMP:
		opcode = vm.OP_JMP_VAL
	case pr.JMPE:
		opcode = vm.OP_JMPE_VAL
	case pr.JMPNE:
		opcode = vm.OP_JMPNE_VAL
	case pr.JMPZ:
		opcode = vm.OP_JMPZ_VAL
	case pr.JMPNZ:
		opcode = vm.OP_JMPNZ_VAL
	case pr.JMPG:
		opcode = vm.OP_JMPG_VAL
	case pr.JMPGE:
		opcode = vm.OP_JMPGE_VAL
	case pr.JMPL:
		opcode = vm.OP_JMPL_VAL
	case pr.JMPLE:
		opcode = vm.OP_JMPLE_VAL
	case pr.JMPS:
		opcode = vm.OP_JMPS_VAL
	case pr.JMPNS:
		opcode = vm.OP_JMPNS_VAL
	case pr.JMPC:
		opcode = vm.OP_JMPC_VAL
	case pr.JMPNC:
		opcode = vm.OP_JMPNC_VAL
	case pr.JMPO:
		opcode = vm.OP_JMPO_VAL
	case pr.JMPNO:
		opcode = vm.OP_JMPNO_VAL
	}
	return opcode
}

func (a *Assembler) absoluteJmpToIPJmp(instTy pr.JmpVariant) (opcode uint32) {
	switch instTy {
	case pr.JMP:
		opcode = vm.OP_JMPIP_VAL
	case pr.JMPE:
		opcode = vm.OP_JMPEIP_VAL
	case pr.JMPNE:
		opcode = vm.OP_JMPNEIP_VAL
	case pr.JMPZ:
		opcode = vm.OP_JMPZIP_VAL
	case pr.JMPNZ:
		opcode = vm.OP_JMPNZIP_VAL
	case pr.JMPG:
		opcode = vm.OP_JMPGIP_VAL
	case pr.JMPGE:
		opcode = vm.OP_JMPGEIP_VAL
	case pr.JMPL:
		opcode = vm.OP_JMPLIP_VAL
	case pr.JMPLE:
		opcode = vm.OP_JMPLEIP_VAL
	case pr.JMPS:
		opcode = vm.OP_JMPSIP_VAL
	case pr.JMPNS:
		opcode = vm.OP_JMPNSIP_VAL
	case pr.JMPC:
		opcode = vm.OP_JMPCIP_VAL
	case pr.JMPNC:
		opcode = vm.OP_JMPNCIP_VAL
	case pr.JMPO:
		opcode = vm.OP_JMPOIP_VAL
	case pr.JMPNO:
		opcode = vm.OP_JMPNOIP_VAL
	}
	return opcode
}
