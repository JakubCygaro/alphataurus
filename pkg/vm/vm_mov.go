package vm

import (
	"encoding/binary"
	"github.com/JakubCygaro/alphataurus/pkg/vm/errors"
)

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

func (state *VmState) getDerefParamsO2(byte2, byte3, byte4 byte) derefParamsO2 {
	ret := derefParamsO2{}
	ret.dest = (byte4 & 0b1111_0000) >> 4
	ret.reg1 = (byte4 & 0b0000_1111)

	ret.reg2 = (byte3 & 0b1111_0000) >> 4
	ret.r1_2sz = (byte3 & 0b0000_1100) >> 2
	ret.opTy = (byte3 & 0b0000_0011)

	ret.destSz = (byte2 & 0b0000_0011)
	return ret
}

func (state *VmState) movRR(
	lastByte byte,
	param []byte,
	signExtend bool,
	zeroExtend bool,
) error {
	var src, dest byte
	src |= (param[7] & 0xf0) >> 4
	dest |= (param[7] & 0x0f)
	srcSz := (lastByte & 0b0000_0011)
	destSz := (lastByte & 0b0000_1100) >> 2
	if !IsMovFromRAllowed(src) {
		return errors.DisallowedSrcRegister(int(src), state.byteCodePos)
	} else if !IsMovIntoRAllowed(dest) {
		return errors.DisallowedDestRegister(int(dest), state.byteCodePos)
	} else if signExtend {
		state.regs.r[dest].CopyFromRegisterWithSizeSx(
			&state.regs.r[src],
			destSz,
			srcSz,
		)
	} else {
		if zeroExtend {
			state.regs.r[dest].PutValWithSize(SZ_64, 0)
		}
		state.regs.r[dest].CopyFromRegisterWithSize(
			&state.regs.r[src],
			srcSz,
		)
	}
	return nil
}
func (state *VmState) movIR(
	lastByte byte,
	param []byte,
	zeroExtend bool,
) error {
	var dest, dataSz byte
	dataSz |= (lastByte & 0b1100_0000) >> 6
	dest |= (lastByte & 0b0000_1111)
	if !IsMovIntoRAllowed(dest) {
		return errors.DisallowedDestRegister(int(dest), state.byteCodePos)
	}
	bits := binary.BigEndian.Uint64(param)
	if zeroExtend {
		state.putValInRegWithSize(int(dest), SZ_64, 0)
	}
	state.putValInRegWithSize(int(dest), dataSz, bits)
	return nil
}

func (state *VmState) movDR(
	lastByte byte,
	param []byte,
	zeroExtend bool,
) error {
	dest := (lastByte & 0b0000_1111)
	dataSz := (lastByte & 0b0011_0000) >> 4
	if !IsMovIntoRAllowed(dest) {
		return errors.DisallowedDestRegister(int(dest), state.byteCodePos)
	}
	if zeroExtend {
		state.putValInRegWithSize(int(dest), SZ_64, 0)
	}
	addr := binary.BigEndian.Uint64(param)
	return state.copyFromAddressToRegister(&state.regs.r[dest], addr, dataSz)
}

