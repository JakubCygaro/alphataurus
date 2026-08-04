package vm

import (
	"encoding/binary"
	"math"
	// "math"

	"github.com/JakubCygaro/alphataurus/pkg/vm/errors"
)

func (state *VmState) logIR(opType int, lastByte byte, param []byte) error {
	first := 0b0000_1111 & lastByte
	dataSize := (0b0011_0000 & lastByte) >> 4
	fVal, sVal :=
		state.GetRegVAsU64(int(first), dataSize),
		binary.BigEndian.Uint64(param)
	fVal = state.logImpl(fVal, sVal, OpCodeVal(opType))
	state.putValInRegWithSize(int(first), dataSize, fVal)
	return nil
}
func (state *VmState) logRR(opType int, param []byte) error {
	data, err := state.arthRRGetParameters(param)
	if err != nil {
		return err
	}
	if data.r1sz != data.r2sz {
		return errors.BadOperandSizes(data.r1sz, data.r2sz, state.byteCodePos)
	}
	fVal, sVal :=
		state.GetRegVAsU64(int(data.src), data.r1sz),
		state.GetRegVAsU64(int(data.dest), data.r2sz)
	fVal = state.logImpl(fVal, sVal, OpCodeVal(opType))
	state.putValInRegWithSize(int(data.src), data.r1sz, fVal)
	return nil
}
func (state *VmState) logImpl(a, b uint64, opType OpCodeVal) uint64 {
	switch {
	case opType == OP_ORRR || opType == OP_ORIR:
		a = a | b
	case opType == OP_ANDRR || opType == OP_ANDIR:
		a = a & b
	case opType == OP_XORRR || opType == OP_XORIR:
		a = a ^ b
	case opType == OP_LSHRR || opType == OP_LSHIR:
		a = a << b
	case opType == OP_RSHRR || opType == OP_RSHIR:
		a = a >> b
	}
	return a
}
func (state *VmState) not(param []byte) error {
	reg := param[0]
	dataSz := param[1]
	if !IsGpReg(byte(reg)) {
		return errors.DisallowedOp1Register(int(reg), state.byteCodePos)
	}
	regV := ^state.GetRegVAsU64(int(reg), dataSz)
	state.putValInRegWithSize(int(reg), dataSz, regV)
	return nil
}
func (state *VmState) neg(param []byte) error {
	reg := param[0]
	dataSz := param[1]
	ty := param[2]
	if !IsGpReg(byte(reg)) {
		return errors.DisallowedOp1Register(int(reg), state.byteCodePos)
	}
	switch ty {
	case TY_FLOAT:
		regV := -state.GetRegVAsF64(int(reg))
		state.putValInRegWithSize(int(reg), dataSz, math.Float64bits(regV))
	default:
		regV := -state.GetRegVAsU64(int(reg), dataSz)
		state.putValInRegWithSize(int(reg), dataSz, regV)
	}
	return nil
}
func (state *VmState) clr() error {
	state.flags = Flags{}
	return nil
}
func (state *VmState) cdf() error {
	state.flags.Df = false
	return nil
}
func (state *VmState) sdf() error {
	state.flags.Df = true
	return nil
}
func (state *VmState) cmpRR(lastByte byte, param []byte) error {
	var subtrahend, minuend, dataSz, ty byte
	subtrahend |= (lastByte & 0b1111_0000) >> 4
	minuend |= (lastByte & 0b0000_1111)
	ty = param[0]
	dataSz = param[1]
	var subV, minV uint64
	minV = state.GetRegVAsU64(int(minuend), SZ_64)
	subV = state.GetRegVAsU64(int(subtrahend), SZ_64)

	if dataSz != SZ_64 && ty == TY_FLOAT {
		return errors.BadArthmeticOperation(state.byteCodePos)
	}
	return state.cmpImpl(minV, subV, ty, dataSz)
}
func (state *VmState) cmpIR(lastByte byte, param []byte) error {
	var minuend, dataSz, ty byte
	minuend |= (lastByte & 0b11110000) >> 4
	dataSz |= (lastByte & 0b00001100) >> 2
	ty = (lastByte & 0b00000011)
	var subV, minV uint64
	minV = state.GetRegVAsU64(int(minuend), SZ_64)
	subV = binary.BigEndian.Uint64(param)

	return state.cmpImpl(minV, subV, ty, dataSz)
}
func (state *VmState) cmpImpl(minV, subV uint64, ty, dataSz byte) error {
	diff := [8]byte{}
	state.flags = Flags{}
	err := state.subValues(minV, subV,
		ty, dataSz,
		diff[:])
	return err
}
