package vm

import (
	"encoding/binary"
	"math"
	"math/bits"

	// "math"

	"github.com/JakubCygaro/alphataurus/pkg/vm/errors"
)

func (state *VmState) test(opType int, lastByte byte, param []byte) error {
	if ((lastByte >> 7) & 1) == 1 {
		return state.logIR(opType, lastByte, param)
	} else {
		dest := 0b0000_1111 & lastByte
		destSz := (0b0011_0000 & lastByte) >> 4
		src := 0b0000_1111 & param[7]
		fVal, sVal :=
			state.GetRegVAsU64(int(dest), destSz),
			state.GetRegVAsU64(int(src), destSz)
		fVal = state.logImpl(fVal, sVal, destSz, OpCodeVal(opType))
		state.putValInRegWithSize(int(dest), destSz, fVal)
	}
	return nil
}

func (state *VmState) logIR(opType int, lastByte byte, param []byte) error {
	first := 0b0000_1111 & lastByte
	dataSize := (0b0011_0000 & lastByte) >> 4
	fVal, sVal :=
		state.GetRegVAsU64(int(first), dataSize),
		binary.BigEndian.Uint64(param)
	fVal = state.logImpl(fVal, sVal, dataSize, OpCodeVal(opType))
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
	fVal = state.logImpl(fVal, sVal, data.r1sz, OpCodeVal(opType))
	state.putValInRegWithSize(int(data.src), data.r1sz, fVal)
	return nil
}
func (state *VmState) logImpl(a, b uint64, dataSz byte, opType OpCodeVal) uint64 {
	switch opType {
	case OP_ORRR:
		fallthrough
	case OP_ORIR:
		a = a | b
	case OP_ANDRR:
		fallthrough
	case OP_ANDIR:
		a = a & b
	case OP_TEST:
		a = a & b
	case OP_XORRR:
		fallthrough
	case OP_XORIR:
		a = a ^ b
	case OP_LSHRR:
		fallthrough
	case OP_LSHIR:
		a = a << b
	case OP_RSHRR:
		fallthrough
	case OP_RSHIR:
		a = a >> b
	}
	state.flags.Sf = (a >> ((uint64(dataSz+1) * 8) - 1)) == 1
	state.flags.Zf = a == 0
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
func rotate(val uint64, dataSz byte, rotateBy int, right bool) uint64 {
	if right {
		rotateBy = -rotateBy
	}
	switch dataSz {
	case SZ_8:
		return uint64(bits.RotateLeft8(
			uint8(val), rotateBy))
	case SZ_16:
		return uint64(bits.RotateLeft16(
			uint16(val), rotateBy))
	case SZ_32:
		return uint64(bits.RotateLeft32(
			uint32(val), rotateBy))
	case SZ_64:
		return uint64(bits.RotateLeft64(
			uint64(val), rotateBy))
	default:
		return 0
	}
}
func (state *VmState) rolr(param []byte) error {
	dest := (param[7] & 0x0f)
	destSz := (param[7] & 0xf0) >> 4
	src := (param[6] & 0x0f)
	srcSz := (param[6] & 0xf0) >> 4
	if !IsGpReg(byte(dest)) {
		return errors.DisallowedOp1Register(int(dest), state.byteCodePos)
	}
	if !IsGpReg(byte(src)) {
		return errors.DisallowedOp1Register(int(src), state.byteCodePos)
	}
	state.putValInRegWithSize(
		int(dest),
		destSz,
		rotate(
			state.GetRegVAsU64(int(dest), destSz),
			destSz,
			int(state.GetRegVAsU64(int(src), srcSz)),
			false))
	return nil
}
func (state *VmState) rorr(param []byte) error {
	dest := (param[7] & 0x0f)
	destSz := (param[7] & 0xf0) >> 4
	src := (param[6] & 0x0f)
	srcSz := (param[6] & 0xf0) >> 4
	if !IsGpReg(byte(dest)) {
		return errors.DisallowedOp1Register(int(dest), state.byteCodePos)
	}
	if !IsGpReg(byte(src)) {
		return errors.DisallowedOp1Register(int(src), state.byteCodePos)
	}
	state.putValInRegWithSize(
		int(dest),
		destSz,
		rotate(
			state.GetRegVAsU64(int(dest), destSz),
			destSz,
			int(state.GetRegVAsU64(int(src), srcSz)),
			true))
	return nil
}
func (state *VmState) rol(lastByte byte, param []byte) error {
	reg := (lastByte & 0xf0) >> 4
	dataSz := (lastByte & 0x0f)
	rotateBy := binary.BigEndian.Uint64(param)
	if !IsGpReg(byte(reg)) {
		return errors.DisallowedOp1Register(int(reg), state.byteCodePos)
	}
	state.putValInRegWithSize(
		int(reg),
		dataSz,
		rotate(
			state.GetRegVAsU64(int(reg), dataSz),
			dataSz,
			int(rotateBy),
			false))
	return nil
	// switch dataSz {
	// case SZ_8:
	// 	regV := bits.RotateLeft8(
	// 		uint8(state.GetRegVAsU64(int(reg), dataSz)), int(rotateBy))
	// 	state.putValInRegWithSize(int(reg), dataSz, uint64(regV))
	// case SZ_16:
	// 	regV := bits.RotateLeft16(
	// 		uint16(state.GetRegVAsU64(int(reg), dataSz)), int(rotateBy))
	// 	state.putValInRegWithSize(int(reg), dataSz, uint64(regV))
	// case SZ_32:
	// 	regV := bits.RotateLeft32(
	// 		uint32(state.GetRegVAsU64(int(reg), dataSz)), int(rotateBy))
	// 	state.putValInRegWithSize(int(reg), dataSz, uint64(regV))
	// case SZ_64:
	// 	regV := bits.RotateLeft64(
	// 		uint64(state.GetRegVAsU64(int(reg), dataSz)), int(rotateBy))
	// 	state.putValInRegWithSize(int(reg), dataSz, uint64(regV))
	// }
	// return nil
}
func (state *VmState) ror(lastByte byte, param []byte) error {
	reg := (lastByte & 0xf0) >> 4
	dataSz := (lastByte & 0x0f)
	rotateBy := binary.BigEndian.Uint64(param)
	if !IsGpReg(byte(reg)) {
		return errors.DisallowedOp1Register(int(reg), state.byteCodePos)
	}
	state.putValInRegWithSize(
		int(reg),
		dataSz,
		rotate(
			state.GetRegVAsU64(int(reg), dataSz),
			dataSz,
			int(rotateBy),
			true))
	return nil
	// switch dataSz {
	// case SZ_8:
	// 	regV := bits.RotateLeft8(
	// 		uint8(state.GetRegVAsU64(int(reg), dataSz)), -int(rotateBy))
	// 	state.putValInRegWithSize(int(reg), dataSz, uint64(regV))
	// case SZ_16:
	// 	regV := bits.RotateLeft16(
	// 		uint16(state.GetRegVAsU64(int(reg), dataSz)), -int(rotateBy))
	// 	state.putValInRegWithSize(int(reg), dataSz, uint64(regV))
	// case SZ_32:
	// 	regV := bits.RotateLeft32(
	// 		uint32(state.GetRegVAsU64(int(reg), dataSz)), -int(rotateBy))
	// 	state.putValInRegWithSize(int(reg), dataSz, uint64(regV))
	// case SZ_64:
	// 	regV := bits.RotateLeft64(
	// 		uint64(state.GetRegVAsU64(int(reg), dataSz)), -int(rotateBy))
	// 	state.putValInRegWithSize(int(reg), dataSz, uint64(regV))
	// }
	// return nil
}
