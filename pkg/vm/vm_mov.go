package vm

import (
	"encoding/binary"
	"github.com/JakubCygaro/alphataurus/pkg/vm/errors"
)

func (state *VmState) movRR(lastByte byte, param []byte) error {
	var src, dest byte
	src |= (param[7] & 0xf0) >> 4
	dest |= (param[7] & 0x0f)
	dataSz := (lastByte & 0b0000_0011)
	if !IsMovRRAllowed(src) {
		return errors.DisallowedSrcRegister(int(src), state.byteCodePos)
	} else if !IsMovRRAllowed(dest) {
		return errors.DisallowedDestRegister(int(dest), state.byteCodePos)
	} else {
		bytes := DataSizeToByteCount(dataSz)
		copy(
			state.regs.r[dest][8-bytes:],
			state.regs.r[src][8-bytes:],
		)
	}
	return nil
}
func (state *VmState) movIR(lastByte byte, param []byte) error {
	var ty, dest, dataSz byte
	// type of value
	dataSz |= (lastByte & 0b1100_0000) >> 6
	ty |= (lastByte & 0b0011_0000) >> 4
	dest |= (lastByte & 0b0000_1111)
	if !IsGpReg(dest) {
		return errors.DisallowedDestRegister(int(dest), state.byteCodePos)
	}
	switch ty {
	case TY_SINT:
		i64 := binary.BigEndian.Uint64(param)
		state.putValInRegWithSize(int(dest), dataSz, i64)
	case TY_FLOAT:
		bits := binary.BigEndian.Uint64(param)
		state.putValInRegWithSize(int(dest), dataSz, bits)
	default:
		u64 := binary.BigEndian.Uint64(param)
		state.putValInRegWithSize(int(dest), dataSz, u64)
	}
	return nil
}
func (state *VmState) movDRI(lastByte byte, param []byte) error {
	dest := (lastByte & 0b0000_1111)
	dataSz := (lastByte & 0b0011_0000)
	bytes := DataSizeToByteCount(dataSz)
	if !IsMovRRAllowed(dest) {
		return errors.DisallowedDestRegister(int(dest), state.byteCodePos)
	}
	addr := binary.BigEndian.Uint64(param)
	if inStack, e := state.isWithinStack(addr); e != nil {
		return e
	} else {
		copy(
			state.regs.r[dest][8-bytes:],
			state.stack[inStack:inStack+bytes],
		)
	}
	return nil
}

type derefParamsO1 struct {
	dest, reg1, opTy, r1sz, destSz byte
}

func (state *VmState) getDerefParamsO1(byte3, byte4 byte) derefParamsO1 {
	ret := derefParamsO1{}
	ret.dest = (byte4 & 0xf0) >> 4
	ret.reg1 = (byte4 & 0x0f)

	ret.destSz = (byte3 & 0b0011_0000) >> 4
	ret.r1sz = (byte3 & 0b0000_1100) >> 2
	ret.opTy = (byte3 & 0b0000_0011)
	return ret
}

type derefParamsO2 struct {
	dest, reg1, reg2, opTy, r1_2sz, destSz byte
}