func (state *VmState) movDRO1(
	byte3, byte4 byte,
	param []byte,
	zeroExtend bool,
) error {
	dParams := state.getDerefParamsO1(byte3, byte4)
	if !IsMovIntoRAllowed(dParams.dest) {
		return errors.DisallowedDestRegister(int(dParams.dest), state.byteCodePos)
	}
	if !IsMovRRAllowed(dParams.reg1) {
		return errors.DisallowedOp2Register(int(dParams.reg1), state.byteCodePos)
	}
	regV := state.GetRegVAsS64(int(dParams.reg1), dParams.r1sz)
	offset := int64(binary.BigEndian.Uint64(param))
	if addr, e := state.movXDO1GetAddr(regV, offset, dParams.opTy); e != nil {
		return e
	} else {
		if zeroExtend {
			state.putValInRegWithSize(
				int(dParams.dest),
				SZ_64,
				0,
			)
		}
		return state.copyFromAddressToRegister(
			&state.regs.r[dParams.dest],
			addr,
			dParams.destSz,
		)
	}
}
func (state *VmState) movDRO2(
	byte2, byte3, byte4 byte,
	param []byte,
	zeroExtend bool,
) error {
	dParams := state.getDerefParamsO2(byte2, byte3, byte4)
	if !IsMovIntoRAllowed(dParams.dest) {
		return errors.DisallowedDestRegister(int(dParams.dest), state.byteCodePos)
	}
	if err := state.isMovXRO2Allowed(dParams); err != nil {
		return err
	}
	reg1V := int64(state.GetRegVAsU64(int(dParams.reg1), dParams.r1_2sz))
	reg2V := int64(state.GetRegVAsU64(int(dParams.reg2), dParams.r1_2sz))
	offset := int64(binary.BigEndian.Uint64(param))
	if addr, e := state.movXDO2GetAddr(reg1V, reg2V, offset, dParams.opTy); e != nil {
		return e
	} else {
		if zeroExtend {
			state.putValInRegWithSize(
				int(dParams.dest),
				SZ_64,
				0,
			)
		}
		return state.copyFromAddressToRegister(
			&state.regs.r[dParams.dest],
			addr,
			dParams.destSz,
		)
	}
}
func (state *VmState) movID(lastByte byte, param []byte) error {
	p := binary.BigEndian.Uint64(param)
	dest := (p & 0xffff_ffff_0000_0000) >> 32
	imm := uint64(int32((p & 0x0000_0000_ffff_ffff)))
	dataSz := lastByte & 0b0000_0011
	return state.putValInStackWithSize(dataSz, imm, dest)
}
func (state *VmState) movRD(lastByte byte, param []byte) error {
	p := binary.BigEndian.Uint64(param)
	source := lastByte & 0b0000_1111
	dataSz := lastByte & 0b0011_0000
	dest := p
	if !IsMovFromRAllowed(byte(source)) {
		return errors.DisallowedSrcRegister(int(source), state.byteCodePos)
	}
	return state.copyFromRegisterToAddress(
		&state.regs.r[source],
		dest,
		dataSz,
	)
}
func (state *VmState) movIDO1NoOffset(byte3, byte4 byte, param []byte) error {
	dParams := state.getDerefParamsO1(byte3, byte4)
	if !IsMovRRAllowed(dParams.reg1) {
		return errors.DisallowedOp1Register(int(dParams.reg1), state.byteCodePos)
	}
	p := binary.BigEndian.Uint64(param)
	imm := p
	regV := state.GetRegVAsS64(int(dParams.reg1), dParams.r1sz)
	if addr, e := state.movXDO1GetAddr(regV, 0, OP_TADD); e != nil {
		return e
	} else {
		return state.putValInStackWithSize(dParams.destSz, imm, addr)
	}
}
func (state *VmState) movIDO1(byte3, byte4 byte, param []byte) error {
	dParams := state.getDerefParamsO1(byte3, byte4)
	if !IsMovRRAllowed(dParams.reg1) {
		return errors.DisallowedOp1Register(int(dParams.reg1), state.byteCodePos)
	}
	p := binary.BigEndian.Uint64(param)
	imm := uint64(int32((p & 0x0000_0000_ffff_ffff)))
	offset := int64((p & 0xffff_ffff_0000_0000) >> 32)
	regV := state.GetRegVAsS64(int(dParams.reg1), dParams.r1sz)
	if addr, e := state.movXDO1GetAddr(regV, offset, dParams.opTy); e != nil {
		return e
	} else {
		return state.putValInStackWithSize(dParams.destSz, imm, addr)
	}
}
func (state *VmState) movRDO1(byte3, byte4 byte, param []byte) error {
	dParams := state.getDerefParamsO1(byte3, byte4)
	source := dParams.dest
	if !IsMovFromRAllowed(source) {
		return errors.DisallowedSrcRegister(int(source), state.byteCodePos)
	}
	if !IsMovRRAllowed(dParams.reg1) {
		return errors.DisallowedOp1Register(int(dParams.reg1), state.byteCodePos)
	}
	offset := int64(binary.BigEndian.Uint64(param))
	regV := state.GetRegVAsS64(int(dParams.reg1), dParams.r1sz)
	if inStack, e := state.movXDO1GetAddr(regV, offset, dParams.opTy); e != nil {
		return e
	} else {
		val := state.GetRegVAsU64(int(source), dParams.destSz)
		return state.putValInStackWithSize(dParams.destSz, val, inStack)
	}
}
func (state *VmState) movXDO1GetAddr(r1V, off int64, opTy byte) (uint64, error) {
	var addr uint64
	switch opTy {
	case OP_TADD:
		addr = uint64(r1V + off)
	case OP_TSUB:
		addr = uint64(r1V - off)
	case OP_TMUL:
		addr = uint64(r1V * off)
	case OP_TDIV:
		addr = uint64(r1V / off)
	default:
		return 0, errors.BadOpcode(uint32(state.currentOpcode), state.byteCodePos)
	}
	return addr, nil
}
func (state *VmState) movIDO2NoOffset(byte2, byte3, byte4 byte, param []byte) error {
	dParams := state.getDerefParamsO2(byte2, byte3, byte4)
	if err := state.isMovXRO2Allowed(dParams); err != nil {
		return err
	}
	p := binary.BigEndian.Uint64(param)
	imm := p
	offset := int64(0)
	reg1V := state.GetRegVAsS64(int(dParams.reg1), dParams.r1_2sz)
	reg2V := state.GetRegVAsS64(int(dParams.reg2), dParams.r1_2sz)
	if inStack, e := state.movXDO2GetAddr(reg1V, reg2V, offset, dParams.opTy); e != nil {
		return e
	} else {
		return state.putValInStackWithSize(dParams.destSz, imm, inStack)
	}
}
func (state *VmState) movIDO2(byte2, byte3, byte4 byte, param []byte) error {
	dParams := state.getDerefParamsO2(byte2, byte3, byte4)
	if err := state.isMovXRO2Allowed(dParams); err != nil {
		return err
	}
	p := binary.BigEndian.Uint64(param)
	imm := uint64(p & 0x0000_0000_ffff_ffff)
	offset := int64((p & 0xffff_ffff_0000_0000) >> 32)
	reg1V := state.GetRegVAsS64(int(dParams.reg1), dParams.r1_2sz)
	reg2V := state.GetRegVAsS64(int(dParams.reg2), dParams.r1_2sz)
	if inStack, e := state.movXDO2GetAddr(reg1V, reg2V, offset, dParams.opTy); e != nil {
		return e
	} else {
		return state.putValInStackWithSize(dParams.destSz, imm, inStack)
	}
}

