package assembler

import (
	"encoding/binary"

	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
	pe "github.com/JakubCygaro/alphataurus/pkg/assembler/parser/errors"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
	"github.com/JakubCygaro/alphataurus/pkg/vm/decls"
)

func (a *Assembler) emitPushR(data pr.InstPushR, at int) error {
	// push := vm.OP_PUSHR_VAL
	push := vm.OP_PUSHR_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(push))
	a.bytecode[at] = 0b0000_0011 & data.Reg.Size
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Reg.Reg))
	return nil
}
func (a *Assembler) emitPushI(data pr.InstPushI, at int) error {
	push := vm.OP_PUSHI_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(push))
	a.bytecode[at] = 0b0000_0011 & data.DataSz
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Imm))
	return nil
}
func (a *Assembler) emitPop(data pr.InstPop, at int) error {
	pop := vm.OP_POP_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(pop))
	a.bytecode[at] = 0b0000_0011 & data.DataSz
	return nil
}
func (a *Assembler) emitPopR(data pr.InstPopR, at int) error {
	pop := vm.OP_POP_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(pop))
	a.bytecode[at] = 0b0000_0011 & data.DataSz
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Reg))
	return nil
}
func (a *Assembler) emitNop(at int) error {
	nop := vm.OP_NOP_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(nop))
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(0))
	return nil
}
func (a *Assembler) emitClr(at int) error {
	clr := vm.OP_CLR_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(clr))
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(0))
	return nil
}
func (a *Assembler) emitSDF(at int) error {
	sdf := vm.OP_SDF_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(sdf))
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(0))
	return nil
}
func (a *Assembler) emitCDF(at int) error {
	cdf := vm.OP_CDF_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(cdf))
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(0))
	return nil
}
func (a *Assembler) emitCallI(data pr.InstCallI, at int) error {
	call := vm.OP_CALL_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(call))
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Address))
	return nil
}
func (a *Assembler) emitCallIP0R(data pr.InstCallIP0R, at int) error {
	call := vm.OP_CALLIP_VAL
	var reg byte
	reg = byte(data.OpTy)
	reg <<= 4
	reg |= 0x0f
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(call))
	a.bytecode[at] = reg
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Offset))
	return nil
}
func (a *Assembler) emitCallIP1R(data pr.InstCallIP1R, at int) error {
	call := vm.OP_CALLIP_VAL
	var reg byte
	reg = byte(data.OpTy)
	reg <<= 4
	reg |= byte(data.Reg.Reg & 0x0f)
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(call))
	a.bytecode[at] = reg
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Offset))
	return nil
}

