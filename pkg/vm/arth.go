package vm

import (
	"encoding/binary"
	"math"

	"github.com/JakubCygaro/alphataurus/pkg/vm/errors"
)

func addUint(a, b uint64, dataSz byte, out []byte) {
	offset := dataSizeToByteCount(dataSz)
	switch dataSz {
	case SZ_8:
		out[8-offset] = byte(a) + byte(b)
	case SZ_16:
		binary.BigEndian.PutUint16(out[8-offset:], uint16(a)+uint16(b))
	case SZ_32:
		binary.BigEndian.PutUint32(out[8-offset:], uint32(a)+uint32(b))
	case SZ_64:
		binary.BigEndian.PutUint64(out[8-offset:], uint64(a)+uint64(b))
	}
}
func addSint(a, b uint64, dataSz byte, out []byte) {
	offset := dataSizeToByteCount(dataSz)
	switch dataSz {
	case SZ_8:
		out[8-offset] = byte(a) + byte(b)
	case SZ_16:
		binary.BigEndian.PutUint16(out[8-offset:], uint16(int16(a)+int16(b)))
	case SZ_32:
		binary.BigEndian.PutUint32(out[8-offset:], uint32(int32(a)+int32(b)))
	case SZ_64:
		binary.BigEndian.PutUint64(out[8-offset:], uint64(int64(a)+int64(b)))
	}
}

func (state *VmState) addValues(a, b uint64, ty, dataSz byte, out []byte) error {
	switch ty {
	case TY_UINT:
		addUint(a, b, dataSz, out)
	case TY_SINT:
		addSint(a, b, dataSz, out)
	case TY_FLOAT:
		if dataSz != SZ_64 {
			return errors.BadArthmeticOperation(state.byteCodePos)
		}
		binary.BigEndian.PutUint64(
			out[:],
			math.Float64bits(math.Float64frombits(a)+math.Float64frombits(b)),
		)
	}
	return nil
}
func subUint(a, b uint64, dataSz byte, out []byte) {
	offset := dataSizeToByteCount(dataSz)
	switch dataSz {
	case SZ_8:
		out[8-offset] = byte(a) - byte(b)
	case SZ_16:
		binary.BigEndian.PutUint16(out[8-offset:], uint16(a)-uint16(b))
	case SZ_32:
		binary.BigEndian.PutUint32(out[8-offset:], uint32(a)-uint32(b))
	case SZ_64:
		binary.BigEndian.PutUint64(out[8-offset:], uint64(a)-uint64(b))
	}
}
func subSint(a, b uint64, dataSz byte, out []byte) {
	offset := dataSizeToByteCount(dataSz)
	switch dataSz {
	case SZ_8:
		out[8-offset] = byte(int8(a) - int8(b))
	case SZ_16:
		binary.BigEndian.PutUint16(out[8-offset:], uint16(int16(a)-int16(b)))
	case SZ_32:
		binary.BigEndian.PutUint32(out[8-offset:], uint32(int32(a)-int32(b)))
	case SZ_64:
		binary.BigEndian.PutUint64(out[8-offset:], uint64(int64(a)-int64(b)))
	}
}
func (state *VmState) subValues(a, b uint64, ty, dataSz byte, out []byte) error {
	switch ty {
	case TY_UINT:
		subUint(a, b, dataSz, out)
	case TY_SINT:
		subSint(a, b, dataSz, out)
	case TY_FLOAT:
		if dataSz != SZ_64 {
			return errors.BadArthmeticOperation(state.byteCodePos)
		}
		binary.BigEndian.PutUint64(
			out[:],
			math.Float64bits(math.Float64frombits(a)-math.Float64frombits(b)),
		)
	}
	return nil
}
func mulUint(a, b uint64, dataSz byte, out []byte) {
	offset := dataSizeToByteCount(dataSz)
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
	offset := dataSizeToByteCount(dataSz)
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
	offset := dataSizeToByteCount(dataSz)
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
	offset := dataSizeToByteCount(dataSz)
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
	offset := dataSizeToByteCount(dataSz)
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
	offset := dataSizeToByteCount(dataSz)
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
