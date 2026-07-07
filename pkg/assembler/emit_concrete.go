package assembler

import (
	"encoding/binary"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
	"github.com/JakubCygaro/alphataurus/pkg/vm/decls"
)

func (a *Assembler) emitPushR(data pr.InstPushR, at int) error {
	push := a.opCodes.GetBytes(vm.OP_PUSHR)
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(push))
	a.bytecode[at] = 0b0000_0011 & data.Reg.Size
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Reg.Reg))
	return nil
}
func (a *Assembler) emitPushI(data pr.InstPushI, at int) error {
	push := a.opCodes.GetBytes(vm.OP_PUSHI)
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(push))
	a.bytecode[at] = 0b0000_0011 & data.DataSz
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Imm))
	return nil
}
func (a *Assembler) emitPop(data pr.InstPop, at int) error {
	pop := a.opCodes.GetBytes(vm.OP_POP)
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(pop))
	a.bytecode[at] = 0b0000_0011 & data.DataSz
	return nil
}
func (a *Assembler) emitPopR(data pr.InstPopR, at int) error {
	pop := a.opCodes.GetBytes(vm.OP_POP)
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(pop))
	a.bytecode[at] = 0b0000_0011 & data.DataSz
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Reg))
	return nil
}
func (a *Assembler) emitNop(at int) error {
	nop := a.opCodes.GetBytes(vm.OP_NOP)
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(nop))
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(0))
	return nil
}
func (a *Assembler) emitClr(at int) error {
	clr := a.opCodes.GetBytes(vm.OP_CLR)
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(clr))
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(0))
	return nil
}
func (a *Assembler) emitCallIP0R(data pr.InstCallIP0R, at int) error {
	call := a.opCodes.GetBytes(vm.OP_CALLIP)
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
	call := a.opCodes.GetBytes(vm.OP_CALLIP)
	var reg byte
	reg = byte(data.OpTy)
	reg <<= 4
	if data.Reg.Size != vm.SZ_64 {
		return errors.BadRegisterSize(a.line, a.col)
	}
	reg |= byte(data.Reg.Reg & 0x0f)
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(call))
	a.bytecode[at] = reg
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Offset))
	return nil
}
func (a *Assembler) emitCall(data pr.InstGenericCall, at int) error {
	call := a.opCodes.GetBytes(vm.OP_CALL)
	//direct call case
	evaluated, err := a.ev.TryEvaluateExpression(data.Expr)
	if err != nil {
		return err
	} else if evaluated != nil {
		if addr, ok := pr.IsConstexprType[pr.ConstExprILit](evaluated); !ok {
			return errors.Expected(
				"Valid address",
				data.Expr.Line,
				data.Expr.Col,
			)
		} else {
			binary.BigEndian.PutUint32(a.bytecode[at:], uint32(call))
			binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(addr.Integer))
		}
	} else {
		position := at
		a.unresolvedJumps[position] = unresolvedJump{
			Expr:    data.Expr,
			PatchTy: PatchCall{},
		}
		binary.BigEndian.AppendUint32(a.bytecode[at:], uint32(a.opCodes.GetBytes(vm.OP_NOP)))
		binary.BigEndian.AppendUint64(a.bytecode[at+4:], 0)
	}
	return nil
}

