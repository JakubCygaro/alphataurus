package vm

import (
	"encoding/binary"
	"math"

	"github.com/JakubCygaro/alphataurus/pkg/vm/errors"
)

func (state *VmState) incR(param []byte) error {
	reg := binary.BigEndian.Uint64(param)
	if !IsGpReg(byte(reg)) {
		return errors.DisallowedOp1Register(int(reg), state.byteCodePos)
	}
	r := &state.regs.r[reg]
	r.IncrementRegU64(1)
	return nil
}
func (state *VmState) decR(param []byte) error {
	reg := binary.BigEndian.Uint64(param)
	if !IsGpReg(byte(reg)) {
		return errors.DisallowedOp1Register(int(reg), state.byteCodePos)
	}
	r := &state.regs.r[reg]
	r.DecrementRegU64(1)
	return nil
}

func (state *VmState) arthRR(opType int, param []byte) error {
	data, err := state.arthRRGetParameters(param)
	if err != nil {
		return err
	}
	srcV := binary.BigEndian.Uint64(state.regs.r[data.src][:])
	destV := binary.BigEndian.Uint64(state.regs.r[data.dest][:])
	switch OpCodeVal(opType) {
	case OP_ADDRR:
		err = state.addValues(srcV, destV,
			data.ty, data.r2sz,
			state.regs.r[data.dest][:])
	case OP_SUBRR:
		err = state.subValues(destV, srcV,
			data.ty, data.r2sz,
			state.regs.r[data.dest][:])
	case OP_MULRR:
		srcV := binary.BigEndian.Uint64(state.regs.r[R0_IDX][:])
		destV := binary.BigEndian.Uint64(state.regs.r[R1_IDX][:])
		err = state.mulValues(srcV, destV,
			data.ty, data.r2sz,
			state.regs.r[R2_IDX][:])
	case OP_DIVRR:
		srcV := binary.BigEndian.Uint64(state.regs.r[R0_IDX][:])
		destV := binary.BigEndian.Uint64(state.regs.r[R1_IDX][:])
		err = state.divValues(srcV, destV,
			data.ty, data.r2sz,
			state.regs.r[R2_IDX][:],
			state.regs.r[R3_IDX][:])
	}
	return err
}
func (state *VmState) arthIR(opType int, lastByte byte, param []byte) error {
	data, err := state.arthIRGetParameters(lastByte)
	if err != nil {
		return err
	}
	immV := binary.BigEndian.Uint64(param)
	regV := state.GetRegVAsU64(int(data.reg), SZ_64)
	switch OpCodeVal(opType) {
	case OP_ADDIR:
		err = state.addValues(regV, immV,
			data.ty, data.r1sz,
			state.regs.r[data.reg][:])
	case OP_SUBIR:
		err = state.subValues(regV, immV,
			data.ty, data.r1sz,
			state.regs.r[data.reg][:])
	}
	return nil
}

func addUint(a, b uint64, dataSz byte, out []byte) (res uint64, of bool) {
	offset := DataSizeToByteCount(dataSz)
	switch dataSz {
	case SZ_8:
		high := (a >> 4) + (b >> 4)
		low := (a & 0x0000_0000_0000_000f) + (b & 0x0000_0000_0000_000f)
		res = (high << 4) + low
		of = (high>>4) > 0 || (low>>4) > 0
		out[8-offset] = byte(res)
	case SZ_16:
		high := (a >> 8) + (b >> 8)
		low := (a & 0x0000_0000_0000_00ff) + (b & 0x0000_0000_0000_00ff)
		res = (high << 8) + low
		of = (high>>8) > 0 || (low>>8) > 0
		binary.BigEndian.PutUint16(out[8-offset:], uint16(res))
	case SZ_32:
		high := (a >> 16) + (b >> 16)
		low := (a & 0x0000_0000_0000_ffff) + (b & 0x0000_0000_0000_ffff)
		res = (high << 16) + low
		of = (high>>16) > 0 || (low>>16) > 0
		binary.BigEndian.PutUint32(out[8-offset:], uint32(res))
	case SZ_64:
		high := (a >> 32) + (b >> 32)
		low := (a & 0x0000_0000_ffff_ffff) + (b & 0x0000_0000_ffff_ffff)
		res = (high << 32) + low
		of = (high>>32) > 0 || (low>>32) > 0
		binary.BigEndian.PutUint64(out[8-offset:], uint64(res))
	}
	return res, of
}
func addSint(a, b uint64, dataSz byte, out []byte) (res int64) {
	offset := DataSizeToByteCount(dataSz)
	res = int64(a) + int64(b)
	switch dataSz {
	case SZ_8:
		out[8-offset] = byte(res)
	case SZ_16:
		binary.BigEndian.PutUint16(out[8-offset:], uint16(res))
	case SZ_32:
		binary.BigEndian.PutUint32(out[8-offset:], uint32(res))
	case SZ_64:
		binary.BigEndian.PutUint64(out[8-offset:], uint64(res))
	}
	return res
}

