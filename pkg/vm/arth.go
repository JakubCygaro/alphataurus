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
func addFloat(a, b uint64, dataSz byte, out []byte) {
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
			math.Float64bits(math.Float64frombits(a) + math.Float64frombits(b)),
		)
	}
	return nil
}
func subValues(a, b uint64, ty byte, out *uint64) {
	switch ty {
	case TY_UINT:
		*out = a - b
	case TY_SINT:
		*out = uint64(int64(a) - int64(b))
	case TY_FLOAT:
		*out = math.Float64bits(math.Float64frombits(a) - math.Float64frombits(b))
	}
}
func mulValues(a, b uint64, ty byte, out *uint64) {
	switch ty {
	case TY_UINT:
		*out = a * b
	case TY_SINT:
		*out = uint64(int64(a) * int64(b))
	case TY_FLOAT:
		*out = math.Float64bits(math.Float64frombits(a) * math.Float64frombits(b))
	}
}
func divValues(a, b uint64, ty byte, quoitent, rem *uint64) {
	switch ty {
	case TY_UINT:
		*quoitent = a / b
		*rem = a % b
	case TY_SINT:
		*quoitent = uint64(int64(a) / int64(b))
		*rem = uint64(int64(a) % int64(b))
	case TY_FLOAT:
		*quoitent = math.Float64bits(math.Float64frombits(a) / math.Float64frombits(b))
		*rem = 0
	}
}