func (a *Assembler) emitRet(at int) error {
	ret := a.opCodes.GetBytes(vm.OP_RET)
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(ret))
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(0))
	return nil
}
func (a *Assembler) emitExitI(data pr.InstExitI, at int) error {
	exit := a.opCodes.GetBytes(vm.OP_EXITI)
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(exit))
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Val))
	return nil
}
func (a *Assembler) emitExitR(data pr.InstExitR, at int) error {
	exit := a.opCodes.GetBytes(vm.OP_EXITR)
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
func (a *Assembler) emitJmp(data pr.InstGenericJmp, at int) error {
	opcode := a.jmpInstToOpCode(data.Variant)
	position := at
	evaluated, err := a.ev.TryEvaluateExpression(data.Address)
	if err != nil {
		return err
	} else if evaluated != nil {
		if addr, ok := pr.IsConstexprType[pr.ConstExprILit](evaluated); !ok {
			return errors.Expected(
				"Valid address",
				data.Address.Line,
				data.Address.Col,
			)
		} else {
			binary.BigEndian.PutUint32(a.bytecode[at:], uint32(opcode))
			binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(addr.Integer))
		}
	} else {
		a.unresolvedJumps[position] = unresolvedJump{
			Expr: data.Address,
			PatchTy: PatchJmp{
				Variant: data.Variant,
			},
			Absolute: data.Absolute,
		}
		binary.BigEndian.PutUint32(a.bytecode[at:],
			uint32(a.opCodes.GetBytes(vm.OP_NOP)))
		binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(0))
	}

	// else if addr, ok := cexpr.Val.(pr.ConstExprIden); ok {
	// 	a.unresolvedJumps[position] = unresolvedJump{
	// 		Ident: data.Address.(string),
	// 		PatchTy: PatchJmp{
	// 			Variant: data.Variant,
	// 		},
	// 		Absolute: data.Absolute,
	// 	}
	// 	*out = binary.BigEndian.AppendUint32(*out, uint32(a.opCodes.GetBytes(vm.OP_NOP)))
	// 	*out = binary.BigEndian.AppendUint64(*out, uint64(0))
	// }
	return nil
}
func (a *Assembler) emitMovIR(data pr.InstMovIR, at int) error {
	mov := a.opCodes.GetBytes(vm.OP_MOVIR)
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	lastByte :=
		(0b0000_1111 & byte(data.Dest.Reg)) |
			(0b0011_0000 & (byte(0) << 4)) |
			(0b1100_0000 & (byte(data.DataSize) << 6))
	a.bytecode[at] = lastByte
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Imm))
	return nil
}