func (a *Assembler) emitRet(at int) error {
	ret := vm.OP_RET_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(ret))
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(0))
	return nil
}
func (a *Assembler) emitExitI(data pr.InstExitI, at int) error {
	exit := vm.OP_EXITI_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(exit))
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Val))
	return nil
}
func (a *Assembler) emitExitR(data pr.InstExitR, at int) error {
	exit := vm.OP_EXITR_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(exit))
	param := [8]byte{
		byte(data.Reg.Reg),
		data.Reg.Size,
		0, 0, 0, 0, 0, 0,
	}
	binary.BigEndian.PutUint64(a.bytecode[at+4:],
		binary.BigEndian.Uint64(param[:]))
	return nil
}
func (a *Assembler) emitJmpIP0R(data pr.InstJmpIP0R, at int) error {
	opcode := a.absoluteJmpToIPJmp(data.JmpTy)
	var reg byte
	reg = byte(data.OpTy)
	reg <<= 4
	// no second offset register
	reg |= 0x0f
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(opcode))
	a.bytecode[at] = reg
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Offset))
	return nil
}
func (a *Assembler) emitJmpIP1R(data pr.InstJmpIP1R, at int) error {
	opcode := a.absoluteJmpToIPJmp(data.JmpTy)
	var reg byte
	reg = byte(data.OpTy)
	reg <<= 4
	//with second offset register
	reg |= byte(data.Reg.Reg & 0x0f)
	reg |= (0b0000_0011 & data.Reg.Size) << 6
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(opcode))
	a.bytecode[at] = reg
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Offset))
	return nil
}
func (a *Assembler) emitJmpI(data pr.InstJmpI, at int) error {
	opcode := a.jmpInstToOpCode(data.JmpTy)
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(opcode))
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Address))
	return nil
}
func (a *Assembler) emitMovIR(data pr.InstMovIR, at int, omitOpcode bool) error {
	if !omitOpcode {
		mov := vm.OP_MOVIR_VAL
		binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	}
	lastByte :=
		(0b0000_1111 & byte(data.Dest.Reg)) |
			(0b0011_0000 & (byte(0) << 4)) |
			(0b1100_0000 & (byte(data.DataSize) << 6))
	a.bytecode[at] = lastByte
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Imm))
	return nil
}
func (a *Assembler) emitMovZXIR(data pr.InstMovZXIR, at int) error {
	mov := vm.OP_MOVIR_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	err := a.emitMovIR(data.Mov, at, true)
	return err
}
func (a *Assembler) emitMovZXRR(data pr.InstMovZXRR, at int) error {
	mov := vm.OP_MOVZXRR_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	err := a.emitMovRR(data.Mov, at, true)
	return err
}
func (a *Assembler) emitMovSXRR(data pr.InstMovSXRR, at int) error {
	mov := vm.OP_MOVSXRR_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	err := a.emitMovRR(data.Mov, at, true)
	return err
}
func (a *Assembler) emitMovRR(data pr.InstMovRR, at int, omitOpcode bool) error {
	if !omitOpcode{
		mov := vm.OP_MOVRR_VAL
		binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	}
	if !vm.IsMovIntoRAllowed(byte(data.Dest.Reg)) {
		return pe.
			DisallowedDestReg(
				a.cInst.Line,
				a.cInst.Col,
				data.Dest,
			)
	}
	if !vm.IsMovFromRAllowed(byte(data.Src.Reg)) {
		return pe.
			DisallowedSrcReg(
				a.cInst.Line,
				a.cInst.Col,
				data.Src,
			)
	}
	destsrc := 0b00001111 & byte(data.Dest.Reg)
	destsrc |= (0b00001111 & byte(data.Src.Reg)) << 4
	dataSz := (0b0000_0011 & data.Dest.Size)
	a.bytecode[at] = dataSz
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(0))
	a.bytecode[at+decls.INSTRUCTION_SIZE-1] = destsrc
	return nil
}

