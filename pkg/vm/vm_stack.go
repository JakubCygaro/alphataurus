package vm

import (
	"encoding/binary"
	"github.com/JakubCygaro/alphataurus/pkg/vm/errors"
)

type VmStack []byte

func (state *VmState) push(opTy int, lastByte byte, param []byte) error {
	var val []byte
	// 2 bits for the data size
	dataSz := (0b00000011 & lastByte)
	byteSpan := DataSizeToByteCount(dataSz)
	switch OpCodeVal(opTy) {
	case OP_PUSHI:
		val = param[8-byteSpan:]
	case OP_PUSHR:
		reg := binary.BigEndian.Uint64(param)
		if IsPushRAllowed(byte(reg)) {
			val = state.regs.r[reg][8-byteSpan:]
		} else {
			return errors.DisallowedOp1Register(int(reg), state.byteCodePos)
		}
	}
	return state.pushImpl(val)
}

func (state *VmState) pushImpl(data []byte) error {
	inc := len(data)
	if state.GetRealSp()+inc >= len(state.stack) {
		return errors.StackOverflow(state.byteCodePos)
	}
	r := &state.regs.r[SP_IDX]
	r.IncrementRegU64(uint64(inc))
	copy(
		state.stack[state.GetRealSp()-inc+1:state.GetRealSp()+1],
		data[:],
	)
	return nil
}
func (state *VmState) popR(lastByte byte, param []byte) error {
	dataSz := lastByte & 0b00000011
	val, err := state.popImpl(
		DataSizeToByteCount(dataSz),
	)
	if err != nil {
		return err
	}
	reg := binary.BigEndian.Uint64(param)
	if IsPushRAllowed(byte(reg)) {
		copy(
			state.regs.r[reg][8-len(val):],
			val[:],
		)
	}
	return nil
}
func (state *VmState) popImpl(bytes int) ([]byte, error) {
	if state.GetRealSp() < 0 {
		return nil, errors.StackUnderflow(state.byteCodePos)
	}
	if state.GetRealSp() >= len(state.stack) {
		return nil, errors.SegmentationFault(uint64(state.GetRealSp()), state.byteCodePos)
	}

	// val := state.stack[state.GetRealSp()-bytes+1 : state.GetRealSp()+1]
	val := state.getStackSliceAt(state.GetRealSp(), bytes)
	var sp = binary.BigEndian.Uint64(state.regs.r[SP_IDX][:])
	sp -= uint64(bytes)
	binary.BigEndian.PutUint64(state.regs.r[SP_IDX][:], sp)
	return val, nil
}
func (state *VmState) getStackSliceAt(realAddr int, bytes int) VmStack {
	s := state.stack[realAddr-bytes+1 : realAddr+1]
	return s
}
func (state *VmState) putValInStackWithSize(
	dataSz byte, val uint64, address uint64) error {
	if inStack, err := state.isWithinStack(address); err != nil {
		return err
	} else {
		// s := state.stack[inStack-bytes+1 : inStack+1]
		bytes := DataSizeToByteCount(dataSz)
		s := state.getStackSliceAt(inStack, bytes)
		switch dataSz {
		case SZ_8:
			s[0] = byte(val)
		case SZ_16:
			binary.BigEndian.PutUint16(s[:], uint16(val))
		case SZ_32:
			binary.BigEndian.PutUint32(s[:], uint32(val))
		case SZ_64:
			binary.BigEndian.PutUint64(s[:], uint64(val))
		}
	}
	return nil
}