func (state *VmState) addValues(a, b uint64, ty, dataSz byte, out []byte) error {
	var res uint64
	switch ty {
	case TY_UINT:
		res, state.flags.Of = addUint(a, b, dataSz, out)
	case TY_SINT:
		res = uint64(addSint(a, b, dataSz, out))
	case TY_FLOAT:
		if dataSz != SZ_64 {
			return errors.BadArthmeticOperation(state.byteCodePos)
		}
		res = math.Float64bits(math.Float64frombits(a) + math.Float64frombits(b))
		binary.BigEndian.PutUint64(
			out[:],
			res,
		)
	}
	state.postArthSetFlags(a, res, dataSz, ty == TY_FLOAT)
	return nil
}
func subUint(a, b uint64, dataSz byte, out []byte) (res uint64) {
	offset := DataSizeToByteCount(dataSz)
	res = a - b
	switch dataSz {
	case SZ_8:
		out[8-offset] = byte(res)
	case SZ_16:
		binary.BigEndian.PutUint16(out[8-offset:], uint16(res))
	case SZ_32:
		binary.BigEndian.PutUint32(out[8-offset:], uint32(res))
	case SZ_64:
		binary.BigEndian.PutUint64(out[8-offset:], uint64(res))
	}
	return res
}
func subSint(a, b uint64, dataSz byte, out []byte) (res int64) {
	offset := DataSizeToByteCount(dataSz)
	res = int64(a) - int64(b)
	switch dataSz {
	case SZ_8:
		out[8-offset] = byte(res)
	case SZ_16:
		binary.BigEndian.PutUint16(out[8-offset:], uint16(res))
	case SZ_32:
		binary.BigEndian.PutUint32(out[8-offset:], uint32(res))
	case SZ_64:
		binary.BigEndian.PutUint64(out[8-offset:], uint64(res))
	}
	return res
}
func (state *VmState) postArthSetFlags(a, res uint64, dataSz byte, isFloat bool) {
	state.flags.Pf = (res & 1) == 0
	switch dataSz {
	case SZ_8:
		state.flags.Cf = a&0xffff_ffff_ffff_ff00 != res&0xffff_ffff_ffff_ff00
		state.flags.Sf = int8(res) < 0
		state.flags.Zf = int8(res) == 0
	case SZ_16:
		state.flags.Cf = a&0xffff_ffff_ffff_0000 != res&0xffff_ffff_ffff_0000
		state.flags.Sf = int16(res) < 0
		state.flags.Zf = int16(res) == 0
	case SZ_32:
		state.flags.Cf = a&0xffff_ffff_0000_0000 != res&0xffff_ffff_0000_0000
		state.flags.Sf = int32(res) < 0
		state.flags.Zf = int32(res) == 0
	case SZ_64:
		if isFloat {
			f := math.Float64frombits(res)
			state.flags.Sf = f < 0.0
			state.flags.Zf = f == 0.0
		} else {
			state.flags.Sf = int64(res) < 0
			state.flags.Zf = int64(res) == 0
		}
	}
}
func (state *VmState) subValues(a, b uint64, ty, dataSz byte, out []byte) error {
	var res uint64
	switch ty {
	case TY_UINT:
		res = subUint(a, b, dataSz, out)
	case TY_SINT:
		res = uint64(subSint(a, b, dataSz, out))
	case TY_FLOAT:
		if dataSz != SZ_64 {
			return errors.BadArthmeticOperation(state.byteCodePos)
		}
		res = math.Float64bits(math.Float64frombits(a) - math.Float64frombits(b))
		binary.BigEndian.PutUint64(
			out[:],
			res,
		)
	}
	state.postArthSetFlags(a, res, dataSz, ty == TY_FLOAT)
	return nil
}
func mulUint(a, b uint64, dataSz byte, out []byte) {
	offset := DataSizeToByteCount(dataSz)
	switch dataSz {
	case SZ_8:
		out[8-offset] = byte(a) * byte(b)
	case SZ_16:
		binary.BigEndian.PutUint16(out[8-offset:], uint16(a)*uint16(b))
	case SZ_32:
		binary.BigEndian.PutUint32(out[8-offset:], uint32(a)*uint32(b))
	case SZ_64:
		binary.BigEndian.PutUint64(out[8-offset:], uint64(a)*uint64(b))
	}
}
func mulSint(a, b uint64, dataSz byte, out []byte) {
	offset := DataSizeToByteCount(dataSz)
	switch dataSz {
	case SZ_8:
		out[8-offset] = byte(int8(a) * int8(b))
	case SZ_16:
		binary.BigEndian.PutUint16(out[8-offset:], uint16(int16(a)*int16(b)))
	case SZ_32:
		binary.BigEndian.PutUint32(out[8-offset:], uint32(int32(a)*int32(b)))
	case SZ_64:
		binary.BigEndian.PutUint64(out[8-offset:], uint64(int64(a)*int64(b)))
	}
}
func (state *VmState) mulValues(a, b uint64, ty, dataSz byte, out []byte) error {
	switch ty {
	case TY_UINT:
		mulUint(a, b, dataSz, out)
	case TY_SINT:
		mulSint(a, b, dataSz, out)
	case TY_FLOAT:
		if dataSz != SZ_64 {
			return errors.BadArthmeticOperation(state.byteCodePos)
		}
		binary.BigEndian.PutUint64(
			out[:],
			math.Float64bits(math.Float64frombits(a)*math.Float64frombits(b)),
		)
	}
	return nil
}
func divUint(a, b uint64, dataSz byte, out []byte) {
	offset := DataSizeToByteCount(dataSz)
	switch dataSz {
	case SZ_8:
		out[8-offset] = byte(a) / byte(b)
	case SZ_16:
		binary.BigEndian.PutUint16(out[8-offset:], uint16(a)/uint16(b))
	case SZ_32:
		binary.BigEndian.PutUint32(out[8-offset:], uint32(a)/uint32(b))
	case SZ_64:
		binary.BigEndian.PutUint64(out[8-offset:], uint64(a)/uint64(b))
	}
}
func divSint(a, b uint64, dataSz byte, out []byte) {
	offset := DataSizeToByteCount(dataSz)
	switch dataSz {
	case SZ_8:
		out[8-offset] = byte(int8(a) / int8(b))
	case SZ_16:
		binary.BigEndian.PutUint16(out[8-offset:], uint16(int16(a)/int16(b)))
	case SZ_32:
		binary.BigEndian.PutUint32(out[8-offset:], uint32(int32(a)/int32(b)))
	case SZ_64:
		binary.BigEndian.PutUint64(out[8-offset:], uint64(int64(a)/int64(b)))
	}
}
func modUint(a, b uint64, dataSz byte, out []byte) {
	offset := DataSizeToByteCount(dataSz)
	switch dataSz {
	case SZ_8:
		out[8-offset] = byte(a) % byte(b)
	case SZ_16:
		binary.BigEndian.PutUint16(out[8-offset:], uint16(a)%uint16(b))
	case SZ_32:
		binary.BigEndian.PutUint32(out[8-offset:], uint32(a)%uint32(b))
	case SZ_64:
		binary.BigEndian.PutUint64(out[8-offset:], uint64(a)%uint64(b))
	}
}
func modSint(a, b uint64, dataSz byte, out []byte) {
	offset := DataSizeToByteCount(dataSz)
	switch dataSz {
	case SZ_8:
		out[8-offset] = byte(int8(a) % int8(b))
	case SZ_16:
		binary.BigEndian.PutUint16(out[8-offset:], uint16(int16(a)%int16(b)))
	case SZ_32:
		binary.BigEndian.PutUint32(out[8-offset:], uint32(int32(a)%int32(b)))
	case SZ_64:
		binary.BigEndian.PutUint64(out[8-offset:], uint64(int64(a)%int64(b)))
	}
}
func (state *VmState) divValues(a, b uint64, ty, dataSz byte, quoitent, rem []byte) error {
	switch ty {
	case TY_UINT:
		divUint(a, b, dataSz, quoitent)
		modUint(a, b, dataSz, rem)
	case TY_SINT:
		divSint(a, b, dataSz, quoitent)
		modSint(a, b, dataSz, rem)
	case TY_FLOAT:
		if dataSz != SZ_64 {
			return errors.BadArthmeticOperation(state.byteCodePos)
		}
		binary.BigEndian.PutUint64(
			quoitent[:],
			math.Float64bits(math.Float64frombits(a)/math.Float64frombits(b)),
		)
		binary.BigEndian.PutUint64(rem, 0)
	}
	return nil
}
