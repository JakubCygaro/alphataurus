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
		uint64(vm.stackSegBase) - 1,
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
		var err error = nil
		instAddr := vm.VirtToRealIp(vm.GetIp())
		vm.byteCodePos = vm.GetIp()
		opCodeBytes := vm.bytecode[instAddr : instAddr+OPCODE_SIZE]
		param := vm.bytecode[instAddr+OPCODE_SIZE : instAddr+INSTRUCTION_SIZE]
		opcode, err := vm.GetOpcode(opCodeBytes)
		vm.currentOpcode = uint32(opcode)
		if err != nil {
			return err
		}
		switch opcode {
		case OP_MOVRR:
			err = vm.movRR(opCodeBytes[0], param)
		case OP_MOVIR:
			err = vm.movIR(opCodeBytes[0], param)
		case OP_MOVDRI:
			err = vm.movDRI(opCodeBytes[0], param)
		case OP_MOVDRO1:
			err = vm.movDRO1(opCodeBytes[1], opCodeBytes[0], param)
		case OP_MOVDRO2:
			err = vm.movDRO2(opCodeBytes[1], opCodeBytes[0], param)
		case OP_MOVID:
			err = vm.movID(opCodeBytes[1], param)
		case OP_MOVRD:
			err = vm.movRD(opCodeBytes[1], param)
		case OP_MOVIDO1:
			err = vm.movIDO1(opCodeBytes[1], opCodeBytes[0], param)
		case OP_MOVRDO1:
			err = vm.movRDO1(opCodeBytes[1], opCodeBytes[0], param)
		case OP_MOVIDO2:
			err = vm.movIDO2(opCodeBytes[1], opCodeBytes[0], param)
		case OP_MOVRDO2:
			err = vm.movRDO2(opCodeBytes[1], opCodeBytes[0], param)
		case OP_ADDRR:
			err = vm.arthRR(int(opcode), param)
		case OP_SUBRR:
			err = vm.arthRR(int(opcode), param)
		case OP_MULRR:
			err = vm.arthRR(int(opcode), param)
		case OP_DIVRR:
			err = vm.arthRR(int(opcode), param)
		case OP_ADDIR:
			err = vm.arthIR(int(opcode), opCodeBytes[0], param)
		case OP_SUBIR:
			err = vm.arthIR(int(opcode), opCodeBytes[0], param)
		case OP_NOT:
			err = vm.not(param)
		case OP_ORRR:
			err = vm.logRR(int(opcode), param)
		case OP_ANDRR:
			err = vm.logRR(int(opcode), param)
		case OP_XORRR:
			err = vm.logRR(int(opcode), param)
		case OP_LSHRR:
			err = vm.logRR(int(opcode), param)
		case OP_RSHRR:
			err = vm.logRR(int(opcode), param)
		case OP_ORIR:
			err = vm.logIR(int(opcode), opCodeBytes[0], param)
		case OP_ANDIR:
			err = vm.logIR(int(opcode), opCodeBytes[0], param)
		case OP_XORIR:
			err = vm.logIR(int(opcode), opCodeBytes[0], param)
		case OP_LSHIR:
			err = vm.logIR(int(opcode), opCodeBytes[0], param)
		case OP_RSHIR:
			err = vm.logIR(int(opcode), opCodeBytes[0], param)
		case OP_INCR:
			err = vm.incR(param)
		case OP_DECR:
			err = vm.decR(param)
		case OP_JMP:
			err = vm.jmp(opCodeBytes[0], param)
		case OP_JMPE:
			err = vm.jmpE(opCodeBytes[0], param)
		case OP_JMPNE:
			err = vm.jmpNE(opCodeBytes[0], param)
		case OP_JMPZ:
			err = vm.jmpZ(opCodeBytes[0], param)
		case OP_JMPNZ:
			err = vm.jmpNZ(opCodeBytes[0], param)
		case OP_JMPG:
			err = vm.jmpG(opCodeBytes[0], param)
		case OP_JMPGE:
			err = vm.jmpGE(opCodeBytes[0], param)
		case OP_JMPL:
			err = vm.jmpL(opCodeBytes[0], param)
		case OP_JMPLE:
			err = vm.jmpLE(opCodeBytes[0], param)
		case OP_JMPIP:
			err = vm.jmpIP(opCodeBytes[0], param)
		case OP_JMPEIP:
			err = vm.jmpEIP(opCodeBytes[0], param)
		case OP_JMPNEIP:
			err = vm.jmpNEIP(opCodeBytes[0], param)
		case OP_JMPZIP:
			err = vm.jmpZIP(opCodeBytes[0], param)
		case OP_JMPNZIP:
			err = vm.jmpNZIP(opCodeBytes[0], param)
		case OP_JMPGIP:
			err = vm.jmpGIP(opCodeBytes[0], param)
		case OP_JMPGEIP:
			err = vm.jmpGEIP(opCodeBytes[0], param)
		case OP_JMPLIP:
			err = vm.jmpLIP(opCodeBytes[0], param)
		case OP_JMPLEIP:
			err = vm.jmpLEIP(opCodeBytes[0], param)
		case OP_PUSHI:
			err = vm.push(int(opcode), opCodeBytes[0], param)
		case OP_PUSHR:
			err = vm.push(int(opcode), opCodeBytes[0], param)
		case OP_POP:
			err = vm.popR(opCodeBytes[0], param)
		case OP_CMPRR:
			err = vm.cmpRR(opCodeBytes[0], param)
		case OP_CMPIR:
			err = vm.cmpIR(opCodeBytes[0], param)
		case OP_NOP:
		case OP_CLR:
			err = vm.clr()
		case OP_CALL:
			err = vm.call(param)
		case OP_CALLIP:
			err = vm.callIP(opCodeBytes[0], param)
		case OP_RET:
			err = vm.ret()
		case OP_EXIT:
			vm.exitCode = binary.BigEndian.Uint64(param)
			vm.exit = true
		default:
			return fmt.Errorf("Unhandled opcode %d, TODO", opcode)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
