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
func (state *VmState) movDRO1(byte3, byte4 byte, param []byte) error {
	dest := (byte4 & 0xf0) >> 4
	reg := (byte4 & 0x0f)
	opTy := byte3
	if !isMovRRAllowed(dest) {
		return errors.DisallowedDestRegister(int(dest), state.byteCodePos)
	}
	if !isMovRRAllowed(reg) {
		return errors.DisallowedOp2Register(int(reg), state.byteCodePos)
	}
	regV := int64(state.regs.r[reg])
	var addr int64
	p := int64(binary.BigEndian.Uint64(param))
	switch opTy {
	case OP_TADD:
		addr = regV + p
	case OP_TSUB:
		addr = regV - p
	case OP_TMUL:
		addr = regV * p
	case OP_TDIV:
		addr = regV / p
	default:
		return errors.BadOpcode(state.currentOpcode, state.byteCodePos)
	}
	inStack := state.VirtToRealSp(int(addr))
	if inStack < 0 || inStack >= len(state.stack) {
		return errors.SegmentationFault(uint64(addr), state.byteCodePos)
	}
	state.regs.r[dest] = state.stack[inStack]
	return nil
}
func (state *VmState) movDRO2(byte3, byte4 byte, param []byte) error {
	dest := (byte4 & 0xf0) >> 4
	reg1 := (byte4 & 0x0f)
	reg2 := (byte3 & 0xf0) >> 4
	opTy := (byte3 & 0x0f)
	if !isMovRRAllowed(dest) {
		return errors.DisallowedDestRegister(int(dest), state.byteCodePos)
	}
	if !isMovRRAllowed(reg1) {
		return errors.DisallowedOp2Register(int(reg1), state.byteCodePos)
	}
	if !isMovRRAllowed(reg2) {
		return errors.DisallowedOp2Register(int(reg2), state.byteCodePos)
	}
	reg1V := int64(state.regs.r[reg1])
	reg2V := int64(state.regs.r[reg2])
	var addr int64
	p := int64(binary.BigEndian.Uint64(param))
	switch opTy {
	case OP_TADD:
		addr = reg1V + reg2V + p
	case OP_TSUB:
		addr = reg1V + reg2V - p
	default:
		return errors.BadOpcode(state.currentOpcode, state.byteCodePos)
	}
	inStack := state.VirtToRealSp(int(addr))
	if inStack < 0 || inStack >= len(state.stack) {
		return errors.SegmentationFault(uint64(addr), state.byteCodePos)
	}
	state.regs.r[dest] = state.stack[inStack]
	return nil
}
