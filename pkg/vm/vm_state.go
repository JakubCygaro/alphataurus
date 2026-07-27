package vm

import (
	"encoding/binary"
	"fmt"
	"github.com/JakubCygaro/alphataurus/pkg/vm/decls"
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
	currentOpcode OpCodeVal
	exitCode      uint64
	exit          bool
	dbgAd         dbgAdapters
}
type OpCodeTraceFn func(OpCodeVal)
type dbgAdapters struct {
	opcodeTrace OpCodeTraceFn
}
type Flags struct {
	Cf, Pf, Zf, Sf, Of, Df bool
}

func CreateVmState(stackSize uint64) VmState {
	state := VmState{
		regs: Registers{
			r: [11]Register{},
		},
		flags: Flags{},
		stack: make(VmStack, stackSize),
		dbgAd: dbgAdapters{},
	}
	return state
}

func (vm *VmState) SetOpCodeTrace(fn OpCodeTraceFn) {
	vm.dbgAd.opcodeTrace = fn
}
func (vm *VmState) ClearState() {
	vm.regs.r = [11]Register{}
	vm.flags = Flags{}
	vm.stack = make(VmStack, cap(vm.stack))
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

// increment IP so it points to the next instruction
func (state *VmState) incIp() {
	state.setIp(state.GetIp() + decls.INSTRUCTION_SIZE)
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
func (vm *VmState) GetRegistersAsU64() []uint64 {
	ret := make([]uint64, len(vm.regs.r))
	for i := range vm.regs.r {
		ret[i] = vm.GetRegVAsU64(i, SZ_64)
	}
	return ret
}
func (vm *VmState) GetGpRXAsU64(register byte) (uint64, error) {
	var out any
	if err := vm.GetGpRXAs(register, TY_UINT, SZ_64, &out); err != nil {
		return 0, err
	}
	return out.(uint64), nil
}
func (vm *VmState) GetGpRXAsS64(register byte) (int64, error) {
	var out any
	if err := vm.GetGpRXAs(register, TY_SINT, SZ_64, &out); err != nil {
		return 0, err
	}
	return out.(int64), nil
}
func (vm *VmState) GetGpRXAsFloat64(register byte) (float64, error) {
	var out any
	if err := vm.GetGpRXAs(register, TY_FLOAT, SZ_64, &out); err != nil {
		return 0.0, err
	}
	return out.(float64), nil
}

// get X general purpose register value as uint64
func (vm *VmState) GetGpRX(register byte) (Register, error) {
	if !IsGpReg(register) {
		return Register{}, fmt.Errorf("Disallowed register index %d", register)
	}
	return vm.regs.r[register], nil
}
func (vm *VmState) getRXAs(register byte, ty, dataSz byte, out *any) error {
	r := &(vm.regs.r[int(register)])
	return r.GetValAs(ty, dataSz, out)
}

// get X general purpose register value and cast it into a supported type value
// returned via out
func (vm *VmState) GetGpRXAs(register byte, ty, dataSz byte, out *any) error {
	if !IsGpReg(register) {
		return fmt.Errorf("Disallowed register index %d", register)
	}
	return vm.getRXAs(register, ty, dataSz, out)
}