func (state *VmState) movRDO2(byte2, byte3, byte4 byte, param []byte) error {
	dParams := state.getDerefParamsO2(byte2, byte3, byte4)
	source := dParams.dest
	if err := state.isMovXRO2Allowed(dParams); err != nil {
		return err
	}
	if !IsMovFromRAllowed(source) {
		return errors.DisallowedSrcRegister(int(source), state.byteCodePos)
	}
	p := binary.BigEndian.Uint64(param)
	offset := int64(p)
	reg1V := state.GetRegVAsS64(int(dParams.reg1), dParams.r1_2sz)
	reg2V := state.GetRegVAsS64(int(dParams.reg2), dParams.r1_2sz)
	if inStack, e := state.movXDO2GetAddr(reg1V, reg2V, offset, dParams.opTy); e != nil {
		return e
	} else {
		val := state.GetRegVAsU64(int(source), dParams.destSz)
		return state.putValInStackWithSize(dParams.destSz, val, inStack)
	}
}

func (state *VmState) movXDO2GetAddr(r1V, r2V, off int64, opTy byte) (uint64, error) {
	var addr uint64
	switch opTy {
	case OP_TADD:
		addr = uint64(r1V + r2V + off)
	case OP_TSUB:
		addr = uint64(r1V + r2V - off)
	default:
		return 0, errors.BadOpcode(uint32(state.currentOpcode), state.byteCodePos)
	}
	return addr, nil
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
func (state *VmState) copyFromAddressToRegister(r *Register,
	addr uint64, dataSz byte) error {
	if inStack, e := state.isWithinStack(addr); e != nil {
		return e
	} else {
		bytes := DataSizeToByteCount(dataSz)
		if s, err := state.getStackSliceAt(inStack, bytes); err != nil {
			return err
		} else {
			copy(
				(*r)[8-bytes:],
				s,
			)
		}
	}
	return nil
}
func (state *VmState) copyFromRegisterToAddress(r *Register,
	addr uint64, dataSz byte) error {
	if inStack, e := state.isWithinStack(addr); e != nil {
		return e
	} else {
		bytes := DataSizeToByteCount(dataSz)
		if s, err := state.getStackSliceAt(inStack, bytes); err != nil {
			return err
		} else {
			copy(
				s,
				(*r)[8-bytes:],
			)
		}
	}
	return nil
}
func (state *VmState) xchgRR(
	lastByte byte,
	param []byte,
) error {
	var src, dest byte
	src |= (param[7] & 0xf0) >> 4
	dest |= (param[7] & 0x0f)
	sz := (lastByte & 0b0000_0011)
	if !IsMovIntoRAllowed(src) {
		return errors.DisallowedSrcRegister(int(src), state.byteCodePos)
	} else if !IsMovIntoRAllowed(dest) {
		return errors.DisallowedDestRegister(int(dest), state.byteCodePos)
	} else {
		tmp := Register{}
		// dest into tmp
		tmp.CopyFromRegisterWithSize(
			&state.regs.r[dest],
			sz,
		)
		// src into dest
		state.regs.r[dest].CopyFromRegisterWithSize(
			&state.regs.r[src],
			sz,
		)
		// tmp into src
		state.regs.r[src].CopyFromRegisterWithSize(
			&tmp,
			sz,
		)
	}
	return nil
}
func (state *VmState) xchgDR(
	lastByte byte,
	param []byte,
) error {
	dest := (lastByte & 0b0000_1111)
	dataSz := (lastByte & 0b0011_0000) >> 4
	if !IsMovIntoRAllowed(dest) {
		return errors.DisallowedDestRegister(int(dest), state.byteCodePos)
	}
	addr := binary.BigEndian.Uint64(param)
	tmp := Register{}
	tmp.CopyFromRegisterWithSize(
		&state.regs.r[dest],
		dataSz,
	)
	if err :=
		state.copyFromAddressToRegister(&state.regs.r[dest], addr, dataSz); err != nil {
		return err
	}
	return state.copyFromRegisterToAddress(&tmp, addr, dataSz)
}
func (state *VmState) xchgDRO1(
	byte3, byte4 byte,
	param []byte,
) error {
	dParams := state.getDerefParamsO1(byte3, byte4)
	if !IsMovIntoRAllowed(dParams.dest) {
		return errors.DisallowedDestRegister(int(dParams.dest), state.byteCodePos)
	}
	if !IsMovRRAllowed(dParams.reg1) {
		return errors.DisallowedOp2Register(int(dParams.reg1), state.byteCodePos)
	}
	regV := state.GetRegVAsS64(int(dParams.reg1), dParams.r1sz)
	offset := int64(binary.BigEndian.Uint64(param))
	if addr, e := state.movXDO1GetAddr(regV, offset, dParams.opTy); e != nil {
		return e
	} else {
		tmp := Register{}
		tmp.CopyFromRegisterWithSize(
			&state.regs.r[dParams.dest],
			dParams.destSz,
		)
		if err :=
			state.copyFromAddressToRegister(
				&state.regs.r[dParams.dest],
				addr,
				dParams.destSz,
			); err != nil {
			return err
		}
		return state.copyFromRegisterToAddress(
			&tmp,
			addr,
			dParams.destSz,
		)
	}
}
func (state *VmState) xchgDRO2(
	byte2, byte3, byte4 byte,
	param []byte,
) error {
	dParams := state.getDerefParamsO2(byte2, byte3, byte4)
	if !IsMovIntoRAllowed(dParams.dest) {
		return errors.DisallowedDestRegister(int(dParams.dest), state.byteCodePos)
	}
	if err := state.isMovXRO2Allowed(dParams); err != nil {
		return err
	}
	reg1V := int64(state.GetRegVAsU64(int(dParams.reg1), dParams.r1_2sz))
	reg2V := int64(state.GetRegVAsU64(int(dParams.reg2), dParams.r1_2sz))
	offset := int64(binary.BigEndian.Uint64(param))
	if addr, e := state.movXDO2GetAddr(reg1V, reg2V, offset, dParams.opTy); e != nil {
		return e
	} else {
		tmp := Register{}
		tmp.CopyFromRegisterWithSize(
			&state.regs.r[dParams.dest],
			dParams.destSz,
		)
		if err :=
			state.copyFromAddressToRegister(
				&state.regs.r[dParams.dest],
				addr,
				dParams.destSz,
			); err != nil {
			return err
		}
		return state.copyFromRegisterToAddress(
			&tmp,
			addr,
			dParams.destSz,
		)
	}
}
