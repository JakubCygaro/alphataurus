package vm

import (
	"encoding/binary"

	"github.com/JakubCygaro/alphataurus/internal/vm/errors"
)

func (state *VmState) movRR(lastByte byte) error {
	var src, dest byte
	src |= (lastByte & 0xf0) >> 4
	dest |= (lastByte & 0x0f)
	if !isMovRRAllowed(src) {
		return errors.DisallowedSrcRegister(int(src), state.byteCodePos)
	} else if !isMovRRAllowed(dest) {
		return errors.DisallowedDestRegister(int(dest), state.byteCodePos)
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
	if !IsGpReg(dest) {
		return errors.DisallowedDestRegister(int(dest), state.byteCodePos)
	}
	switch ty {
	case TY_INT64:
		i64 := binary.BigEndian.Uint64(param)
		state.regs.r[dest] = i64
	case TY_FLOAT64:
		bits := binary.BigEndian.Uint64(param)
		state.regs.r[dest] = bits
	default:
		u64 := binary.BigEndian.Uint64(param)
		state.regs.r[dest] = u64
	}
	return nil
}
func (state *VmState) movDRI(lastByte byte, param []byte) error {
	dest := lastByte
	if !isMovRRAllowed(dest) {
		return errors.DisallowedDestRegister(int(dest), state.byteCodePos)
	}
	addr := binary.BigEndian.Uint64(param)
	inStack := state.VirtToRealSp(int(addr))
	if inStack < 0 || inStack >= len(state.stack) {
		return errors.SegmentationFault(addr, state.byteCodePos)
	}
	state.regs.r[dest] = state.stack[inStack]
	return nil
}
func (state *VmState) movDRO1(lastByte byte, param []byte) error {
	dest := (lastByte & 0xf0) >> 4
	reg := (lastByte & 0x0f)
	if !isMovRRAllowed(dest) {
		return errors.DisallowedDestRegister(int(dest), state.byteCodePos)
	}
	if !isMovRRAllowed(reg) {
		return errors.DisallowedOp2Register(int(reg), state.byteCodePos)
	}
	regV := int64(state.regs.r[reg])
	addr := regV + int64(binary.BigEndian.Uint64(param))
	inStack := state.VirtToRealSp(int(addr))
	if inStack < 0 || inStack >= len(state.stack) {
		return errors.SegmentationFault(uint64(addr), state.byteCodePos)
	}
	state.regs.r[dest] = state.stack[inStack]
	return nil
}
