package vm

import (
	"encoding/binary"
	"fmt"
	"github.com/JakubCygaro/alphataurus/pkg/vm/errors"
	"math"
)

type VmStack []byte

const (
	ADDRESSDEADZONE_SIZE = 0x1000
)

type VmState struct {
	regs     Registers
	flags    Flags
	stack    VmStack
	codeSize uint64
	bytecode []byte
	// this is the virtual address of the stack, it is supposed to start right after the code section
	stackSegBase  int
	byteCodePos   uint64
	exeSegBase    uint64
	currentOpcode uint32
	exitCode      uint64
	exit          bool
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
	TY_SINT  = iota
	TY_UINT  = iota
	TY_FLOAT = iota
)
const (
	SZ_8  = iota
	SZ_16 = iota
	SZ_32
	SZ_64
)

type Register [8]byte

func (r Register) ToDisplayString() string {
	return fmt.Sprintf(` [%vu8 | %vu16 | %vu32 | %vu64]`,
		r[7],
		binary.BigEndian.Uint16(r[6:]),
		binary.BigEndian.Uint32(r[4:]),
		binary.BigEndian.Uint64(r[:]),
	)
}

type Registers struct {
	r [IP_IDX + 1]Register
}

func (state *VmState) GetExitCode() uint64 {
	return state.exitCode
}

func (state *VmState) GetBp() uint64 {
	return binary.BigEndian.Uint64(state.regs.r[BP_IDX][:])
}

func (state *VmState) GetSp() uint64 {
	return binary.BigEndian.Uint64(state.regs.r[SP_IDX][:])
}

// takes the virtual instruction pointer and transforms it into the real position of the
// instruction in the bytecode []byte array
//
// basically subtracts state.exeSegBase from the input value
func (state *VmState) VirtToRealIp(virtual uint64) uint64 {
	return virtual - state.exeSegBase
}
func (state *VmState) RealToVirtIp(r uint64) uint64 {
	return r + state.exeSegBase
}

func (state *VmState) GetRealSp() int {
	return state.VirtToRealSp(int(binary.BigEndian.Uint64(state.regs.r[SP_IDX][:])))
}

func (state *VmState) VirtToRealSp(virtual int) int {
	return virtual - state.stackSegBase
}

func (state *VmState) RealToVirtSp(r int) int {
	return r + state.stackSegBase
}

func (state *VmState) GetIp() uint64 {
	return binary.BigEndian.Uint64(state.regs.r[IP_IDX][:])
}

func (state *VmState) setIp(v uint64) {
	binary.BigEndian.PutUint64(state.regs.r[IP_IDX][:], v)
}
func (state *VmState) incIp() {
	state.setIp(state.GetIp() + INSTRUCTION_SIZE)
}

type Flags struct {
	Cf, Pf, Zf, Sf, Of bool
}

