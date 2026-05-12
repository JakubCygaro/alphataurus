package vm

import (
	"encoding/binary"
	"fmt"
	"github.com/JakubCygaro/alphataurus/pkg/vm/errors"
)

func (vm *VmState) load(elf AlphaELFFile) error {
	if !elf.HasEntry {
		return errors.NoEntry()
	}
	bytecode := elf.Data[uint64(elf.HeaderSize)+elf.CodeStart : uint64(elf.HeaderSize)+elf.CodeStart+elf.CodeSize]
	if len(bytecode)%INSTRUCTION_SIZE != 0 {
		return errors.BadCodeSectionSize()
	}
	codeSize := len(bytecode)
	// the code section starts after the deadzone (for now)
	vm.exeSegBase = ADDRESSDEADZONE_SIZE
	// the stack starts after the code section
	vm.stackSegBase = int(vm.exeSegBase) + codeSize
	// the base pointer points to right before the stack
	binary.BigEndian.PutUint64(
		vm.regs.r[BP_IDX][:],
		uint64(vm.stackSegBase)-1,
	)
	// the stack pointer points to the base pointer
	vm.regs.r[SP_IDX] = vm.regs.r[BP_IDX]
	vm.setIp(elf.Entry)

	vm.bytecode = bytecode
	vm.codeSize = uint64(codeSize)
	return nil
}
func (vm *VmState) Execute(elf AlphaELFFile) error {
	if err := vm.load(elf); err != nil {
		return err
	}
	for ; vm.GetIp() < vm.exeSegBase+vm.codeSize && !vm.exit; vm.incIp() {
		_, opCodeBytes, param := vm.fetch()
		if err := vm.decode(opCodeBytes); err != nil {
			return err
		}
		if err := vm.exec(opCodeBytes, param); err != nil {
			return err
		}
	}
	return nil
}
func (state *VmState) fetch() (instAddr uint64, opCodeBytes, param []byte) {
	instAddr = state.VirtToRealIp(state.GetIp())
	state.byteCodePos = state.GetIp()
	opCodeBytes = state.bytecode[instAddr : instAddr+OPCODE_SIZE]
	param = state.bytecode[instAddr+OPCODE_SIZE : instAddr+INSTRUCTION_SIZE]
	return instAddr, opCodeBytes, param
}
func (state *VmState) decode(opCodeBytes []byte) error {
	opcode, err := state.GetOpcode(opCodeBytes)
	state.currentOpcode = uint32(opcode)
	return err
}
func (state *VmState) exec(opCodeBytes, param []byte) error {
	var err error
	opcode := state.currentOpcode
	switch opcode {
	case OP_MOVRR:
		err = state.movRR(opCodeBytes[0], param)
	case OP_MOVIR:
		err = state.movIR(opCodeBytes[0], param)
	case OP_MOVDRI:
		err = state.movDRI(opCodeBytes[0], param)
	case OP_MOVID:
		err = state.movID(opCodeBytes[1], param)
	case OP_MOVRD:
		err = state.movRD(opCodeBytes[1], param)
	case OP_MOVDRO1:
		err = state.movDRO1(opCodeBytes[1], opCodeBytes[0], param)
	case OP_MOVDRO2:
		err = state.movDRO2(opCodeBytes[2],opCodeBytes[1], opCodeBytes[0], param)
	case OP_MOVRDO1:
		err = state.movRDO1(opCodeBytes[1], opCodeBytes[0], param)
	case OP_MOVRDO2:
		err = state.movRDO2(opCodeBytes[2],opCodeBytes[1], opCodeBytes[0], param)
	case OP_MOVIDO1_NO:
		err = state.movIDO1NoOffset(opCodeBytes[1], opCodeBytes[0], param)
	case OP_MOVIDO1:
		err = state.movIDO1(opCodeBytes[1], opCodeBytes[0], param)
	case OP_MOVIDO2_NO:
		err = state.movIDO2NoOffset(opCodeBytes[2],opCodeBytes[1], opCodeBytes[0], param)
	case OP_MOVIDO2:
		err = state.movIDO2(opCodeBytes[2],opCodeBytes[1], opCodeBytes[0], param)
	case OP_ADDRR:
		err = state.arthRR(int(opcode), param)
	case OP_SUBRR:
		err = state.arthRR(int(opcode), param)
	case OP_MULRR:
		err = state.arthRR(int(opcode), param)
	case OP_DIVRR:
		err = state.arthRR(int(opcode), param)
	case OP_ADDIR:
		err = state.arthIR(int(opcode), opCodeBytes[0], param)
	case OP_SUBIR:
		err = state.arthIR(int(opcode), opCodeBytes[0], param)
	case OP_NOT:
		err = state.not(param)
	case OP_ORRR:
		err = state.logRR(int(opcode), param)
	case OP_ANDRR:
		err = state.logRR(int(opcode), param)
	case OP_XORRR:
		err = state.logRR(int(opcode), param)
	case OP_LSHRR:
		err = state.logRR(int(opcode), param)
	case OP_RSHRR:
		err = state.logRR(int(opcode), param)
	case OP_ORIR:
		err = state.logIR(int(opcode), opCodeBytes[0], param)
	case OP_ANDIR:
		err = state.logIR(int(opcode), opCodeBytes[0], param)
	case OP_XORIR:
		err = state.logIR(int(opcode), opCodeBytes[0], param)
	case OP_LSHIR:
		err = state.logIR(int(opcode), opCodeBytes[0], param)
	case OP_RSHIR:
		err = state.logIR(int(opcode), opCodeBytes[0], param)
	case OP_INCR:
		err = state.incR(param)
	case OP_DECR:
		err = state.decR(param)
	case OP_JMP:
		err = state.jmp(opCodeBytes[0], param)
	case OP_JMPE:
		err = state.jmpE(opCodeBytes[0], param)
	case OP_JMPNE:
		err = state.jmpNE(opCodeBytes[0], param)
	case OP_JMPZ:
		err = state.jmpZ(opCodeBytes[0], param)
	case OP_JMPNZ:
		err = state.jmpNZ(opCodeBytes[0], param)
	case OP_JMPG:
		err = state.jmpG(opCodeBytes[0], param)
	case OP_JMPGE:
		err = state.jmpGE(opCodeBytes[0], param)
	case OP_JMPL:
		err = state.jmpL(opCodeBytes[0], param)
	case OP_JMPLE:
		err = state.jmpLE(opCodeBytes[0], param)
	case OP_JMPIP:
		err = state.jmpIP(opCodeBytes[0], param)
	case OP_JMPEIP:
		err = state.jmpEIP(opCodeBytes[0], param)
	case OP_JMPNEIP:
		err = state.jmpNEIP(opCodeBytes[0], param)
	case OP_JMPZIP:
		err = state.jmpZIP(opCodeBytes[0], param)
	case OP_JMPNZIP:
		err = state.jmpNZIP(opCodeBytes[0], param)
	case OP_JMPGIP:
		err = state.jmpGIP(opCodeBytes[0], param)
	case OP_JMPGEIP:
		err = state.jmpGEIP(opCodeBytes[0], param)
	case OP_JMPLIP:
		err = state.jmpLIP(opCodeBytes[0], param)
	case OP_JMPLEIP:
		err = state.jmpLEIP(opCodeBytes[0], param)
	case OP_PUSHI:
		err = state.push(int(opcode), opCodeBytes[0], param)
	case OP_PUSHR:
		err = state.push(int(opcode), opCodeBytes[0], param)
	case OP_POP:
		err = state.popR(opCodeBytes[0], param)
	case OP_CMPRR:
		err = state.cmpRR(opCodeBytes[0], param)
	case OP_CMPIR:
		err = state.cmpIR(opCodeBytes[0], param)
	case OP_NOP:
	case OP_CLR:
		err = state.clr()
	case OP_CALL:
		err = state.call(param)
	case OP_CALLIP:
		err = state.callIP(opCodeBytes[0], param)
	case OP_RET:
		err = state.ret()
	case OP_EXITI:
		state.exitI(param)
	case OP_EXITR:
		state.exitR(param)
	default:
		err = fmt.Errorf("Unhandled opcode %d, TODO", opcode)
	}
	return err
}

func (state *VmState) exitI(param []byte) {
	state.exitImpl(binary.BigEndian.Uint64(param))
}
func (state *VmState) exitR(param []byte) {
	reg := param[0]
	dataSz := param[1]
	code := state.GetRegVAsU64(int(reg), dataSz)
	state.exitImpl(code)
}
func (state *VmState) exitImpl(code uint64) {
	state.exitCode = code
	state.exit = true
}
