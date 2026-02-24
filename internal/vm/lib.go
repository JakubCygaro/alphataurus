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
	r0_IDX = 0
	r1_IDX = iota
	r2_IDX
	r3_IDX
	r4_IDX
	r5_IDX
	r6_IDX
	r7_IDX
	bp_IDX
	sp_IDX
	ip_IDX
)

const (
	TY_INT64   = iota
	TY_UINT64  = iota
	TY_FLOAT64 = iota
)

type Registers struct {
	r [ip_IDX + 1]uint64
}

func (state *VmState) getIp() uint64 {
	return state.regs.r[ip_IDX]
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
			r: [11]uint64{uint64(0), uint64(0), uint64(0), uint64(0), uint64(0), uint64(0), uint64(0), uint64(0), uint64(0), uint64(0), uint64(0)},
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
			err = vm.arthRR(int(opcode), param)
		case OP_SUBRR:
			err = vm.arthRR(int(opcode), param)
		case OP_MULRR:
			err = vm.arthRR(int(opcode), param)
		case OP_DIVRR:
			err = vm.arthRR(int(opcode), param)
		case OP_INCR:
			err = vm.incR(param)
		case OP_DECR:
			err = vm.decR(param)
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
	return b < 8
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
	ty |= (lastByte & 0xf0) >> 4
	dest |= (lastByte & 0x0f)
	if !isGpReg(dest) {
		return fmt.Errorf("Bad MOVIR opcode, disallowed destination register %04b", dest)
	}
	switch ty {
	case TY_INT64:
		i64 := binary.BigEndian.Uint64(param)
		state.regs.r[dest] = i64
	case TY_UINT64:
		u64 := binary.BigEndian.Uint64(param)
		state.regs.r[dest] = u64
	case TY_FLOAT64:
		bits := binary.BigEndian.Uint64(param)
		state.regs.r[dest] = bits
	default:
		return fmt.Errorf("Bad MOVIR opcode, unknown type 0x%x", ty)
	}
	return nil
}
func (state *VmState) incR(param []byte) error {
	reg := binary.BigEndian.Uint64(param)
	if !isGpReg(byte(reg)){
		return fmt.Errorf("Bad INCR parameter, disallowed register 0x%x", reg)
	}
	state.regs.r[reg]++
	return nil
}
func (state *VmState) decR(param []byte) error {
	reg := binary.BigEndian.Uint64(param)
	if !isGpReg(byte(reg)){
		return fmt.Errorf("Bad DECR parameter, disallowed register 0x%x", reg)
	}
	state.regs.r[reg]--
	return nil
}
func arthRRGetParameters(param []byte) (src, dest, ty byte, err error) {
	src = param[0]
	dest = param[1]
	ty = param[3]
	if !isGpReg(src) {
		return 0, 0, 0, fmt.Errorf("Bad opcode, disallowed source register %04b", dest)
	}
	if !isGpReg(dest) {
		return 0, 0, 0, fmt.Errorf("Bad opcode, disallowed destination register %04b", dest)
	}
	return src, dest, ty, nil
}
func (state *VmState) arthRR(opType int, param []byte) error {
	src, dest, ty, err := arthRRGetParameters(param)
	if err != nil {
		return err
	}
	srcV := state.regs.r[src]
	destV := state.regs.r[dest]
	var out uint64
	switch opType {
	case OP_ADDRR:
		addValues(srcV, destV, ty, &out)
	case OP_SUBRR:
		subValues(srcV, destV, ty, &out)
	case OP_MULRR:
		srcV := state.regs.r[r0_IDX]
		destV := state.regs.r[r1_IDX]
		dest = r2_IDX
		mulValues(srcV, destV, ty, &out)
	case OP_DIVRR:
		srcV := state.regs.r[r0_IDX]
		destV := state.regs.r[r1_IDX]
		dest = r2_IDX
		divValues(srcV, destV, ty, &out, &state.regs.r[r3_IDX])
	}
	state.regs.r[dest] = out
	return nil
}
func (state *VmState) jmp(param []byte) {
	dest := binary.BigEndian.Uint64(param)
	state.setIp(dest)
}