func (state *VmState) getDerefParamsO2(byte3, byte4, byten1 byte) derefParamsO2 {
	ret := derefParamsO2{}
	ret.dest = (byte4 & 0b1111_0000) >> 4
	ret.reg1 = (byte4 & 0b0000_1111)

	ret.reg2 = (byte3 & 0b1111_0000) >> 4
	ret.r1_2sz = (byte3 & 0b0000_1100) >> 2
	ret.opTy = (byte3 & 0b0000_0011)

	ret.destSz = (byten1 & 0b1100_0000) >> 6
	return ret
}
func (state *VmState) movDRO1(byte3, byte4 byte, param []byte) error {
	dParams := state.getDerefParamsO1(byte3, byte4)
	if !IsMovRRAllowed(dParams.dest) {
		return errors.DisallowedDestRegister(int(dParams.dest), state.byteCodePos)
	}
	if !IsMovRRAllowed(dParams.reg1) {
		return errors.DisallowedOp2Register(int(dParams.reg1), state.byteCodePos)
	}
	regV := state.GetRegVAsU64(int(dParams.reg1), dParams.r1sz)
	var addr uint64
	p := binary.BigEndian.Uint64(param)
	switch dParams.opTy {
	case OP_TADD:
		addr = regV + p
	case OP_TSUB:
		addr = regV - p
	case OP_TMUL:
		addr = regV * p
	case OP_TDIV:
		if p == 0 {
			return errors.BadArthmeticOperation(state.byteCodePos)
		}
		addr = regV / p
	default:
		return errors.BadOpcode(state.currentOpcode, state.byteCodePos)
	}
	inStack := state.VirtToRealSp(int(addr))
	if inStack < 0 || inStack >= len(state.stack) {
		return errors.SegmentationFault(uint64(addr), state.byteCodePos)
	}
	bytes := DataSizeToByteCount(dParams.destSz)
	copy(
		state.regs.r[dParams.dest][8-bytes:],
		state.stack[inStack:inStack+bytes],
	)
	return nil
}
func (state *VmState) movDRO2(byte3, byte4 byte, param []byte) error {
	dParams := state.getDerefParamsO2(byte3, byte4, param[0])
	if !IsMovRRAllowed(dParams.dest) {
		return errors.DisallowedDestRegister(int(dParams.dest), state.byteCodePos)
	}
	if err := state.isMovXRO2Allowed(dParams); err != nil {
		return err
	}
	reg1V := int64(state.GetRegVAsU64(int(dParams.reg1), dParams.r1_2sz))
	reg2V := int64(state.GetRegVAsU64(int(dParams.reg2), dParams.r1_2sz))
	offset := int64(binary.BigEndian.Uint64(param) & (0x3fff_ffff_ffff_ffff))
	if inStack, e := state.movXDO2GetAddr(reg1V, reg2V, offset, dParams.opTy); e != nil {
		return e
	} else {
		bytes := DataSizeToByteCount(dParams.destSz)
		copy(
			state.regs.r[dParams.dest][8-bytes:],
			state.stack[inStack:inStack+bytes],
		)
	}
	return nil
}
func (state *VmState) movID(lastByte byte, param []byte) error {
	p := binary.BigEndian.Uint64(param)
	dest := (p & 0xffff_ffff_0000_0000) >> 32
	imm := uint64(int32((p & 0x0000_0000_ffff_ffff)))
	dataSz := lastByte & 0b0000_0011
	if inStack, e := state.isWithinStack(dest); e != nil {
		return e
	} else {
		return state.putValInStackWithSize(dataSz, imm, inStack)
	}
}
func (state *VmState) movRD(lastByte byte, param []byte) error {
	p := binary.BigEndian.Uint64(param)
	source := lastByte & 0b0000_1111
	dataSz := lastByte & 0b0011_0000
	dest := p
	if !IsMovRRAllowed(byte(source)) {
		return errors.DisallowedSrcRegister(int(source), state.byteCodePos)
	}
	if inStack, e := state.isWithinStack(dest); e != nil {
		return e
	} else {
		bytes := DataSizeToByteCount(dataSz)
		copy(
			state.regs.r[source][8-bytes:],
			state.stack[inStack:inStack+bytes],
		)
	}
	return nil
}
func (state *VmState) movIDO1(byte3, byte4 byte, param []byte) error {
	dParams := state.getDerefParamsO1(byte3, byte4)
	if !IsMovRRAllowed(dParams.reg1) {
		return errors.DisallowedOp1Register(int(dParams.reg1), state.byteCodePos)
	}
	p := binary.BigEndian.Uint64(param)
	imm := uint64(int32((p & 0x0000_0000_ffff_ffff)))
	offset := (p & 0xffff_ffff_0000_0000) >> 32
	regV := state.GetRegVAsU64(int(dParams.reg1), dParams.r1sz)
	if inStack, e := state.movXDO1GetAddr(regV, offset, dParams.opTy); e != nil {
		return e
	} else {
		return state.putValInStackWithSize(dParams.destSz, imm, inStack)
	}
}
func (state *VmState) movRDO1(byte3, byte4 byte, param []byte) error {
	dParams := state.getDerefParamsO1(byte3, byte4)
	source := dParams.dest
	if !IsMovRRAllowed(source) {
		return errors.DisallowedSrcRegister(int(source), state.byteCodePos)
	}
	if !IsMovRRAllowed(dParams.reg1) {
		return errors.DisallowedOp1Register(int(dParams.reg1), state.byteCodePos)
	}
	p := binary.BigEndian.Uint64(param)
	regV := state.GetRegVAsU64(int(dParams.reg1), dParams.r1sz)
	if inStack, e := state.movXDO1GetAddr(regV, p, dParams.opTy); e != nil {
		return e
	} else {
		val := state.GetRegVAsU64(int(source), dParams.destSz)
		return state.putValInStackWithSize(dParams.destSz, val, inStack)
	}
}
func (state *VmState) movXDO1GetAddr(r1V, off uint64, opTy byte) (int, error) {
	var addr uint64
	switch opTy {
	case OP_TADD:
		addr = r1V + off
	case OP_TSUB:
		addr = r1V - off
	case OP_TMUL:
		addr = r1V * off
	case OP_TDIV:
		addr = r1V / off
	default:
		return -1, errors.BadOpcode(state.currentOpcode, state.byteCodePos)
	}
	if inStack, e := state.isWithinStack(addr); e != nil {
		return inStack, e
	} else {
		return inStack, nil
	}
}
func (state *VmState) movIDO2(byte3, byte4 byte, param []byte) error {
	dParams := state.getDerefParamsO2(byte3, byte4, param[0])
	if err := state.isMovXRO2Allowed(dParams); err != nil {
		return err
	}
	p := binary.BigEndian.Uint64(param) & (0x3fff_ffff_ffff_ffff)
	imm := uint64(int32((p & 0x0000_0000_ffff_ffff)))
	offset := int64((p & 0x3fff_ffff_0000_0000) >> 32)
	reg1V := state.GetRegVAsS64(int(dParams.reg1), dParams.r1_2sz)
	reg2V := state.GetRegVAsS64(int(dParams.reg2), dParams.r1_2sz)
	if inStack, e := state.movXDO2GetAddr(reg1V, reg2V, offset, dParams.opTy); e != nil {
		return e
	} else {
		return state.putValInStackWithSize(dParams.destSz, imm, int(inStack))
	}
}

