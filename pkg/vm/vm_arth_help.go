package vm

import (
	"github.com/JakubCygaro/alphataurus/pkg/vm/errors"
)

type arthIRParamData struct {
	reg, ty byte
	r1sz    byte
}

// Get parameters for IR arth instruction
func (state *VmState) arthIRGetParameters(lastByte byte) (data arthIRParamData, err error) {
	ret := arthIRParamData{}
	ret.reg |= (lastByte & 0b0000_1111)
	ret.r1sz |= (lastByte & 0b0011_0000) >> 4
	ret.ty |= (lastByte & 0b1100_0000) >> 6
	if !IsArthRAllowed(ret.reg) && ret.reg != BP_IDX && ret.reg != SP_IDX {
		return ret, errors.DisallowedDestRegister(int(ret.reg), state.byteCodePos)
	}
	return ret, nil
}

type arthRRParamData struct {
	src, dest, ty byte
	r1sz, r2sz    byte
}

func (state *VmState) arthRRGetParameters(param []byte) (data arthRRParamData, err error) {
	data.src = param[0]
	data.dest = param[1]
	data.ty = (param[2] & 0b00000011)
	data.r1sz = (param[2] & 0b00001100) >> 2
	data.r2sz = (param[2] & 0b00110000) >> 4
	if !IsMovFromRAllowed(data.src) {
		return data, errors.DisallowedSrcRegister(int(data.src), state.byteCodePos)
	}
	if !IsArthRAllowed(data.dest) {
		return data, errors.DisallowedDestRegister(int(data.dest), state.byteCodePos)
	}
	return data, nil
}
