package vm

import (
	"encoding/binary"
	"fmt"
)

type VmState struct {
	regs  Registers
	flags Flags
	stack []any
}

const (
	r1_IDX = 0
	r2_IDX = iota
	r3_IDX = iota
	r4_IDX = iota
	r5_IDX = iota
	r6_IDX = iota
	r7_IDX = iota
	r8_IDX = iota
	bp_IDX = iota
	sp_IDX = iota
	ip_IDX = iota
)

const (
	TY_INT64 = iota
	TY_UINT64 = iota
)

type Registers struct {
	r [ip_IDX + 1]any
	// ip, sp, bp uint64
	// ip uint64
}

func (state *VmState) getIp() uint64 {
	return (state.regs.r[ip_IDX]).(uint64)
}

func (state *VmState) setIp(v uint64) {
	state.regs.r[ip_IDX] = v
}
func (state *VmState) incIp() {
	state.setIp(state.getIp() + 1)
}

type Flags struct {
	cf, pf, zf, sf, of bool
}

func CreateVmState(stackSize uint64) VmState {
	state := VmState{
		regs: Registers{
			r: [11]any{uint64(0), uint64(0), uint64(0), uint64(0), uint64(0), uint64(0), uint64(0), uint64(0), uint64(0), uint64(0), uint64(0)},
		},
		flags: Flags{},
		stack: make([]any, stackSize),
	}
	return state
}

func (vm *VmState) Execute(bytecode []byte) error {
	codeSize := len(bytecode) / INSTRUCTION_SIZE
	fmt.Printf("codeSize: %d\n", codeSize)
	vm.setIp(0)
	fmt.Println("STARTING REGISTER STATE")
	fmt.Println(vm.regs.r)
	for ; vm.getIp() < uint64(codeSize); vm.incIp() {
		var err error = nil
		instAddr := vm.getIp() * INSTRUCTION_SIZE
		opCodeBytes := bytecode[instAddr : instAddr+OPCODE_SIZE]
		param := bytecode[instAddr+OPCODE_SIZE : instAddr+INSTRUCTION_SIZE]
		opcode, err := GetOpcode(opCodeBytes)
		if err != nil {
			return err
		}
		fmt.Printf("opcode: %d\n", opcode)
		switch opcode {
		case OP_MOVRR:
			err = vm.movRR(opCodeBytes[0])
		case OP_MOVIR:
			err = vm.movIR(opCodeBytes[0], param)
		case OP_ADDRR:
			err = vm.addRR(param)
		case OP_JMP:
			vm.jmp(param)
		default:
			return fmt.Errorf("Unhandled opcode %d, TODO", opcode)
		}
		if err != nil {
			return err
		}
		fmt.Printf("ip: %d & REGSTATE\n", vm.getIp())
		fmt.Println(vm.regs.r)
	}

	fmt.Println("ENDING REGISTER STATE")
	fmt.Println(vm.regs.r)

	return nil
}
func isGpReg(b byte) bool {
	return b >= 0 && b < 8
}
func isMovRRAllowed(b byte) bool {
	return b <= sp_IDX
}
func (state *VmState) movRR(lastByte byte) error {
	var src, dest byte
	src |= (lastByte & 0xf0)
	dest |= (lastByte & 0x0f)
	if !isMovRRAllowed(src) {
		return fmt.Errorf("Bad MOVRR opcode, disallowed source register")
	} else if !isMovRRAllowed(dest) {
		return fmt.Errorf("Bad MOVRR opcode, disallowed destination register")
	} else {
		state.regs.r[dest] = state.regs.r[src]
	}
	return nil
}
func (state *VmState) movIR(lastByte byte, param []byte) error {
	var ty, dest byte
	// type of value
	ty |= (lastByte & 0xf0)
	dest |= (lastByte & 0x0f)
	if !isGpReg(dest) {
		return fmt.Errorf("Bad MOVIR opcode, disallowed destination register %04b", dest)
	}
	switch ty {
	case TY_INT64:
		i64 := int64(binary.BigEndian.Uint64(param))
		state.regs.r[dest] = i64
	case TY_UINT64:
		u64 := binary.BigEndian.Uint64(param)
		state.regs.r[dest] = u64
	default:
		return fmt.Errorf("Bad MOVIR opcode, unknown type %04b", ty)
	}
	return nil
}
func addValuesSigned(a, b any) (int64, error) {
	aInt, ok := a.(int64)
	if !ok {
		return 0, fmt.Errorf("First operand to signed addition was not an integer")
	}
	bInt, ok := b.(int64)
	if !ok {
		return 0, fmt.Errorf("Second operand to signed addition was not an integer")
	}
	return aInt + bInt, nil
}
func (state *VmState) addRR(param []byte) error {
	var src, dest, sign byte
	src = param[0]
	dest = param[1]
	sign = param[2]
	if !isGpReg(src) {
		return fmt.Errorf("Bad ADDRR opcode, disallowed source register %04b", dest)
	}
	if !isGpReg(dest) {
		return fmt.Errorf("Bad ADDRR opcode, disallowed destination register %04b", dest)
	}
	srcV := state.regs.r[src]
	destV := state.regs.r[dest]
	var out any
	var err error
	if sign != 0 {
		out, err = addValuesSigned(srcV, destV)
	} else {
		return fmt.Errorf("Unsigned ADDRR TODO")
	}
	if err != nil {
		return err
	}
	state.regs.r[dest] = out
	return nil
}
func (state *VmState) jmp(param []byte){
	dest := binary.BigEndian.Uint64(param)
	state.setIp(dest)
}
