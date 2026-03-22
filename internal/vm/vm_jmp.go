package vm

import (
	"encoding/binary"

	"github.com/JakubCygaro/alphataurus/internal/vm/errors"
)

func (state *VmState) jmpImpl(dest uint64) error {
	if dest < state.exeSegBase || dest >= uint64(state.stackSegBase) {
		return errors.SegmentationFault(dest, uint64(state.byteCodePos))
	}
	state.setIp(dest)
	return nil
}
func (state *VmState) jmp(lastByte byte, param []byte) error {
	dest := binary.BigEndian.Uint64(param)
	return state.jmpImpl(dest)
}
func (state *VmState) jmpE(lastByte byte, param []byte) error {
	if state.flags.Zf {
		return state.jmp(lastByte, param)
	}
	return nil
}
func (state *VmState) jmpNE(lastByte byte, param []byte) error {
	if !state.flags.Zf {
		return state.jmp(lastByte, param)
	}
	return nil
}
func (state *VmState) jmpZ(lastByte byte, param []byte) error {
	return state.jmpE(lastByte, param)
}
func (state *VmState) jmpNZ(lastByte byte, param []byte) error {
	return state.jmpNE(lastByte, param)
}
func (state *VmState) jmpG(lastByte byte, param []byte) error {
	if !state.flags.Zf && !state.flags.Sf {
		return state.jmp(lastByte, param)
	}
	return nil
}
func (state *VmState) jmpGE(lastByte byte, param []byte) error {
	if state.flags.Zf || !state.flags.Sf {
		return state.jmp(lastByte, param)
	}
	return nil
}
func (state *VmState) jmpL(lastByte byte, param []byte) error {
	if !state.flags.Zf && state.flags.Sf {
		return state.jmp(lastByte, param)
	}
	return nil
}
func (state *VmState) jmpLE(lastByte byte, param []byte) error {
	if state.flags.Zf || state.flags.Sf {
		return state.jmp(lastByte, param)
	}
	return nil
}
func (state *VmState) jmpIP(lastByte byte, param []byte) error {
	offset := int64(binary.BigEndian.Uint64(param))
	ip := int64(state.regs.r[IP_IDX])
	reg := (lastByte & 0x0f)
	regV := int64(0)
	if isMovRRAllowed(reg) {
		regV = int64(state.regs.r[reg])
	}
	opTy := (lastByte & 0xf0) >> 4
	dest := uint64(0)
	switch opTy {
	case OP_TADD:
		dest = uint64(ip + regV + offset)
	case OP_TSUB:
		dest = uint64(ip + regV - offset)
	case OP_TDIV:
		if offset == 0 {
			return errors.BadArthmeticOperation(state.byteCodePos)
		}
		dest = uint64(ip + regV / offset)
	case OP_TMUL:
		dest = uint64(ip + regV * offset)
	}
	return state.jmpImpl(dest)
}
func (state *VmState) jmpEIP(lastByte byte, param []byte) error {
	if state.flags.Zf {
		return state.jmpIP(lastByte, param)
	}
	return nil
}
func (state *VmState) jmpNEIP(lastByte byte, param []byte) error {
	if !state.flags.Zf {
		return state.jmpIP(lastByte, param)
	}
	return nil
}
