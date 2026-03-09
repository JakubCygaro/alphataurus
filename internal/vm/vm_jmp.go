package vm

import (
	"encoding/binary"

	"github.com/JakubCygaro/alphataurus/internal/vm/errors"
)

func (state *VmState) jmp(lastByte byte, param []byte) error {
	dest := binary.BigEndian.Uint64(param)
	if dest < state.exeSegBase || dest >= uint64(state.stackSegBase) {
		return errors.SegmentationFault(dest, uint64(state.byteCodePos))
	}
	state.setIp(dest)
	return nil
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
