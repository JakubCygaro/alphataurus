package vm

import (
	"encoding/binary"
	"fmt"
	"math"
)

type VmStack []uint64

type VmState struct {
	regs  Registers
	flags Flags
	stack VmStack
}

const (
	R0_IDX = 0
	R1_IDX = iota
	R2_IDX
	R3_IDX
	R4_IDX
	R5_IDX
	R6_IDX
	R7_IDX
	BP_IDX
	SP_IDX
	IP_IDX
)
const (
	GP_REG_MAX = R7_IDX
)

const (
	TY_INT64   = iota
	TY_UINT64  = iota
	TY_FLOAT64 = iota
)

type Registers struct {
	r [IP_IDX + 1]uint64
}

func (state *VmState) GetBp() uint64 {
	return state.regs.r[BP_IDX]
}

func (state *VmState) GetSp() uint64 {
	return state.regs.r[SP_IDX]
}

func (state *VmState) GetIp() uint64 {
	return state.regs.r[IP_IDX]
}

func (state *VmState) setIp(v uint64) {
	state.regs.r[IP_IDX] = v
}
func (state *VmState) incIp() {
	state.setIp(state.GetIp() + 1)
}

type Flags struct {
	cf, pf, zf, sf, of bool
}

func CreateVmState(stackSize uint64) VmState {
	state := VmState{
		regs: Registers{
			r: [11]uint64{},
		},
		flags: Flags{},
		stack: make(VmStack, 0, stackSize),
	}
	return state
}

// get X general purpose register value as uint64
func (vm *VmState) GetGpRX(register byte) (uint64, error) {
	if !isGpReg(register) {
		return 0, fmt.Errorf("Disallowed register index %d", register)
	}
	return vm.regs.r[register], nil
}

// get X general purpose register value and cast it into a supported type value
// returned via out
func (vm *VmState) GetGpRXAs(register byte, ty byte, out *any) error {
	val, err := vm.GetGpRX(register)
	if err != nil {
		return err
	}
	switch ty {
	case TY_UINT64:
		*out = val
	case TY_INT64:
		*out = int64(val)
	case TY_FLOAT64:
		*out = float64(math.Float64frombits(val))
	default:
		return fmt.Errorf("Unsupported type for register value conversion")
	}
	return nil
}
func (vm *VmState) GetGpRXAsUint64(register byte) (uint64, error) {
	val, err := vm.GetGpRX(register)
	if err != nil {
		return 0, err
	}
	return val, nil
}
func (vm *VmState) GetGpRXAsInt64(register byte) (int64, error) {
	val, err := vm.GetGpRX(register)
	if err != nil {
		return 0, err
	}
	return int64(val), nil
}
func (vm *VmState) GetGpRXAsFloat64(register byte) (float64, error) {
	val, err := vm.GetGpRX(register)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(val), nil
}
func (vm *VmState) ClearState() {
	vm.regs.r = [11]uint64{}
	vm.flags = Flags{}
	vm.stack = make(VmStack, 0, cap(vm.stack))
}