func (state *VmState) movRDO2(byte3, byte4 byte, param []byte) error {
	dParams := state.getDerefParamsO2(byte3, byte4, param[0])
	source := dParams.dest
	if err := state.isMovXRO2Allowed(dParams); err != nil {
		return err
	}
	if !IsMovRRAllowed(source) {
		return errors.DisallowedSrcRegister(int(source), state.byteCodePos)
	}
	p := binary.BigEndian.Uint64(param) & (0x3fff_ffff_ffff_ffff)
	offset := int64(p)
	reg1V := state.GetRegVAsS64(int(dParams.reg1), dParams.r1_2sz)
	reg2V := state.GetRegVAsS64(int(dParams.reg2), dParams.r1_2sz)
	if inStack, e := state.movXDO2GetAddr(reg1V, reg2V, offset, dParams.opTy); e != nil {
		return e
	} else {
		val := state.GetRegVAsU64(int(source), dParams.destSz)
		return state.putValInStackWithSize(dParams.destSz, val, int(inStack))
	}
}

func (state *VmState) movXDO2GetAddr(r1V, r2V, off int64, opTy byte) (int, error) {
	var addr uint64
	switch opTy {
	case OP_TADD:
		addr = uint64(r1V + r2V + off)
	case OP_TSUB:
		addr = uint64(r1V + r2V - off)
	default:
		return -1, errors.BadOpcode(state.currentOpcode, state.byteCodePos)
	}
	if inStack, e := state.isWithinStack(addr); e != nil {
		return inStack, e
	} else {
		return inStack, nil
	}
}

func (state *VmState) isMovXRO2Allowed(params derefParamsO2) error {
	if !IsMovRRAllowed(params.reg1) {
		return errors.DisallowedOp1Register(int(params.reg1), state.byteCodePos)
	}
	if !IsMovRRAllowed(params.reg2) {
		return errors.DisallowedOp2Register(int(params.reg2), state.byteCodePos)
	}
	return nil
}
