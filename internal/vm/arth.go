package vm

import (
	"math"
)

func addValues(a, b any, ty byte, out *uint64) error {
	aInt := a.(uint64)
	bInt := b.(uint64)
	switch ty {
	case TY_UINT64:
		*out = aInt + bInt
	case TY_INT64:
		*out = uint64(int64(aInt) + int64(bInt))
	case TY_FLOAT64:
		*out = math.Float64bits(math.Float64frombits(aInt) + math.Float64frombits(bInt))
	}
	return nil
}
func subValues(a, b any, ty byte, out *uint64) error {
	aInt := a.(uint64)
	bInt := b.(uint64)
	switch ty {
	case TY_UINT64:
		*out = aInt - bInt
	case TY_INT64:
		*out = uint64(int64(aInt) - int64(bInt))
	case TY_FLOAT64:
		*out = math.Float64bits(math.Float64frombits(aInt) - math.Float64frombits(bInt))
	}
	return nil
}