func (vm *VmState) Execute(bytecode []byte) error {
	codeSize := len(bytecode) / INSTRUCTION_SIZE
	vm.setIp(0)
	// fmt.Println("STARTING REGISTER STATE")
	// fmt.Println(vm.regs.r)
	for ; vm.GetIp() < uint64(codeSize); vm.incIp() {
		var err error = nil
		instAddr := vm.GetIp() * INSTRUCTION_SIZE
		opCodeBytes := bytecode[instAddr : instAddr+OPCODE_SIZE]
		param := bytecode[instAddr+OPCODE_SIZE : instAddr+INSTRUCTION_SIZE]
		opcode, err := GetOpcode(opCodeBytes)
		if err != nil {
			return err
		}
		// fmt.Printf("opcode: %d\n", opcode)
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
		case OP_ADDIR:
			err = vm.arthIR(int(opcode), opCodeBytes[0], param)
		case OP_SUBIR:
			err = vm.arthIR(int(opcode), opCodeBytes[0], param)
		case OP_INCR:
			err = vm.incR(param)
		case OP_DECR:
			err = vm.decR(param)
		case OP_JMP:
			vm.jmp(param)
		case OP_JMPE:
			vm.jmpE(param)
		case OP_JMPG:
			vm.jmpG(param)
		case OP_CMP:
			err = vm.cmp(opCodeBytes[0], param)
		default:
			return fmt.Errorf("Unhandled opcode %d, TODO", opcode)
		}
		if err != nil {
			return err
		}
		// fmt.Printf("ip: %d & REGSTATE\n", vm.GetIp())
		// fmt.Println(vm.flags)
		// fmt.Println(vm.regs.r)
	}

	// fmt.Println("ENDING REGISTER STATE")
	// fmt.Println(vm.regs.r)

	return nil
}
func isGpReg(b byte) bool {
	return b <= GP_REG_MAX
}
func isMovRRAllowed(b byte) bool {
	return b <= SP_IDX
}
func (state *VmState) movRR(lastByte byte) error {
	var src, dest byte
	src |= (lastByte & 0xf0) >> 4
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
	if !isGpReg(byte(reg)) {
		return fmt.Errorf("Bad INCR parameter, disallowed register 0x%x", reg)
	}
	state.regs.r[reg]++
	return nil
}
func (state *VmState) decR(param []byte) error {
	reg := binary.BigEndian.Uint64(param)
	if !isGpReg(byte(reg)) {
		return fmt.Errorf("Bad DECR parameter, disallowed register 0x%x", reg)
	}
	state.regs.r[reg]--
	return nil
}
func arthIRGetParameters(lastByte byte) (reg, ty byte, err error) {
	ty |= (lastByte & 0xf0) >> 4
	reg |= (lastByte & 0x0f)
	if !isGpReg(reg) && reg != BP_IDX && reg != SP_IDX {
		return reg, ty, fmt.Errorf("Bad ADDIR parameter, disallowed target register 0x%x", reg)
	}
	return reg, ty, nil
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
	switch opType {
	case OP_ADDRR:
		addValues(srcV, destV, ty, &state.regs.r[dest])
	case OP_SUBRR:
		subValues(destV, srcV, ty, &state.regs.r[dest])
	case OP_MULRR:
		srcV := state.regs.r[R0_IDX]
		destV := state.regs.r[R1_IDX]
		mulValues(srcV, destV, ty, &state.regs.r[R2_IDX])
	case OP_DIVRR:
		srcV := state.regs.r[R0_IDX]
		destV := state.regs.r[R1_IDX]
		divValues(srcV, destV, ty, &state.regs.r[R2_IDX], &state.regs.r[R3_IDX])
	}
	return nil
}
func (state *VmState) arthIR(opType int, lastByte byte, param []byte) error {
	reg, ty, err := arthIRGetParameters(lastByte)
	if err != nil {
		return err
	}
	immV := binary.BigEndian.Uint64(param)
	regV := state.regs.r[reg]
	switch opType {
	case OP_ADDIR:
		addValues(regV, immV, ty, &state.regs.r[reg])
	case OP_SUBIR:
		subValues(regV, immV, ty, &state.regs.r[reg])
	}
	return nil
}
func (state *VmState) jmp(param []byte) {
	dest := binary.BigEndian.Uint64(param)
	state.setIp(dest)
}
func (state *VmState) jmpE(param []byte) {
	dest := binary.BigEndian.Uint64(param)
	if state.flags.zf {
		state.setIp(dest)
	}
}
func (state *VmState) jmpG(param []byte) {
	dest := binary.BigEndian.Uint64(param)
	if !state.flags.zf && !state.flags.sf {
		state.setIp(dest)
	}
}
func (state *VmState) cmp(lastByte byte, param []byte) error {
	var subtrahend, minuend byte
	var diff int64
	subtrahend |= (lastByte & 0xf0) >> 4
	minuend |= (lastByte & 0x0f)
	if !isGpReg(minuend) {
		return fmt.Errorf("Bad CMP instruction, minuend was not a valid register 0x%x", minuend)
	}
	switch {
	// if subtrahend is not a valid gp reg, then this is an immediate value cmp
	case !isGpReg(subtrahend):
		imm := int64(binary.BigEndian.Uint64(param))
		diff = int64(state.regs.r[minuend]) - imm
	default:
		diff = int64(state.regs.r[minuend]) - int64(state.regs.r[subtrahend])
	}
	state.flags.sf = diff <= 0
	state.flags.zf = diff == 0
	return nil
}
