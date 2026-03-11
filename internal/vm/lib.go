package vm

import (
	"encoding/binary"
	"fmt"
	"github.com/JakubCygaro/alphataurus/internal/vm/errors"
	"math"
)

type VmStack []uint64

const (
	ADDRESSDEADZONE_SIZE = 0xff
)

type VmState struct {
	regs  Registers
	flags Flags
	stack VmStack
	// this is the virtual address of the stack, it is supposed to start right after the code section
	stackSegBase     int
	byteCodePos   uint64
	exeSegBase   uint64
	currentOpcode uint32
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
func (state *VmState) VirtToRealIp(virtual uint64) uint64{
	return virtual - state.exeSegBase
}
func (state *VmState) RealToVirtIp(r uint64) uint64{
	return r + state.exeSegBase
}

func (state *VmState) GetRealSp() int {
	return state.VirtToRealSp(int(state.regs.r[SP_IDX]))
}

func (state *VmState) VirtToRealSp(virtual int) int {
	return virtual - state.stackSegBase
}

func (state *VmState) RealToVirtSp(r int) int {
	return r + state.stackSegBase
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
	Cf, Pf, Zf, Sf, Of bool
}

func CreateVmState(stackSize uint64) VmState {
	state := VmState{
		regs: Registers{
			r: [11]uint64{},
		},
		flags: Flags{},
		stack: make(VmStack, stackSize),
	}
	return state
}

// get X general purpose register value as uint64
func (vm *VmState) GetGpRX(register byte) (uint64, error) {
	if !IsGpReg(register) {
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
func (vm *VmState) GetFlags() Flags {
	return vm.flags
}
func (vm *VmState) GetStack() VmStack {
	ret := make(VmStack, len(vm.stack))
	copy(ret[:], vm.stack[:])
	return ret
}
func (vm *VmState) GetRegisters() []uint64 {
	regs := make([]uint64, len(vm.regs.r))
	copy(regs[:], vm.regs.r[:])
	return regs
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
	vm.stack = make(VmStack, cap(vm.stack))
}

func (vm *VmState) Execute(bytecode []byte) error {
	codeSize := len(bytecode) / INSTRUCTION_SIZE
	// the code section starts after the deadzone (for now)
	vm.exeSegBase = ADDRESSDEADZONE_SIZE
	// the stack starts after the code section
	vm.stackSegBase =  int(vm.exeSegBase) + codeSize
	// the base pointer points right before the beggining of the stack section
	vm.regs.r[BP_IDX] = uint64(vm.stackSegBase) - 1
	// the stack pointer points to the base pointer
	vm.regs.r[SP_IDX] = vm.regs.r[BP_IDX]
	vm.setIp(vm.exeSegBase)
	for ; vm.GetIp()-ADDRESSDEADZONE_SIZE < uint64(codeSize); vm.incIp() {
		var err error = nil
		instAddr := (vm.VirtToRealIp(vm.GetIp())) * INSTRUCTION_SIZE
		vm.byteCodePos = vm.GetIp()
		opCodeBytes := bytecode[instAddr : instAddr+OPCODE_SIZE]
		param := bytecode[instAddr+OPCODE_SIZE : instAddr+INSTRUCTION_SIZE]
		opcode, err := vm.GetOpcode(opCodeBytes)
		vm.currentOpcode = uint32(opcode)
		if err != nil {
			return err
		}
		switch opcode {
		case OP_MOVRR:
			err = vm.movRR(opCodeBytes[0])
		case OP_MOVIR:
			err = vm.movIR(opCodeBytes[0], param)
		case OP_MOVDRI:
			err = vm.movDRI(opCodeBytes[0], param)
		case OP_MOVDRO1:
			err = vm.movDRO1(opCodeBytes[1], opCodeBytes[0], param)
		case OP_MOVDRO2:
			err = vm.movDRO2(opCodeBytes[1], opCodeBytes[0], param)
		case OP_MOVID:
			err = vm.movID(param)
		case OP_MOVRD:
			err = vm.movRD(param)
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
		case OP_PUSHI:
			err = vm.push(int(opcode), param)
		case OP_PUSHR:
			err = vm.push(int(opcode), param)
		case OP_POP:
			err = vm.popR(param)
		case OP_CMP:
			err = vm.cmp(opCodeBytes[0], param)
		case OP_NOP:
		case OP_CLR:
			err = vm.clr()
		default:
			return fmt.Errorf("Unhandled opcode %d, TODO", opcode)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
func IsGpReg(b byte) bool {
	return b <= GP_REG_MAX
}
func isMovRRAllowed(b byte) bool {
	return b <= SP_IDX
}
func isPushRAllowed(b byte) bool {
	return IsGpReg(b) || b == SP_IDX || b == BP_IDX
}
func (state *VmState) incR(param []byte) error {
	reg := binary.BigEndian.Uint64(param)
	if !IsGpReg(byte(reg)) {
		return errors.DisallowedOp1Register(int(reg), state.byteCodePos)
	}
	state.regs.r[reg]++
	return nil
}
func (state *VmState) decR(param []byte) error {
	reg := binary.BigEndian.Uint64(param)
	if !IsGpReg(byte(reg)) {
		return errors.DisallowedOp1Register(int(reg), state.byteCodePos)
	}
	state.regs.r[reg]--
	return nil
}
func (state *VmState) arthIRGetParameters(lastByte byte) (reg, ty byte, err error) {
	ty |= (lastByte & 0xf0) >> 4
	reg |= (lastByte & 0x0f)
	if !IsGpReg(reg) && reg != BP_IDX && reg != SP_IDX {
		return reg, ty, errors.DisallowedDestRegister(int(reg), state.byteCodePos)
	}
	return reg, ty, nil
}
func (state *VmState) arthRRGetParameters(param []byte) (src, dest, ty byte, err error) {
	src = param[0]
	dest = param[1]
	ty = param[3]
	if !IsGpReg(src) {
		return 0, 0, 0, errors.DisallowedSrcRegister(int(src), state.byteCodePos)
	}
	if !IsGpReg(dest) {
		return 0, 0, 0, errors.DisallowedDestRegister(int(dest), state.byteCodePos)
	}
	return src, dest, ty, nil
}
func (state *VmState) arthRR(opType int, param []byte) error {
	src, dest, ty, err := state.arthRRGetParameters(param)
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
	reg, ty, err := state.arthIRGetParameters(lastByte)
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
func (state *VmState) clr() error {
	state.flags = Flags{}
	return nil
}
func (state *VmState) cmp(lastByte byte, param []byte) error {
	var subtrahend, minuend byte
	subtrahend |= (lastByte & 0xf0) >> 4
	minuend |= (lastByte & 0x0f)
	if !IsGpReg(minuend) {
		return errors.DisallowedOp1Register(int(minuend), state.byteCodePos)
	}
	var subV, minV uint64
	minV = state.regs.r[minuend]

	var ty byte
	if !IsGpReg(subtrahend) {
		// in this case the subtrahend is an immediate value
		// and the type of the operation is determined by
		// subtrahend - GP_REG_MAX
		ty = subtrahend - GP_REG_MAX
		subV = binary.BigEndian.Uint64(param)
	} else {
		ty = param[0]
		subV = state.regs.r[subtrahend]
	}
	var diff uint64
	subValues(minV, subV, ty, &diff)
	state.flags = Flags{}

	if ty == TY_FLOAT64 {
		state.flags.Sf = math.Float64frombits(diff) <= 0.0
		state.flags.Zf = math.Float64frombits(diff) == 0.0
	} else {
		state.flags.Sf = int64(diff) <= 0
		state.flags.Zf = int64(diff) == 0
	}

	return nil
}
func (state *VmState) push(ty int, param []byte) error {
	var val uint64
	switch ty {
	case OP_PUSHI:
		val = binary.BigEndian.Uint64(param)
	case OP_PUSHR:
		reg := binary.BigEndian.Uint64(param)
		if isPushRAllowed(byte(reg)) {
			val = state.regs.r[reg]
		} else {
			return errors.DisallowedOp1Register(int(reg), state.byteCodePos)
		}
	}
	if state.GetRealSp() >= len(state.stack) {
		return errors.StackOverflow(state.byteCodePos)
	}
	state.stack[state.GetRealSp()+1] = val
	state.regs.r[SP_IDX]++
	return nil
}
func (state *VmState) popR(param []byte) error {
	if state.GetRealSp() < 0 {
		return errors.StackUnderflow(state.byteCodePos)
	}
	val := state.stack[state.GetRealSp()]
	state.regs.r[SP_IDX]--
	reg := binary.BigEndian.Uint64(param)
	if isPushRAllowed(byte(reg)) {
		state.regs.r[reg] = val
	}
	return nil
}
