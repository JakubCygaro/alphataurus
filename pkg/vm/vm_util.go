package vm

import (
	"fmt"

	"github.com/JakubCygaro/alphataurus/pkg/vm/errors"
)

func (state *VmState) isWithinStack(addr uint64) (int, error) {
	inStack := state.VirtToRealSp(int(addr))
	if inStack < 0 || inStack >= len(state.stack) {
		return -1, errors.SegmentationFault(addr, state.byteCodePos)
	}
	return inStack, nil
}

func IsValidDataSize(sz byte) error {
	if sz < SZ_8 || sz > SZ_64 {
		return fmt.Errorf("Invalid data size")
	}
	return nil
}
func IsValidDataType(sz byte) error {
	if sz < TY_SINT || sz > TY_FLOAT {
		return fmt.Errorf("Invalid data type")
	}
	return nil
}

// Check if register is a valid general purpose register index
func IsGpReg(b byte) bool {
	return b <= GP_REG_MAX
}
func IsMovIntoRAllowed(b byte) bool {
	return b < IP_IDX
}
func IsMovFromRAllowed(b byte) bool {
	return b <= IP_IDX
}
func IsArthRAllowed(a byte) bool {
	return IsMovIntoRAllowed(a)
}
func IsLogRAllowed(a byte) bool {
	return IsMovIntoRAllowed(a)
}

// Check if moving from or into this register is allowed
func IsMovRRAllowed(b byte) bool {
	return IsMovFromRAllowed(b) && IsMovIntoRAllowed(b)
}

// Check if pushing the value of this register onto the stack is allowed
func IsPushRAllowed(b byte) bool {
	return IsGpReg(b) || b == SP_IDX || b == BP_IDX
}

// Get the byte count for a given data size value
func DataSizeToByteCount(dataSz byte) int {
	switch dataSz {
	case SZ_8:
		return 1
	case SZ_16:
		return 2
	case SZ_32:
		return 4
	case SZ_64:
		return 8
	}
	return -1
}
func (state *VmState) GetRegVAsU64(reg int, dataSz byte) uint64 {
	ret := uint64(0)
	r := &state.regs.r[reg]
	switch dataSz {
	case SZ_8:
		ret = uint64(r.GetValAsU8())
	case SZ_16:
		ret = uint64(r.GetValAsU16())
	case SZ_32:
		ret = uint64(r.GetValAsU32())
	case SZ_64:
		ret = uint64(r.GetValAsU64())
	}
	return ret
}
func (state *VmState) GetRegVAsS64(reg int, dataSz byte) int64 {
	ret := int64(0)
	r := &state.regs.r[reg]
	switch dataSz {
	case SZ_8:
		ret = int64(r.GetValAsU8())
	case SZ_16:
		ret = int64(r.GetValAsU16())
	case SZ_32:
		ret = int64(r.GetValAsU32())
	case SZ_64:
		ret = int64(r.GetValAsU64())
	}
	return ret
}
func (state *VmState) putValInRegWithSize(reg int, dataSz byte, val uint64) {
	r := &(state.regs.r[reg])
	r.PutValWithSize(dataSz, val)
}
