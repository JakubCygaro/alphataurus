package vm

import (
	"fmt"
	"math"
)

func addValues(a, b uint64, ty byte, out *uint64){
	switch ty {
	case TY_UINT64:
		fmt.Println("UNSIGNED ADDITION")
		*out = a + b
	case TY_INT64:
		fmt.Println("SIGNED ADDITION")
		*out = uint64(int64(a) + int64(b))
	case TY_FLOAT64:
		fmt.Println("FLOAT ADDITION")
		*out = math.Float64bits(math.Float64frombits(a) + math.Float64frombits(b))
	}
}
func subValues(a, b uint64, ty byte, out *uint64){
	switch ty {
	case TY_UINT64:
		*out = a - b
	case TY_INT64:
		*out = uint64(int64(a) - int64(b))
	case TY_FLOAT64:
		*out = math.Float64bits(math.Float64frombits(a) - math.Float64frombits(b))
	}
}
func mulValues(a, b uint64, ty byte, out *uint64){
	switch ty {
	case TY_UINT64:
		*out = a * b
	case TY_INT64:
		*out = uint64(int64(a) * int64(b))
	case TY_FLOAT64:
		*out = math.Float64bits(math.Float64frombits(a) * math.Float64frombits(b))
	}
}
func divValues(a, b uint64, ty byte, quoitent, rem *uint64){
	switch ty {
	case TY_UINT64:
		*quoitent = a / b
		*rem = a % b
	case TY_INT64:
		*quoitent = uint64(int64(a) / int64(b))
		*rem = uint64(int64(a) % int64(b))
	case TY_FLOAT64:
		*quoitent = math.Float64bits(math.Float64frombits(a) / math.Float64frombits(b))
		*rem = 0
	}
}