func (a *Assembler) emitMovDR(data pr.InstMovDR, at int, omitOpcode bool) error {
	if !omitOpcode {
		mov := vm.OP_MOVDR_VAL
		binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	}
	if !vm.IsMovIntoRAllowed(byte(data.Dest.Reg)) {
		return pe.
			DisallowedDestReg(
				a.cInst.Line,
				a.cInst.Col,
				data.Dest,
			)
	}
	lastByte := (0b0000_1111 & byte(data.Dest.Reg))
	lastByte |= (0b0000_0011 & data.Dest.Size) << 4
	a.bytecode[at] = lastByte
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Address))
	return nil
}
func (a *Assembler) emitMovZXDR(data pr.InstMovZXDR, at int) error {
	mov := vm.OP_MOVZXDR_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	err := a.emitMovDR(data.Mov, at, true)
	return err
}
func (a *Assembler) emitMovDRO1(data pr.InstMovDRO1, at int, omitOpcode bool) error {
	if !omitOpcode {
		mov := vm.OP_MOVDRO1_VAL
		binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	}
	if !vm.IsMovIntoRAllowed(byte(data.Base.Reg)) {
		return pe.
			DisallowedDestReg(
				a.cInst.Line,
				a.cInst.Col,
				data.Base,
			)
	}
	lastByte := (0b0000_1111 & byte(data.Dest.Reg)) << 4
	lastByte |= (0b0000_1111 & byte(data.Base.Reg))
	penultByte := (0b0000_0011 & byte(data.Dest.Size)) << 4
	penultByte |= (0b0000_0011 & byte(data.Base.Size)) << 2
	penultByte |= (0b0000_0011 & byte(data.SF)) << 6
	off, ok := lx.TokenTToOpT(data.DispOp)
	if !ok {
		panic(
			"bad internal assembler state - displacement operator of invalid type")
	}
	penultByte |= (0b0000_0011 & byte(off))
	a.bytecode[at] = lastByte
	a.bytecode[at+1] = penultByte
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Offset))
	return nil
}
func (a *Assembler) emitMovZXDRO1(data pr.InstMovZXDRO1, at int) error {
	mov := vm.OP_MOVZXDRO1_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	err := a.emitMovDRO1(data.Mov, at, true)
	return err
}
func (a *Assembler) emitMovDRO2(data pr.InstMovDRO2, at int, omitOpcode bool) error {
	if !omitOpcode {
		mov := vm.OP_MOVDRO2_VAL
		binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	}
	if !vm.IsMovIntoRAllowed(byte(data.Dest.Reg)) {
		return pe.
			DisallowedDestReg(
				a.cInst.Line,
				a.cInst.Col,
				data.Dest,
			)
	}
	byte4 := (0b0000_1111 & byte(data.Dest.Reg)) << 4
	byte4 |= (0b0000_1111 & byte(data.Base.Reg))
	byte3 := (0b0000_1111 & byte(data.Index.Reg)) << 4
	byte3 |= (0b0000_0011 & byte(data.Base.Size)) << 2
	off, ok := lx.TokenTToOpT(data.DispOp)
	if !ok {
		panic(
			"bad internal assembler state - register operator of invalid type")
	}
	byte3 |= (0b0000_0011 & byte(off))
	byte2 := (data.Dest.Size & 0b0000_0011)
	byte2 |= (byte(data.SF) & 0b0000_0011) << 2
	a.bytecode[at] = byte4
	a.bytecode[at+1] = byte3
	a.bytecode[at+2] = byte2
	param := uint64(data.Disp)
	binary.BigEndian.PutUint64(a.bytecode[at+4:], param)
	return nil
}
func (a *Assembler) emitMovZXDRO2(data pr.InstMovZXDRO2, at int) error {
	mov := vm.OP_MOVZXDRO2_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	err := a.emitMovDRO2(data.Mov, at, true)
	return err
}
func (a *Assembler) emitMovID(data pr.InstMovID, at int) error {
	mov := vm.OP_MOVID_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	lastByte := 0b0000_0011 & data.DataSize
	a.bytecode[at] = lastByte
	binary.BigEndian.PutUint32(a.bytecode[at+4:], uint32(data.Address))
	binary.BigEndian.PutUint32(a.bytecode[at+8:], uint32(data.Imm))
	return nil
}
func (a *Assembler) emitMovRD(data pr.InstMovRD, at int) error {
	mov := vm.OP_MOVRD_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	if !vm.IsMovFromRAllowed(byte(data.Src.Reg)) {
		return pe.DisallowedSrcReg(a.cInst.Line, a.cInst.Col, data.Src)
	}
	lastByte := 0b0000_1111 & byte(data.Src.Reg)
	lastByte |= (0b0000_0011 & data.Src.Size) << 4
	a.bytecode[at] = lastByte
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Address))
	return nil
}
func (a *Assembler) emitMovIDO1(data pr.InstMovIDO1, at int) error {
	lastByte := byte(0)
	lastByte |= (0b0000_1111 & byte(data.Base.Reg))
	penultByte := (0b0000_0011 & byte(data.DataSize)) << 4
	penultByte |= (0b0000_0011 & byte(data.Base.Size)) << 2
	penultByte |= (0b0000_0011 & byte(data.SF)) << 6
	off, ok := lx.TokenTToOpT(data.DispOp)
	if !ok {
		panic(
			"bad internal assembler state - displacement operator of invalid type")
	}
	penultByte |= (0b0000_0011 & byte(off))
	param := uint64(0)
	var mov uint32
	if data.NoOff {
		mov = vm.OP_MOVIDO1_NO_VAL
		param = data.Imm
	} else {
		mov = vm.OP_MOVIDO1_VAL
		param = uint64(int64(data.Disp) << 32)
		param |= 0x0000_0000_ffff_ffff & uint64(int32(data.Imm))
	}
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	a.bytecode[at] = lastByte
	a.bytecode[at+1] = penultByte
	binary.BigEndian.PutUint64(a.bytecode[at+4:], param)
	return nil
}
func (a *Assembler) emitMovRDO1(data pr.InstMovRDO1, at int) error {
	mov := vm.OP_MOVRDO1_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	if !vm.IsMovFromRAllowed(byte(data.Src.Reg)) {
		return pe.DisallowedSrcReg(a.cInst.Line, a.cInst.Col, data.Src)
	}
	lastByte := (0b0000_1111 & byte(data.Src.Reg)) << 4
	lastByte |= (0b0000_1111 & byte(data.Base.Reg))
	penultByte := (0b0000_0011 & byte(data.Src.Size)) << 4
	penultByte |= (0b0000_0011 & byte(data.Base.Size)) << 2
	penultByte |= (0b0000_0011 & byte(data.SF)) << 6
	off, ok := lx.TokenTToOpT(data.DispOp)
	if !ok {
		panic(
			"bad internal assembler state - displacement operator of invalid type")
	}
	penultByte |= (0b0000_0011 & byte(off))
	a.bytecode[at] = lastByte
	a.bytecode[at+1] = penultByte
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Disp))
	return nil
}
func (a *Assembler) emitMovIDO2(data pr.InstMovIDO2, at int) error {
	byte4 := (0b0000_1111 & byte(0)) << 4
	byte4 |= (0b0000_1111 & byte(data.Base.Reg))
	byte3 := (0b0000_1111 & byte(data.Index.Reg)) << 4
	byte3 |= (0b0000_0011 & byte(data.Base.Size)) << 2
	off, ok := lx.TokenTToOpT(data.DispOp)
	if !ok {
		panic(
			"bad internal assembler state - register operator of invalid type")
	}
	byte3 |= (0b0000_0011 & byte(off))
	byte2 := (0b0000_0011 & byte(data.DataSize))
	byte2 |= (0b0000_0011 & byte(data.SF)) << 2
	var mov uint32
	var param uint64
	if data.NoOff {
		mov = vm.OP_MOVIDO2_NO_VAL
		param = uint64(data.Imm)
	} else {
		mov = vm.OP_MOVIDO2_VAL
		param |= uint64(data.Disp) << 32
		param |= 0x0000_0000_ffff_ffff & uint64(data.Imm)
	}
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	a.bytecode[at] = byte4
	a.bytecode[at+1] = byte3
	a.bytecode[at+2] = byte2
	binary.BigEndian.PutUint64(a.bytecode[at+4:], param)
	return nil
}
func (a *Assembler) emitMovRDO2(data pr.InstMovRDO2, at int) error {
	mov := vm.OP_MOVRDO1_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	byte4 := (0b0000_1111 & byte(data.Src.Reg)) << 4
	byte4 |= (0b0000_1111 & byte(data.Base.Reg))
	byte3 := (0b0000_1111 & byte(data.Index.Reg)) << 4
	byte3 |= (0b0000_0011 & byte(data.Base.Size)) << 2
	byte3 |= (0b0000_0011 & byte(data.DispOp))
	byte2 := (0b0000_0011 & data.Src.Size)
	byte2 |= (0b0000_0011 & byte(data.SF)) << 2
	a.bytecode[at] = byte4
	a.bytecode[at+1] = byte3
	a.bytecode[at+2] = byte2
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Disp))
	return nil
}
func (a *Assembler) emitArthRR(data pr.InstArthRR, at int) error {
	var opCode uint32
	sized := false
	switch data.ArthTy {
	case pr.ADD:
		opCode = vm.OP_ADDRR_VAL
	case pr.SUB:
		opCode = vm.OP_SUBRR_VAL
	case pr.DIV:
		opCode = vm.OP_DIVRR_VAL
		sized = true
	case pr.MUL:
		opCode = vm.OP_MULRR_VAL
		sized = true
	}
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(opCode))
	tySizesByte := (0b0000_0011 & byte(data.Ty))

	if sized {
		tySizesByte |= (0b000_0011 & byte(data.DataSize)) << 2
		tySizesByte |= (0b000_0011 & byte(data.DataSize)) << 4
	} else {
		tySizesByte |= (0b000_0011 & byte(data.Src.Size)) << 2
		tySizesByte |= (0b000_0011 & byte(data.Dest.Size)) << 4
	}
	param := [8]byte{
		byte(data.Src.Reg),
		byte(data.Dest.Reg),
		byte(tySizesByte),
		0,
		0,
		0,
		0,
		0,
	}
	binary.BigEndian.PutUint64(a.bytecode[at+4:],
		binary.BigEndian.Uint64(param[:]))
	return nil
}
func (a *Assembler) emitLogRR(data pr.InstLogicalRR, at int) error {
	var opCode uint32
	switch data.LogTy {
	case pr.AND:
		opCode = vm.OP_ANDRR_VAL
	case pr.OR:
		opCode = vm.OP_ORRR_VAL
	case pr.XOR:
		opCode = vm.OP_XORRR_VAL
	case pr.LSH:
		opCode = vm.OP_LSHRR_VAL
	case pr.RSH:
		opCode = vm.OP_RSHRR_VAL
	}
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(opCode))
	sizesByte := (0b0000_0011&byte(data.First.Size))<<2 |
		((0b000_0011 & byte(data.Second.Size)) << 4)
	param := [8]byte{
		byte(data.First.Reg),
		byte(data.Second.Reg),
		byte(sizesByte),
		0,
		0,
		0,
		0,
		0,
	}
	binary.BigEndian.PutUint64(a.bytecode[at+4:],
		binary.BigEndian.Uint64(param[:]))
	return nil
}
func (a *Assembler) emitLogIR(data pr.InstLogicalIR, at int) error {
	var opCode uint32
	switch data.LogTy {
	case pr.AND:
		opCode = vm.OP_ANDIR_VAL
	case pr.OR:
		opCode = vm.OP_ORIR_VAL
	case pr.XOR:
		opCode = vm.OP_XORIR_VAL
	case pr.LSH:
		opCode = vm.OP_LSHIR_VAL
	case pr.RSH:
		opCode = vm.OP_RSHIR_VAL
	}
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(opCode))
	lastByte := byte(0)
	lastByte |= 0b0000_1111 & byte(data.First.Reg)
	lastByte |= (0b0000_0011 & byte(data.First.Size)) << 4
	a.bytecode[at] = lastByte
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Imm))
	return nil
}
func (a *Assembler) emitArthIR(data pr.InstArthIR, at int) error {
	var opCode uint32
	switch data.ArthTy {
	case pr.ADD:
		opCode = vm.OP_ADDIR_VAL
	case pr.SUB:
		opCode = vm.OP_SUBIR_VAL
	}
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(opCode))
	lastByte := byte(0)
	lastByte |= 0b0000_1111 & byte(data.Dest.Reg)
	lastByte |= (0b0000_0011 & byte(data.Dest.Size)) << 4
	lastByte |= (0b0000_0011 & byte(data.Ty)) << 6
	a.bytecode[at] = lastByte
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Imm))
	return nil
}
func (a *Assembler) emitNot(data pr.InstNotR, at int) error {
	opCode := vm.OP_NOT_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(opCode))
	param := [8]byte{
		byte(data.First.Reg),
		data.First.Size,
		0, 0, 0, 0, 0, 0,
	}
	binary.BigEndian.PutUint64(a.bytecode[at+4:],
		binary.BigEndian.Uint64(param[:]))
	return nil
}
func (a *Assembler) emitRotate(data pr.InstRotate, at int) error {
	var opCode uint32
	if data.Left {
		opCode = vm.OP_ROL_VAL
	} else {
		opCode = vm.OP_ROR_VAL
	}
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(opCode))
	regSz := (byte(data.Arg.Reg) << 4)
	regSz |= (byte(data.Arg.Reg) & 0x0f)
	a.bytecode[at] = regSz
	binary.BigEndian.PutUint64(a.bytecode[at+4:], data.RotateBy)
	return nil
}
func (a *Assembler) emitNegR(data pr.InstNegR, at int) error {
	opCode := vm.OP_NEG_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(opCode))
	ty := byte(vm.TY_SINT)
	if data.Float {
		ty = vm.TY_FLOAT
	}
	param := [8]byte{
		byte(data.Arg.Reg),
		data.Arg.Size,
		ty,
		0, 0, 0, 0, 0,
	}
	binary.BigEndian.PutUint64(a.bytecode[at+4:],
		binary.BigEndian.Uint64(param[:]))
	return nil
}
func (a *Assembler) emitInc(data pr.InstInc, at int) error {
	inc := vm.OP_INCR_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(inc))
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Reg.Reg))
	return nil
}
func (a *Assembler) emitDec(data pr.InstDec, at int) error {
	dec := vm.OP_DECR_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(dec))
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Reg.Reg))
	return nil
}
func (a *Assembler) emitCmpRR(data pr.InstCmpRR, at int) error {
	cmp := vm.OP_CMPRR_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(cmp))
	regs := (0b0000_1111 & byte(data.Sub.Reg)) << 4
	regs |= 0b0000_1111 & byte(data.Min.Reg)
	// regs |= 0b0000_0011 & byte(data.Min.Size) << 2
	a.bytecode[at] = regs
	param := [8]byte{byte(data.Ty), byte(data.Min.Size), 0, 0, 0, 0, 0, 0}
	binary.BigEndian.PutUint64(a.bytecode[at+4:],
		binary.BigEndian.Uint64(param[:]))
	return nil
}
func (a *Assembler) emitCmpIR(data pr.InstCmpIR, at int) error {
	cmp := vm.OP_CMPIR_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(cmp))
	// subtrahend |= (lastByte & 0xf0) >> 4
	// minuend |= (lastByte & 0x0f)
	regs := (0b00001111 & byte(data.Min.Reg)) << 4
	regs |= (0b00000011 & byte(data.Min.Size)) << 2
	regs |= (0b0000_0011 & byte(data.Ty))
	a.bytecode[at] = regs
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Imm))
	return nil
}
func (a *Assembler) emitXCHGRR(data pr.InstXCHGRR, at int) error {
	xchg := vm.OP_XCHGRR_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(xchg))
	err := a.emitMovRR(data.Mov, at, true)
	return err
}
func (a *Assembler) emitXCHGDR(data pr.InstXCHGDR, at int) error {
	mov := vm.OP_XCHGDR_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	err := a.emitMovDR(data.Mov, at, true)
	return err
}
func (a *Assembler) emitXCHGDRO1(data pr.InstXCHGDRO1, at int) error {
	mov := vm.OP_XCHGDRO1_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	err := a.emitMovDRO1(data.Mov, at, true)
	return err
}
func (a *Assembler) emitXCHGDRO2(data pr.InstXCHGDRO2, at int) error {
	mov := vm.OP_XCHGDRO2_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	err := a.emitMovDRO2(data.Mov, at, true)
	return err
}