func (a *Assembler) emitMovRR(data pr.InstMovRR, at int) error {
	mov := a.opCodes.GetBytes(vm.OP_MOVRR)
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	if !vm.IsMovIntoRAllowed(byte(data.Dest.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	if !vm.IsMovFromRAllowed(byte(data.Src.Reg)) {
		return errors.DisallowedSourceRegister(a.line, a.col)
	}
	destsrc := 0b00001111 & byte(data.Dest.Reg)
	destsrc |= (0b00001111 & byte(data.Src.Reg)) << 4
	dataSz := (0b0000_0011 & data.Dest.Size)
	a.bytecode[at] = dataSz
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(0))
	a.bytecode[at+decls.INSTRUCTION_SIZE-1] = destsrc
	return nil
}

func (a *Assembler) emitMovDRI(data pr.InstMovDR, at int) error {
	mov := a.opCodes.GetBytes(vm.OP_MOVDRI)
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	if !vm.IsMovIntoRAllowed(byte(data.Dest.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	lastByte := (0b0000_1111 & byte(data.Dest.Reg))
	lastByte |= (0b0000_0011 & data.Dest.Size) << 4
	a.bytecode[at] = lastByte
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Address))
	return nil
}
func (a *Assembler) emitMovDRO1(data pr.InstMovDRO1, at int) error {
	mov := a.opCodes.GetBytes(vm.OP_MOVDRO1)
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	if !vm.IsMovIntoRAllowed(byte(data.Dest.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	lastByte := (0b0000_1111 & byte(data.Dest.Reg)) << 4
	lastByte |= (0b0000_1111 & byte(data.OReg1.Reg))
	penultByte := (0b0000_0011 & byte(data.Dest.Size)) << 4
	penultByte |= (0b0000_0011 & byte(data.OReg1.Size)) << 2
	off, _ := lx.TokenTToOpT(data.OffOp)
	penultByte |= (0b0000_0011 & byte(off))
	a.bytecode[at] = lastByte
	a.bytecode[at+1] = penultByte
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Offset))
	return nil
}
func (a *Assembler) emitMovDRO2(data pr.InstMovDRO2, at int) error {
	mov := a.opCodes.GetBytes(vm.OP_MOVDRO2)
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	if !vm.IsMovIntoRAllowed(byte(data.Dest.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	byte4 := (0b0000_1111 & byte(data.Dest.Reg)) << 4
	byte4 |= (0b0000_1111 & byte(data.OReg1.Reg))
	byte3 := (0b0000_1111 & byte(data.OReg2.Reg)) << 4
	byte3 |= (0b0000_0011 & byte(data.OReg1.Size)) << 2
	off, _ := lx.TokenTToOpT(data.RegOp)
	byte3 |= (0b0000_0011 & byte(off))
	byte2 := (data.Dest.Size & 0b0000_0011)
	a.bytecode[at] = byte4
	a.bytecode[at+1] = byte3
	a.bytecode[at+2] = byte2
	param := uint64(data.Offset)
	binary.BigEndian.PutUint64(a.bytecode[at+4:], param)
	return nil
}
func (a *Assembler) emitMovID(data pr.InstMovID, at int) error {
	mov := a.opCodes.GetBytes(vm.OP_MOVID)
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	lastByte := 0b0000_0011 & data.DataSize
	a.bytecode[at] = lastByte
	binary.BigEndian.PutUint32(a.bytecode[at+4:], uint32(data.Address))
	binary.BigEndian.PutUint32(a.bytecode[at+8:], uint32(data.Imm))
	return nil
}
func (a *Assembler) emitMovRD(data pr.InstMovRD, at int) error {
	mov := a.opCodes.GetBytes(vm.OP_MOVRD)
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	if !vm.IsMovFromRAllowed(byte(data.Src.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	lastByte := 0b0000_1111 & byte(data.Src.Reg)
	lastByte |= (0b0000_0011 & data.Src.Size) << 4
	a.bytecode[at] = lastByte
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Address))
	return nil
}
func (a *Assembler) emitMovIDO1(data pr.InstMovIDO1, at int) error {
	lastByte := byte(0)
	lastByte |= (0b0000_1111 & byte(data.OReg1.Reg))
	penultByte := (0b0000_0011 & byte(data.DataSize)) << 4
	penultByte |= (0b0000_0011 & byte(data.OReg1.Size)) << 2
	off, _ := lx.TokenTToOpT(data.OffOp)
	penultByte |= (0b0000_0011 & byte(off))
	param := uint64(0)
	var mov uint32
	if data.NoOff {
		mov = a.opCodes.GetBytes(vm.OP_MOVIDO1_NO)
		param = data.Imm
	} else {
		mov = a.opCodes.GetBytes(vm.OP_MOVIDO1)
		param = uint64(data.Offset) << 32
		param |= 0x0000_0000_ffff_ffff & uint64(data.Imm)
	}
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	a.bytecode[at] = lastByte
	a.bytecode[at+1] = penultByte
	binary.BigEndian.PutUint64(a.bytecode[at+4:], param)
	return nil
}
func (a *Assembler) emitMovRDO1(data pr.InstMovRDO1, at int) error {
	mov := a.opCodes.GetBytes(vm.OP_MOVRDO1)
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	if !vm.IsMovFromRAllowed(byte(data.Src.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	lastByte := (0b0000_1111 & byte(data.Src.Reg)) << 4
	lastByte |= (0b0000_1111 & byte(data.OReg1.Reg))
	penultByte := (0b0000_0011 & byte(data.Src.Size)) << 4
	penultByte |= (0b0000_0011 & byte(data.OReg1.Size)) << 2
	off, _ := lx.TokenTToOpT(data.OffOp)
	penultByte |= (0b0000_0011 & byte(off))
	a.bytecode[at] = lastByte
	a.bytecode[at+1] = penultByte
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Offset))
	return nil
}
func (a *Assembler) emitMovIDO2(data pr.InstMovIDO2, at int) error {
	if data.OReg1.Size != data.OReg2.Size {
		return errors.MismatchedRegisterSizes(
			a.parser.CurrentInst().Line,
			a.parser.CurrentInst().Col,
		)
	}
	byte4 := (0b0000_1111 & byte(0)) << 4
	byte4 |= (0b0000_1111 & byte(data.OReg1.Reg))
	byte3 := (0b0000_1111 & byte(data.OReg2.Reg)) << 4
	byte3 |= (0b0000_0011 & byte(data.OReg1.Size)) << 2
	byte3 |= (0b0000_0011 & byte(data.RegOp))
	byte2 := (0b0000_0011 & byte(data.DataSize))
	var mov uint32
	var param uint64
	if data.NoOff {
		mov = a.opCodes.GetBytes(vm.OP_MOVIDO2_NO)
		param = uint64(data.Imm)
	} else {
		mov = a.opCodes.GetBytes(vm.OP_MOVIDO2)
		param |= uint64(data.Offset) << 32
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
	if data.OReg1.Size != data.OReg2.Size ||
		data.Src.Size < data.OReg1.Size {
		return errors.MismatchedRegisterSizes(
			a.parser.CurrentInst().Line,
			a.parser.CurrentInst().Col,
		)
	}
	if !vm.IsMovFromRAllowed(byte(data.Src.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	mov := a.opCodes.GetBytes(vm.OP_MOVRDO1)
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(mov))
	byte4 := (0b0000_1111 & byte(data.Src.Reg)) << 4
	byte4 |= (0b0000_1111 & byte(data.OReg1.Reg))
	byte3 := (0b0000_1111 & byte(data.OReg2.Reg)) << 4
	byte3 |= (0b0000_0011 & byte(data.OReg1.Size)) << 2
	byte3 |= (0b0000_0011 & byte(data.RegOp))
	byte2 := (0b0000_0011 & data.Src.Size)
	a.bytecode[at] = byte4
	a.bytecode[at+1] = byte3
	a.bytecode[at+2] = byte2
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Offset))
	return nil
}
func (a *Assembler) emitArthRR(data pr.InstArthRR, at int) error {
	var opCode uint32
	sized := false
	if !vm.IsArthRAllowed(byte(data.Src.Reg)) {
		return errors.DisallowedSourceRegister(a.line, a.col)
	}
	if !vm.IsArthRAllowed(byte(data.Dest.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	switch data.ArthTy {
	case pr.ADD:
		opCode = a.opCodes.GetBytes(vm.OP_ADDRR)
	case pr.SUB:
		opCode = a.opCodes.GetBytes(vm.OP_SUBRR)
	case pr.DIV:
		opCode = a.opCodes.GetBytes(vm.OP_DIVRR)
		sized = true
	case pr.MUL:
		opCode = a.opCodes.GetBytes(vm.OP_MULRR)
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
	if data.First.Size != data.Second.Size {
		return errors.MismatchedRegisterSizes(
			a.parser.CurrentInst().Line,
			a.parser.CurrentInst().Col,
		)
	}
	if !vm.IsArthRAllowed(byte(data.First.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	if !vm.IsLogRAllowed(byte(data.Second.Reg)) {
		return errors.DisallowedSourceRegister(a.line, a.col)
	}
	var opCode uint32
	switch data.LogTy {
	case pr.AND:
		opCode = a.opCodes.GetBytes(vm.OP_ANDRR)
	case pr.OR:
		opCode = a.opCodes.GetBytes(vm.OP_ORRR)
	case pr.XOR:
		opCode = a.opCodes.GetBytes(vm.OP_XORRR)
	case pr.LSH:
		opCode = a.opCodes.GetBytes(vm.OP_LSHRR)
	case pr.RSH:
		opCode = a.opCodes.GetBytes(vm.OP_RSHRR)
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
	if !vm.IsArthRAllowed(byte(data.First.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	var opCode uint32
	switch data.LogTy {
	case pr.AND:
		opCode = a.opCodes.GetBytes(vm.OP_ANDIR)
	case pr.OR:
		opCode = a.opCodes.GetBytes(vm.OP_ORIR)
	case pr.XOR:
		opCode = a.opCodes.GetBytes(vm.OP_XORIR)
	case pr.LSH:
		opCode = a.opCodes.GetBytes(vm.OP_LSHIR)
	case pr.RSH:
		opCode = a.opCodes.GetBytes(vm.OP_RSHIR)
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
	if !vm.IsArthRAllowed(byte(data.Dest.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	var opCode uint32
	switch data.ArthTy {
	case pr.ADD:
		opCode = a.opCodes.GetBytes(vm.OP_ADDIR)
	case pr.SUB:
		opCode = a.opCodes.GetBytes(vm.OP_SUBIR)
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
	opCode := a.opCodes.GetBytes(vm.OP_NOT)
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
func (a *Assembler) emitInc(data pr.InstInc, at int) error {
	if !vm.IsArthRAllowed(byte(data.Reg.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	inc := a.opCodes.GetBytes(vm.OP_INCR)
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(inc))
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Reg.Reg))
	return nil
}
func (a *Assembler) emitDec(data pr.InstDec, at int) error {
	if !vm.IsArthRAllowed(byte(data.Reg.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	dec := a.opCodes.GetBytes(vm.OP_DECR)
	binary.BigEndian.PutUint32(a.bytecode[at:], uint32(dec))
	binary.BigEndian.PutUint64(a.bytecode[at+4:], uint64(data.Reg.Reg))
	return nil
}
func (a *Assembler) emitCmpRR(data pr.InstCmpRR, at int) error {
	cmp := a.opCodes.GetBytes(vm.OP_CMPRR)
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
	cmp := a.opCodes.GetBytes(vm.OP_CMPIR)
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
