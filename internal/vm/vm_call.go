package vm

import (
	"encoding/binary"
)

func (state *VmState) call(param []byte) error {
	nextInst := state.GetIp()
	if err := state.pushImpl(nextInst); err != nil {
		return err
	}
	jumpTo := binary.BigEndian.Uint64(param)
	return state.jmpImpl(jumpTo)
}
func (state *VmState) ret() error {
	if returnAddr, err := state.popImpl(); err != nil {
		return err
	} else {
		return state.jmpImpl(returnAddr)
	}
}