func (a *Assembler) emitMovSB(data pr.InstMovSB, at int) error {
	mov := vm.OP_MOVSB_VAL
	if data.Rep {
		mov = vm.OP_MOVREPSB_VAL
	}
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(0))
	return nil
}
func (a *Assembler) emitMovSQ(data pr.InstMovSQ, at int) error {
	mov := vm.OP_MOVSQ_VAL
	if data.Rep {
		mov = vm.OP_MOVREPSQ_VAL
	}
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(0))
	return nil
}
func (a *Assembler) emitMovSH(data pr.InstMovSH, at int) error {
	mov := vm.OP_MOVSH_VAL
	if data.Rep {
		mov = vm.OP_MOVREPSH_VAL
	}
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(0))
	return nil
}
func (a *Assembler) emitMovSW(data pr.InstMovSW, at int) error {
	mov := vm.OP_MOVSW_VAL
	if data.Rep {
		mov = vm.OP_MOVREPSW_VAL
	}
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(0))
	return nil
}
func (a *Assembler) emitLeaO1(data pr.InstLeaO1, at int) error {
	lea := vm.OP_LEADRO1_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(lea))
	err := a.emitMovDRO1(data.Mov, at, true)
	return err
}
func (a *Assembler) emitLeaO2(data pr.InstLeaO2, at int) error {
	lea := vm.OP_LEADRO2_VAL
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(lea))
	err := a.emitMovDRO2(data.Mov, at, true)
	return err
}
