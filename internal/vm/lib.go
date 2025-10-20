package vm

import (
	"encoding/binary"
	"fmt"
	"reflect"
)


type VmState struct {
	regs  Registers
	flags Flags
	stack []any
}

type Registers struct {
	r          [8]any
	ip, sp, bp uint64
}

type Flags struct {
	cf, pf, zf, sf, of bool
}

func CreateVmState(stackSize uint64) VmState {
	state := VmState{
		regs: Registers{
			ip: 0,
			r:  [8]any{uint64(0), uint64(0), uint64(0), uint64(0), uint64(0), uint64(0), uint64(0), uint64(0)},
		},
		flags: Flags{},
		stack: make([]any, stackSize),
	}
	return state
}

func (vm *VmState) Execute(bytecode []byte) error {
	codeSize := len(bytecode) / INSTRUCTION_SIZE
	for vm.regs.ip = 0; vm.regs.ip < uint64(codeSize); vm.regs.ip++ {
		instAddr := vm.regs.ip * INSTRUCTION_SIZE
		opcode := binary.BigEndian.Uint32(bytecode[instAddr : instAddr+OPCODE_SIZE])
		param := binary.BigEndian.Uint64(bytecode[instAddr+OPCODE_SIZE : instAddr+INSTRUCTION_SIZE])
		switch opcode {
		case OP_MOVIR0:
			vm.regs.r[0] = uint64(param)
		case OP_MOVIR1:
			vm.regs.r[1] = uint64(param)
		case OP_ADDSR0R1:
			res, err := addValuesSigned(vm.regs.r[0], vm.regs.r[1])
			if err != nil {
				return err
			}
			vm.regs.r[1] = res
		case OP_JMP:
			vm.regs.ip = param;
		case OP_TESTR0R1:

		}
	}

	fmt.Println(vm.regs.r)

	return nil
}
func addValuesSigned(a, b any) (any, error) {
	aType, bType := reflect.TypeOf(a).Kind(), reflect.TypeOf(b).Kind()
	switch true {
	case aType == reflect.Uint64 && bType == reflect.Uint64:
		aInt, bInt := (a.(uint64)), b.(uint64)
		return int64(aInt) + int64(bInt), nil

	default:
		return nil, fmt.Errorf("Unsupported ADD operation between types (%s) and (%s) ", aType, bType)
	}
}