func CreateVmState(stackSize uint64) VmState {
	state := VmState{
		regs: Registers{
			r: [11]Register{},
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
	return binary.BigEndian.Uint64(vm.regs.r[register][:]), nil
}

// get X general purpose register value and cast it into a supported type value
// returned via out
func (vm *VmState) GetGpRXAs(register byte, ty byte, out *any) error {
	val, err := vm.GetGpRX(register)
	if err != nil {
		return err
	}
	switch ty {
	case TY_UINT:
		*out = val
	case TY_SINT:
		*out = int64(val)
	case TY_FLOAT:
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
func (vm *VmState) GetRegisters() []Register {
	regs := make([]Register, len(vm.regs.r))
	copy(regs[:], vm.regs.r[:])
	return regs
}
func (vm *VmState) GetRegistersAsUInt64() []uint64 {
	ret := make([]uint64, len(vm.regs.r))
	for i := range vm.regs.r {
		ret[i] = vm.getRegVAsUint64(i, SZ_64)
	}
	return ret
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
	vm.regs.r = [11]Register{}
	vm.flags = Flags{}
	vm.stack = make(VmStack, cap(vm.stack))
}

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
		case OP_CMP:
			err = vm.cmp(opCodeBytes[0], param)
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
func IsGpReg(b byte) bool {
	return b <= GP_REG_MAX
}
func isMovRRAllowed(b byte) bool {
	return b <= SP_IDX
}
func isPushRAllowed(b byte) bool {
	return IsGpReg(b) || b == SP_IDX || b == BP_IDX
}

func (state *VmState) incrementRegU64(reg int, amount uint64) {
	v := binary.BigEndian.Uint64(state.regs.r[reg][:])
	v += amount
	binary.BigEndian.PutUint64(state.regs.r[reg][:], v)
}
func (state *VmState) decrementRegUS(reg int, amount uint64) {
	v := binary.BigEndian.Uint64(state.regs.r[reg][:])
	v -= amount
	binary.BigEndian.PutUint64(state.regs.r[reg][:], v)
}

func (state *VmState) incR(param []byte) error {
	reg := binary.BigEndian.Uint64(param)
	if !IsGpReg(byte(reg)) {
		return errors.DisallowedOp1Register(int(reg), state.byteCodePos)
	}
	state.incrementRegU64(int(reg), 1)
	return nil
}
func (state *VmState) decR(param []byte) error {
	reg := binary.BigEndian.Uint64(param)
	if !IsGpReg(byte(reg)) {
		return errors.DisallowedOp1Register(int(reg), state.byteCodePos)
	}
	state.decrementRegUS(int(reg), 1)
	return nil
}

type arthIRParamData struct {
	reg, ty byte
	r1sz    byte
}

func (state *VmState) arthIRGetParameters(lastByte byte) (data arthIRParamData, err error) {
	ret := arthIRParamData{}
	ret.reg |= (lastByte & 0b0000_1111)
	ret.r1sz |= (lastByte & 0b0011_0000) >> 4
	ret.ty |= (lastByte & 0b1100_0000) >> 6
	if !IsGpReg(ret.reg) && ret.reg != BP_IDX && ret.reg != SP_IDX {
		return ret, errors.DisallowedDestRegister(int(ret.reg), state.byteCodePos)
	}
	return ret, nil
}

type arthRRParamData struct {
	src, dest, ty byte
	r1sz, r2sz    byte
}

func (state *VmState) arthRRGetParameters(param []byte) (data arthRRParamData, err error) {
	data.src = param[0]
	data.dest = param[1]
	data.ty = (param[3] & 0b00000011)
	data.r1sz = (param[3] & 0b00001100) >> 2
	data.r2sz = (param[3] & 0b00110000) >> 4
	if !IsGpReg(data.src) {
		return data, errors.DisallowedSrcRegister(int(data.src), state.byteCodePos)
	}
	if !IsGpReg(data.dest) {
		return data, errors.DisallowedDestRegister(int(data.dest), state.byteCodePos)
	}
	return data, nil
}
func (state *VmState) logIR(opType int, lastByte byte, param []byte) error {
	first := 0b0000_1111 & lastByte
	dataSize := (0b0011_0000 & lastByte) >> 4
	fVal, sVal :=
		state.getRegVAsUint64(int(first), dataSize),
		binary.BigEndian.Uint64(param)
	switch opType {
	case OP_ORIR:
		fVal = fVal | sVal
	case OP_ANDIR:
		fVal = fVal & sVal
	case OP_XORIR:
		fVal = fVal ^ sVal
	case OP_LSHIR:
		fVal = fVal << sVal
	case OP_RSHIR:
		fVal = fVal >> sVal
	}
	state.putValInRegWithSize(int(first), dataSize, fVal)
	return nil
}
func (state *VmState) logRR(opType int, param []byte) error {
	data, err := state.arthRRGetParameters(param)
	if err != nil {
		return err
	}
	if data.r1sz < data.r2sz {
		return errors.BadOperandSizes(data.r1sz, data.r2sz, state.byteCodePos)
	}
	fVal, sVal :=
		state.getRegVAsUint64(int(data.src), data.r1sz),
		state.getRegVAsUint64(int(data.dest), data.r2sz)
	switch opType {
	case OP_ORRR:
		fVal = fVal | sVal
	case OP_ANDRR:
		fVal = fVal & sVal
	case OP_XORRR:
		fVal = fVal ^ sVal
	case OP_LSHRR:
		fVal = fVal << sVal
	case OP_RSHRR:
		fVal = fVal >> sVal
	}
	state.putValInRegWithSize(int(data.src), data.r1sz, fVal)
	return nil
}
func (state *VmState) not(param []byte) error {
	reg := binary.BigEndian.Uint64(param)
	if !IsGpReg(byte(reg)) {
		return errors.DisallowedOp1Register(int(reg), state.byteCodePos)
	}
	regV := binary.BigEndian.Uint64(state.regs.r[reg][:])

	binary.BigEndian.PutUint64(state.regs.r[reg][:], ^regV)
	return nil
}
func (state *VmState) arthRR(opType int, param []byte) error {
	data, err := state.arthRRGetParameters(param)
	if err != nil {
		return err
	}
	srcV := binary.BigEndian.Uint64(state.regs.r[data.src][:])
	destV := binary.BigEndian.Uint64(state.regs.r[data.dest][:])
	switch opType {
	case OP_ADDRR:
		err = state.addValues(srcV, destV,
			data.ty, data.r2sz,
			state.regs.r[data.dest][:])
	case OP_SUBRR:
		err = state.subValues(destV, srcV,
			data.ty, data.r2sz,
			state.regs.r[data.dest][:])
	case OP_MULRR:
		srcV := binary.BigEndian.Uint64(state.regs.r[R0_IDX][:])
		destV := binary.BigEndian.Uint64(state.regs.r[R1_IDX][:])
		err = state.mulValues(srcV, destV,
			data.ty, data.r2sz,
			state.regs.r[R2_IDX][:])
	case OP_DIVRR:
		srcV := binary.BigEndian.Uint64(state.regs.r[R0_IDX][:])
		destV := binary.BigEndian.Uint64(state.regs.r[R1_IDX][:])
		err = state.divValues(srcV, destV,
			data.ty, data.r2sz,
			state.regs.r[R2_IDX][:], state.regs.r[R3_IDX][:])
	}
	return err
}
func (state *VmState) arthIR(opType int, lastByte byte, param []byte) error {
	data, err := state.arthIRGetParameters(lastByte)
	if err != nil {
		return err
	}
	immV := binary.BigEndian.Uint64(param)
	regV := state.getRegVAsUint64(int(data.reg), data.r1sz)
	switch opType {
	case OP_ADDIR:
		err = state.addValues(regV, immV,
			data.ty, data.r1sz,
			state.regs.r[data.reg][:])
	case OP_SUBIR:
		err = state.subValues(regV, immV,
			data.ty, data.r1sz,
			state.regs.r[data.reg][:])
	}
	return nil
}
func (state *VmState) clr() error {
	state.flags = Flags{}
	return nil
}
func (state *VmState) cmp(lastByte byte, param []byte) error {
	var subtrahend, minuend, dataSz byte
	subtrahend |= (lastByte & 0b11110000) >> 4
	minuend |= (lastByte & 0b00000011)
	dataSz |= (lastByte & 0b00001100) >> 2
	if !IsGpReg(minuend) {
		return errors.DisallowedOp1Register(int(minuend), state.byteCodePos)
	}
	var subV, minV uint64
	minV = binary.BigEndian.Uint64(state.regs.r[minuend][:])

	var ty byte
	if !IsGpReg(subtrahend) {
		// in this case the subtrahend is an immediate value
		// and the type of the operation is determined by
		// subtrahend - GP_REG_MAX - 1
		ty = subtrahend - GP_REG_MAX - 1
		subV = binary.BigEndian.Uint64(param)
	} else {
		ty = param[0]
		subV = binary.BigEndian.Uint64(state.regs.r[subtrahend][:])
	}
	if dataSz != SZ_64 && ty == TY_FLOAT {
		return errors.BadArthmeticOperation(state.byteCodePos)
	}
	diff := [8]byte{}
	err := state.subValues(minV, subV,
		ty, dataSz,
		diff[:])
	state.flags = Flags{}

	if ty == TY_FLOAT {
		f := math.Float64frombits(binary.BigEndian.Uint64(diff[:]))
		state.flags.Sf = f < 0.0
		state.flags.Zf = f == 0.0
	} else {
		i := binary.BigEndian.Uint64(diff[:])
		switch dataSz {
		case SZ_8:
			state.flags.Sf = int8(i) < 0
			state.flags.Zf = int8(i) == 0
		case SZ_16:
			state.flags.Sf = int16(i) < 0
			state.flags.Zf = int16(i) == 0
		case SZ_32:
			state.flags.Sf = int32(i) < 0
			state.flags.Zf = int32(i) == 0
		case SZ_64:
			state.flags.Sf = int64(i) < 0
			state.flags.Zf = int64(i) == 0
		}
	}

	return err
}
func (state *VmState) pushImpl(data []byte) error {
	inc := len(data)
	if state.GetRealSp()+inc >= len(state.stack) {
		return errors.StackOverflow(state.byteCodePos)
	}
	state.incrementRegU64(SP_IDX, uint64(inc))
	copy(
		state.stack[state.GetRealSp()-inc+1:state.GetRealSp()+1],
		data[:],
	)
	return nil
}
func dataSizeToByteCount(dataSz byte) int {
	switch dataSz {
	case SZ_8:
		return 1
	case SZ_16:
		return 2
	case SZ_32:
		return 4
	case SZ_64:
		return 8
	}
	return -1
}
func (state *VmState) push(opTy int, lastByte byte, param []byte) error {
	var val []byte
	// 2 bits for the data size
	dataSz := (0b00000011 & lastByte)
	byteSpan := dataSizeToByteCount(dataSz)
	switch opTy {
	case OP_PUSHI:
		val = param[8-byteSpan:]
	case OP_PUSHR:
		reg := binary.BigEndian.Uint64(param)
		if isPushRAllowed(byte(reg)) {
			val = state.regs.r[reg][8-byteSpan:]
		} else {
			return errors.DisallowedOp1Register(int(reg), state.byteCodePos)
		}
	}
	return state.pushImpl(val)
}
func (state *VmState) popImpl(bytes int) ([]byte, error) {
	if state.GetRealSp() < 0 {
		return nil, errors.StackUnderflow(state.byteCodePos)
	}
	val := state.stack[state.GetRealSp()-bytes+1:state.GetRealSp()+1]
	var sp = binary.BigEndian.Uint64(state.regs.r[SP_IDX][:])
	sp -= uint64(bytes)
	binary.BigEndian.PutUint64(state.regs.r[SP_IDX][:], sp)
	return val, nil
}
func (state *VmState) popR(lastByte byte, param []byte) error {
	dataSz := lastByte & 0b00000011
	val, err := state.popImpl(
		dataSizeToByteCount(dataSz),
	)
	if err != nil {
		return err
	}
	reg := binary.BigEndian.Uint64(param)
	if isPushRAllowed(byte(reg)) {
		copy(
			state.regs.r[reg][8-len(val):],
			val[:],
		)
	}
	return nil
}

func (state *VmState) getRegVAsUint64(reg int, dataSz byte) uint64 {
	bytes := dataSizeToByteCount(dataSz)
	ret := uint64(0)
	r := state.regs.r[reg][:]
	switch dataSz {
	case SZ_8:
		ret = uint64(r[7])
	case SZ_16:
		ret = uint64(binary.BigEndian.Uint16(r[8-bytes:]))
	case SZ_32:
		ret = uint64(binary.BigEndian.Uint32(r[8-bytes:]))
	case SZ_64:
		ret = uint64(binary.BigEndian.Uint64(r[8-bytes:]))
	}

	return ret
}
func (state *VmState) putValInStackWithSize(dataSz byte, val uint64, address int) error {
	if address+int(dataSz) > len(state.stack) {
		return errors.StackOverflow(state.byteCodePos)
	}
	bytes := dataSizeToByteCount(dataSz)
	s := state.stack[address:bytes]
	switch dataSz {
	case SZ_8:
		s[0] = byte(val)
	case SZ_16:
		binary.BigEndian.PutUint16(s[8-bytes:], uint16(val))
	case SZ_32:
		binary.BigEndian.PutUint32(s[8-bytes:], uint32(val))
	case SZ_64:
		binary.BigEndian.PutUint64(s[8-bytes:], uint64(val))
	}
	return nil
}
func (state *VmState) putValInRegWithSize(reg int, dataSz byte, val uint64) {
	bytes := dataSizeToByteCount(dataSz)
	r := state.regs.r[reg][:]
	switch dataSz {
	case SZ_8:
		r[7] = byte(val)
	case SZ_16:
		binary.BigEndian.PutUint16(r[8-bytes:], uint16(val))
	case SZ_32:
		binary.BigEndian.PutUint32(r[8-bytes:], uint32(val))
	case SZ_64:
		binary.BigEndian.PutUint64(r[8-bytes:], uint64(val))
	}
}
