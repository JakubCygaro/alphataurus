package vm

import (
	"encoding/binary"
	"fmt"
	"math"
)

type Register [8]byte

func RegisterWithValue(val uint64) (Register) {
	r := Register{}
	r.PutValWithSize(SZ_64, val)
	return r
}
func RegisterWithValueSized(val uint64, size byte) (Register) {
	r := Register{}
	r.PutValWithSize(size, val)
	return r
}

func (r Register) ToDisplayString() string {
	return fmt.Sprintf(`[%v_u8 | %v_u16 | %v_u32 | %v_u64 | %v_f64]`,
		r[7],
		binary.BigEndian.Uint16(r[6:]),
		binary.BigEndian.Uint32(r[4:]),
		binary.BigEndian.Uint64(r[:]),
		math.Float64frombits(binary.BigEndian.Uint64(r[:])),
	)
}
func (r *Register) PutValWithSize(dataSz byte, val uint64) {
	bytes := dataSizeToByteCount(dataSz)
	switch dataSz {
	case SZ_8:
		(*r)[7] = byte(val)
	case SZ_16:
		binary.BigEndian.PutUint16((*r)[8-bytes:], uint16(val))
	case SZ_32:
		binary.BigEndian.PutUint32((*r)[8-bytes:], uint32(val))
	case SZ_64:
		binary.BigEndian.PutUint64((*r)[8-bytes:], uint64(val))
	}
}
func (r* Register) GetValAs(ty, dataSz byte, out *any) error {
	if err := IsValidDataSize(dataSz); err != nil {
		return err
	}
	if err := IsValidDataType(ty); err != nil {
		return err
	}
	if ty == TY_FLOAT && dataSz != SZ_64 {
		return fmt.Errorf("Data type FLOAT64 only supports 64-bit size")
	}
	val := binary.BigEndian.Uint64((*r)[:])
	switch ty {
	case TY_UINT:
		switch dataSz {
		case SZ_8:
			*out = uint8(val)
		case SZ_16:
			*out = uint16(val)
		case SZ_32:
			*out = uint32(val)
		case SZ_64:
			*out = uint64(val)
		}
	case TY_SINT:
		switch dataSz {
		case SZ_8:
			*out = int8(val)
		case SZ_16:
			*out = int16(val)
		case SZ_32:
			*out = int32(val)
		case SZ_64:
			*out = int64(val)
		}
	case TY_FLOAT:
		*out = float64(math.Float64frombits(val))
	}
	return nil
}
