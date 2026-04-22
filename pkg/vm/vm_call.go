package vm

import (
	"encoding/binary"
)

func (state *VmState) call(param []byte) error {
	if err := state.pushImpl(state.regs.r[IP_IDX][:]); err != nil {
		return err
	}
	jumpTo := binary.BigEndian.Uint64(param)
	return state.jmpImpl(jumpTo)
}
func (state *VmState) callIP(lastByte byte, param []byte) error {
	if err := state.pushImpl(state.regs.r[IP_IDX][:]); err != nil {
		return err
	}
	return state.jmpIP(lastByte, param)
}
func (state *VmState) ret() error {
	if returnAddr, err := state.popImpl(8); err != nil {
		return err
	} else {
		return state.jmpImpl(binary.BigEndian.Uint64(returnAddr))
	}
}
