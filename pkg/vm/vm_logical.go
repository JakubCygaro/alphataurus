package vm

import (
	"encoding/binary"
	"math"

	"github.com/JakubCygaro/alphataurus/pkg/vm/errors"
)

func (state *VmState) logIR(opType int, lastByte byte, param []byte) error {
	first := 0b0000_1111 & lastByte
	dataSize := (0b0011_0000 & lastByte) >> 4
	fVal, sVal :=
		state.GetRegVAsUint64(int(first), dataSize),
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
		state.GetRegVAsUint64(int(data.src), data.r1sz),
		state.GetRegVAsUint64(int(data.dest), data.r2sz)
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
func (state *VmState) clr() error {
	state.flags = Flags{}
	return nil
}
func (state *VmState) cmpRR(lastByte byte, param []byte) error {
	var subtrahend, minuend, dataSz, ty byte
	subtrahend |= (lastByte & 0b1111_0000) >> 4
	minuend |= (lastByte & 0b0000_1111)
	ty = param[0]
	dataSz = param[1]
	if !IsGpReg(minuend) {
		return errors.DisallowedOp1Register(int(minuend), state.byteCodePos)
	}
	var subV, minV uint64
	minV = binary.BigEndian.Uint64(state.regs.r[minuend][:])
	subV = binary.BigEndian.Uint64(state.regs.r[subtrahend][:])

	if dataSz != SZ_64 && ty == TY_FLOAT {
		return errors.BadArthmeticOperation(state.byteCodePos)
	}
	return state.cmpImpl(minV, subV, ty, dataSz)
}
func (state *VmState) cmpIR(lastByte byte, param []byte) error {
	var subtrahend, minuend, dataSz, ty byte
	subtrahend |= (lastByte & 0b11110000) >> 4
	dataSz |= (lastByte & 0b00001100) >> 2
	ty = (lastByte & 0b00000001)
	if ty == 1 {
		ty = TY_FLOAT
	}
	if !IsGpReg(minuend) {
		return errors.DisallowedOp1Register(int(minuend), state.byteCodePos)
	}
	var subV, minV uint64
	minV = binary.BigEndian.Uint64(state.regs.r[minuend][:])
	subV = binary.BigEndian.Uint64(param)

	if dataSz != SZ_64 && ty == TY_FLOAT {
		return errors.BadArthmeticOperation(state.byteCodePos)
	}
	return state.cmpImpl(minV, subV, ty, dataSz)
}
func (state *VmState) cmpImpl(minV, subV uint64, ty, dataSz byte) error {
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
