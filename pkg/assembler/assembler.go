package assembler

import (
	"bufio"
	"encoding/binary"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
	decls "github.com/JakubCygaro/alphataurus/pkg/vm/decls"
	aobj "github.com/JakubCygaro/alphataurus/pkg/vm/obj"
)

type PatchCall void
type PatchJmp struct {
	Variant JmpVariant
}

type unresolvedJump struct {
	Ident string
	// what instruction is gonna get patched, type of Patch... struct
	PatchTy  any
	Absolute bool
}
type unresolvedJumpMap map[int]unresolvedJump

type AssemblerWarningData struct {
	Line, Col int
	Message   string
}

type Assembler struct {
	parser          *Parser
	opCodes         vm.OpCodeMap
	unresolvedJumps unresolvedJumpMap
	//relating to the current instruction
	line, col int
	// labels          labelMap
	lastInst    Instruction
	bytecode    []byte
	instCount   int
	symbols     aobj.SymbolTable
	relocations aobj.RelocationTable
	hasEntry    bool
	entry       uint64
	// pointer to a function that recieves warnings emitted by the assembler
	WarningSink func(AssemblerWarningData)
}

func (a *Assembler) InstructionCount() int {
	return a.instCount
}
func aInitialState() Assembler {
	a := Assembler{
		opCodes:         vm.GenerateOpcodeMap(),
		unresolvedJumps: make(unresolvedJumpMap),
		bytecode:        make([]byte, 0, 64),
		symbols:         aobj.NewSymbolTable(),
		relocations:     make(aobj.RelocationTable, 0),
		hasEntry:        false,
	}
	return a
}
func (a *Assembler) clearState() {
	clear(a.unresolvedJumps)
	clear(a.bytecode)
	clear(a.bytecode)
	a.symbols.Clear()
	clear(a.relocations)
	a.hasEntry = false
}
func NewAssembler(reader *bufio.Reader) *Assembler {
	a := aInitialState()
	a.parser = NewParser(reader)
	a.parser.WarningSink = a.parserWarningHandler
	return &a
}

// reset the state of the assembler and load new reader input
func (a *Assembler) LoadNew(reader *bufio.Reader) {
	p := a.parser
	p.LoadNew(reader)
	ws := a.WarningSink
	a.clearState()
	a.parser = p
	a.WarningSink = ws
}

func (a *Assembler) currentCodePos() (byte uint64, address uint64) {
	position := uint64(len(a.bytecode))
	posAsInstAddr := uint64(position + vm.ADDRESSDEADZONE_SIZE)
	return position, posAsInstAddr
}

func (a *Assembler) EmitBytecode() (int, error) {
	instCount := 0
	var ok bool
	var err error = nil
	ok, err = a.parser.ParseNext()
	for ; ok && err == nil; ok, err = a.parser.ParseNext() {
		inst := a.parser.CurrentInst()
		a.col, a.line = inst.Col, inst.Line
		switch i := inst.Data.(type) {
		case InstMovIR:
			err = a.emitMovIR(i, &(a.bytecode))
		case InstMovRR:
			err = a.emitMovRR(i, &(a.bytecode))
		case InstMovDRI:
			err = a.emitMovDRI(i, &(a.bytecode))
		case InstMovDRO1:
			err = a.emitMovDRO1(i, &(a.bytecode))
		case InstMovDRO2:
			err = a.emitMovDRO2(i, &(a.bytecode))
		case InstMovID:
			err = a.emitMovID(i, &(a.bytecode))
		case InstMovRD:
			err = a.emitMovRD(i, &(a.bytecode))
		case InstMovIDO1:
			err = a.emitMovIDO1(i, &(a.bytecode))
		case InstMovRDO1:
			err = a.emitMovRDO1(i, &(a.bytecode))
		case InstMovIDO2:
			err = a.emitMovIDO2(i, &(a.bytecode))
		case InstMovRDO2:
			err = a.emitMovRDO2(i, &(a.bytecode))
		case InstArthRR:
			err = a.emitArthRR(i, &(a.bytecode))
		case InstArthIR:
			err = a.emitArthIR(i, &(a.bytecode))
		case InstNot:
			err = a.emitNot(i, &(a.bytecode))
		case InstLogicalRR:
			err = a.emitLogRR(i, &(a.bytecode))
		case InstLogicalIR:
			err = a.emitLogIR(i, &(a.bytecode))
		case InstInc:
			err = a.emitInc(i, &(a.bytecode))
		case InstDec:
			err = a.emitDec(i, &(a.bytecode))
		case InstCmpRR:
			err = a.emitCmpRR(i, &(a.bytecode))
		case InstCmpIR:
			err = a.emitCmpIR(i, &(a.bytecode))
		case InstJmp:
			err = a.emitJmp(i, &(a.bytecode))
		case InstJmpIP0R:
			err = a.emitJmpIP0R(i, &(a.bytecode))
		case InstJmpIP1R:
			err = a.emitJmpIP1R(i, &(a.bytecode))
		case InstLab:
			err = a.declareLabel(i)
			instCount--
		case InstPushR:
			err = a.emitPushR(i, &(a.bytecode))
		case InstPushI:
			err = a.emitPushI(i, &(a.bytecode))
		case InstPop:
			err = a.emitPop(i, &(a.bytecode))
		case InstNop:
			err = a.emitNop(&(a.bytecode))
		case InstCall:
			err = a.emitCall(i, &(a.bytecode))
		case InstCallIP0R:
			err = a.emitCallIP0R(i, &(a.bytecode))
		case InstCallIP1R:
			err = a.emitCallIP1R(i, &(a.bytecode))
		case InstRet:
			err = a.emitRet(&(a.bytecode))
		case InstExitI:
			err = a.emitExitI(i, &(a.bytecode))
		case InstExitR:
			err = a.emitExitR(i, &(a.bytecode))
		case InstEntry:
			if a.hasEntry {
				err = errors.MultipleEntry(inst.Line, inst.Col)
			} else {
				a.hasEntry = true
				_, ent := a.currentCodePos()
				a.entry = ent
			}
		case InstClr:
			err = a.emitClr(&(a.bytecode))
		default:
			a.lastInst = inst
			return instCount, err
		}
		if err != nil {
			return instCount, err
		}
		instCount++
	}
	return instCount, err
}

func (a *Assembler) emitMovIR(data InstMovIR, out *[]byte) error {
	mov := a.opCodes.GetBytes(vm.OP_MOVIR)
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	lastByte :=
		(0b0000_1111 & byte(data.Dest)) |
			(0b0011_0000 & (byte(0) << 4)) |
			(0b1100_0000 & (byte(data.DataSize) << 6))
	(*out)[len(*out)-4] = lastByte
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Imm))
	return nil
}

func (a *Assembler) emitMovRR(data InstMovRR, out *[]byte) error {
	mov := a.opCodes.GetBytes(vm.OP_MOVRR)
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	if !vm.IsMovIntoRAllowed(byte(data.Dest)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	if !vm.IsMovFromRAllowed(byte(data.Src)) {
		return errors.DisallowedSourceRegister(a.line, a.col)
	}
	destsrc := 0b00001111 & byte(data.Dest)
	destsrc |= (0b00001111 & byte(data.Src)) << 4
	dataSz := (0b0000_0011 & data.DataSize)
	(*out)[len(*out)-4] = dataSz
	*out = binary.BigEndian.AppendUint64(*out, uint64(0))
	(*out)[len(*out)-1] = destsrc
	return nil
}

func (a *Assembler) emitMovDRI(data InstMovDRI, out *[]byte) error {
	mov := a.opCodes.GetBytes(vm.OP_MOVDRI)
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	if !vm.IsMovIntoRAllowed(byte(data.Dest.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	lastByte := (0b0000_1111 & byte(data.Dest.Reg))
	lastByte |= (0b0000_0011 & data.Dest.Size) << 4
	(*out)[len(*out)-4] = lastByte
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Offset))
	return nil
}
func (a *Assembler) emitMovDRO1(data InstMovDRO1, out *[]byte) error {
	mov := a.opCodes.GetBytes(vm.OP_MOVDRO1)
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	if !vm.IsMovIntoRAllowed(byte(data.Dest.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	lastByte := (0b0000_1111 & byte(data.Dest.Reg)) << 4
	lastByte |= (0b0000_1111 & byte(data.OReg1.Reg))
	penultByte := (0b0000_0011 & byte(data.Dest.Size)) << 4
	penultByte |= (0b0000_0011 & byte(data.OReg1.Size)) << 2
	penultByte |= (0b0000_0011 & byte(data.OpTy))
	(*out)[len(*out)-4] = lastByte
	(*out)[len(*out)-3] = penultByte
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Offset))
	return nil
}
func (a *Assembler) emitMovDRO2(data InstMovDRO2, out *[]byte) error {
	mov := a.opCodes.GetBytes(vm.OP_MOVDRO2)
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	if !vm.IsMovIntoRAllowed(byte(data.Dest.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	byte4 := (0b0000_1111 & byte(data.Dest.Reg)) << 4
	byte4 |= (0b0000_1111 & byte(data.OReg1.Reg))
	byte3 := (0b0000_1111 & byte(data.OReg2.Reg)) << 4
	byte3 |= (0b0000_0011 & byte(data.OReg1.Size)) << 2
	byte3 |= (0b0000_0011 & byte(data.OpTy))
	byte2 := (data.Dest.Size & 0b0000_0011)
	(*out)[len(*out)-4] = byte4
	(*out)[len(*out)-3] = byte3
	(*out)[len(*out)-2] = byte2
	param := uint64(data.Offset)
	*out = binary.BigEndian.AppendUint64(*out, param)
	return nil
}
func (a *Assembler) emitMovID(data InstMovID, out *[]byte) error {
	mov := a.opCodes.GetBytes(vm.OP_MOVID)
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	lastByte := 0b0000_0011 & data.DataSize
	(*out)[len(*out)-4] = lastByte
	*out = binary.BigEndian.AppendUint32(*out, uint32(data.Offset))
	*out = binary.BigEndian.AppendUint32(*out, uint32(data.Imm))
	return nil
}
func (a *Assembler) emitMovRD(data InstMovRD, out *[]byte) error {
	mov := a.opCodes.GetBytes(vm.OP_MOVRD)
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	if !vm.IsMovFromRAllowed(byte(data.Src.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	lastByte := 0b0000_1111 & byte(data.Src.Reg)
	lastByte |= (0b0000_0011 & data.Src.Size) << 4
	(*out)[len(*out)-4] = lastByte
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Offset))
	return nil
}
func (a *Assembler) emitMovIDO1(data InstMovIDO1, out *[]byte) error {
	lastByte := byte(0)
	lastByte |= (0b0000_1111 & byte(data.OReg1.Reg))
	penultByte := (0b0000_0011 & byte(data.DataSize)) << 4
	penultByte |= (0b0000_0011 & byte(data.OReg1.Size)) << 2
	penultByte |= (0b0000_0011 & byte(data.OpTy))
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
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	(*out)[len(*out)-4] = lastByte
	(*out)[len(*out)-3] = penultByte
	*out = binary.BigEndian.AppendUint64(*out, param)
	return nil
}
func (a *Assembler) emitMovRDO1(data InstMovRDO1, out *[]byte) error {
	mov := a.opCodes.GetBytes(vm.OP_MOVRDO1)
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	if !vm.IsMovFromRAllowed(byte(data.Src.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	lastByte := (0b0000_1111 & byte(data.Src.Reg)) << 4
	lastByte |= (0b0000_1111 & byte(data.OReg1.Reg))
	penultByte := (0b0000_0011 & byte(data.Src.Size)) << 4
	penultByte |= (0b0000_0011 & byte(data.OReg1.Size)) << 2
	penultByte |= (0b0000_0011 & byte(data.OpTy))
	(*out)[len(*out)-4] = lastByte
	(*out)[len(*out)-3] = penultByte
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Offset))
	return nil
}
func (a *Assembler) emitMovIDO2(data InstMovIDO2, out *[]byte) error {
	if data.OReg1.Size != data.OReg2.Size {
		return errors.MismatchedRegisterSizes(
			a.parser.currentInst.Line,
			a.parser.currentInst.Col,
		)
	}
	byte4 := (0b0000_1111 & byte(0)) << 4
	byte4 |= (0b0000_1111 & byte(data.OReg1.Reg))
	byte3 := (0b0000_1111 & byte(data.OReg2.Reg)) << 4
	byte3 |= (0b0000_0011 & byte(data.OReg1.Size)) << 2
	byte3 |= (0b0000_0011 & byte(data.OpTy))
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
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	(*out)[len(*out)-4] = byte4
	(*out)[len(*out)-3] = byte3
	(*out)[len(*out)-2] = byte2
	*out = binary.BigEndian.AppendUint64(*out, param)
	return nil
}
func (a *Assembler) emitMovRDO2(data InstMovRDO2, out *[]byte) error {
	if data.OReg1.Size != data.OReg2.Size ||
		data.Src.Size < data.OReg1.Size {
		return errors.MismatchedRegisterSizes(
			a.parser.currentInst.Line,
			a.parser.currentInst.Col,
		)
	}
	if !vm.IsMovFromRAllowed(byte(data.Src.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	mov := a.opCodes.GetBytes(vm.OP_MOVRDO1)
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	byte4 := (0b0000_1111 & byte(data.Src.Reg)) << 4
	byte4 |= (0b0000_1111 & byte(data.OReg1.Reg))
	byte3 := (0b0000_1111 & byte(data.OReg2.Reg)) << 4
	byte3 |= (0b0000_0011 & byte(data.OReg1.Size)) << 2
	byte3 |= (0b0000_0011 & byte(data.OpTy))
	byte2 := (0b0000_0011 & data.Src.Size)
	(*out)[len(*out)-4] = byte4
	(*out)[len(*out)-3] = byte3
	(*out)[len(*out)-2] = byte2
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Offset))
	return nil
}
func (a *Assembler) emitArthRR(data InstArthRR, out *[]byte) error {
	var opCode uint32
	sized := false
	if !vm.IsArthRAllowed(byte(data.Src.Reg)) {
		return errors.DisallowedSourceRegister(a.line, a.col)
	}
	if !vm.IsArthRAllowed(byte(data.Dest.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	switch data.ArthTy {
	case ADD:
		opCode = a.opCodes.GetBytes(vm.OP_ADDRR)
	case SUB:
		opCode = a.opCodes.GetBytes(vm.OP_SUBRR)
	case DIV:
		opCode = a.opCodes.GetBytes(vm.OP_DIVRR)
		sized = true
	case MUL:
		opCode = a.opCodes.GetBytes(vm.OP_MULRR)
		sized = true
	}
	*out = binary.BigEndian.AppendUint32(*out, uint32(opCode))
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
	*out = append(*out, param[:]...)
	return nil
}
func (a *Assembler) emitLogRR(data InstLogicalRR, out *[]byte) error {
	if data.First.Size != data.Second.Size {
		return errors.MismatchedRegisterSizes(
			a.parser.currentInst.Line,
			a.parser.currentInst.Col,
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
	case AND:
		opCode = a.opCodes.GetBytes(vm.OP_ANDRR)
	case OR:
		opCode = a.opCodes.GetBytes(vm.OP_ORRR)
	case XOR:
		opCode = a.opCodes.GetBytes(vm.OP_XORRR)
	case LSH:
		opCode = a.opCodes.GetBytes(vm.OP_LSHRR)
	case RSH:
		opCode = a.opCodes.GetBytes(vm.OP_RSHRR)
	}
	*out = binary.BigEndian.AppendUint32(*out, uint32(opCode))
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
	*out = append(*out, param[:]...)
	return nil
}
func (a *Assembler) emitLogIR(data InstLogicalIR, out *[]byte) error {
	if !vm.IsArthRAllowed(byte(data.First.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	var opCode uint32
	switch data.LogTy {
	case AND:
		opCode = a.opCodes.GetBytes(vm.OP_ANDIR)
	case OR:
		opCode = a.opCodes.GetBytes(vm.OP_ORIR)
	case XOR:
		opCode = a.opCodes.GetBytes(vm.OP_XORIR)
	case LSH:
		opCode = a.opCodes.GetBytes(vm.OP_LSHIR)
	case RSH:
		opCode = a.opCodes.GetBytes(vm.OP_RSHIR)
	}
	*out = binary.BigEndian.AppendUint32(*out, uint32(opCode))
	lastByte := byte(0)
	lastByte |= 0b0000_1111 & byte(data.First.Reg)
	lastByte |= (0b0000_0011 & byte(data.First.Size)) << 4
	(*out)[len(*out)-4] = lastByte
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Imm))
	return nil
}
func (a *Assembler) emitArthIR(data InstArthIR, out *[]byte) error {
	if !vm.IsArthRAllowed(byte(data.Dest.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	var opCode uint32
	switch data.ArthTy {
	case ADD:
		opCode = a.opCodes.GetBytes(vm.OP_ADDIR)
	case SUB:
		opCode = a.opCodes.GetBytes(vm.OP_SUBIR)
	}
	*out = binary.BigEndian.AppendUint32(*out, uint32(opCode))
	lastByte := byte(0)
	lastByte |= 0b0000_1111 & byte(data.Dest.Reg)
	lastByte |= (0b0000_0011 & byte(data.Dest.Size)) << 4
	lastByte |= (0b0000_0011 & byte(data.Ty)) << 6
	(*out)[len(*out)-4] = lastByte
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Imm))
	return nil
}
func (a *Assembler) emitNot(data InstNot, out *[]byte) error {
	opCode := a.opCodes.GetBytes(vm.OP_NOT)
	*out = binary.BigEndian.AppendUint32(*out, uint32(opCode))
	param := [8]byte{
		byte(data.First.Reg),
		data.First.Size,
		0, 0, 0, 0, 0, 0,
	}
	*out = append(*out, param[:]...)
	return nil
}
func (a *Assembler) emitInc(data InstInc, out *[]byte) error {
	if !vm.IsArthRAllowed(byte(data.Reg.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	inc := a.opCodes.GetBytes(vm.OP_INCR)
	*out = binary.BigEndian.AppendUint32(*out, uint32(inc))
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Reg.Reg))
	return nil
}
func (a *Assembler) emitDec(data InstDec, out *[]byte) error {
	if !vm.IsArthRAllowed(byte(data.Reg.Reg)) {
		return errors.DisallowedDestinationRegister(a.line, a.col)
	}
	dec := a.opCodes.GetBytes(vm.OP_DECR)
	*out = binary.BigEndian.AppendUint32(*out, uint32(dec))
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Reg.Reg))
	return nil
}
func (a *Assembler) emitCmpRR(data InstCmpRR, out *[]byte) error {
	cmp := a.opCodes.GetBytes(vm.OP_CMPRR)
	*out = binary.BigEndian.AppendUint32(*out, uint32(cmp))
	regs := (0b0000_1111 & byte(data.Sub.Reg)) << 4
	regs |= 0b0000_1111 & byte(data.Min.Reg)
	// regs |= 0b0000_0011 & byte(data.Min.Size) << 2
	(*out)[len(*out)-4] = regs
	*out = append(*out, byte(data.Ty), byte(data.Min.Size), 0, 0, 0, 0, 0, 0)
	return nil
}
func (a *Assembler) emitCmpIR(data InstCmpIR, out *[]byte) error {
	cmp := a.opCodes.GetBytes(vm.OP_CMPIR)
	*out = binary.BigEndian.AppendUint32(*out, uint32(cmp))
	// subtrahend |= (lastByte & 0xf0) >> 4
	// minuend |= (lastByte & 0x0f)
	regs := (0b00001111 & byte(data.Min.Reg)) << 4
	regs |= (0b00000011 & byte(data.Min.Size)) << 2
	regs |= (0b0000_0011 & byte(data.Ty))
	(*out)[len(*out)-4] = regs
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Imm))
	return nil
}
func (a *Assembler) jmpInstToOpCode(ty JmpVariant) uint32 {
	var opcode uint32
	switch ty {
	case JMP:
		opcode = a.opCodes.GetBytes(vm.OP_JMP)
	case JMPE:
		opcode = a.opCodes.GetBytes(vm.OP_JMPE)
	case JMPNE:
		opcode = a.opCodes.GetBytes(vm.OP_JMPNE)
	case JMPZ:
		opcode = a.opCodes.GetBytes(vm.OP_JMPZ)
	case JMPNZ:
		opcode = a.opCodes.GetBytes(vm.OP_JMPNZ)
	case JMPG:
		opcode = a.opCodes.GetBytes(vm.OP_JMPG)
	case JMPGE:
		opcode = a.opCodes.GetBytes(vm.OP_JMPGE)
	case JMPL:
		opcode = a.opCodes.GetBytes(vm.OP_JMPL)
	case JMPLE:
		opcode = a.opCodes.GetBytes(vm.OP_JMPLE)
	case JMPS:
		opcode = a.opCodes.GetBytes(vm.OP_JMPS)
	case JMPNS:
		opcode = a.opCodes.GetBytes(vm.OP_JMPNS)
	case JMPC:
		opcode = a.opCodes.GetBytes(vm.OP_JMPC)
	case JMPNC:
		opcode = a.opCodes.GetBytes(vm.OP_JMPNC)
	case JMPO:
		opcode = a.opCodes.GetBytes(vm.OP_JMPO)
	case JMPNO:
		opcode = a.opCodes.GetBytes(vm.OP_JMPNO)
	}
	return opcode
}
func (a *Assembler) emitJmp(data InstJmp, out *[]byte) error {
	opcode := a.jmpInstToOpCode(data.Variant)
	position := len(*out)
	if addr, ok := data.Address.(uint64); ok {
		*out = binary.BigEndian.AppendUint32(*out, uint32(opcode))
		*out = binary.BigEndian.AppendUint64(*out, uint64(addr))
	} else {
		a.unresolvedJumps[position] = unresolvedJump{
			Ident: data.Address.(string),
			PatchTy: PatchJmp{
				Variant: data.Variant,
			},
			Absolute: data.Absolute,
		}
		*out = binary.BigEndian.AppendUint32(*out, uint32(a.opCodes.GetBytes(vm.OP_NOP)))
		*out = binary.BigEndian.AppendUint64(*out, uint64(0))
	}
	return nil
}

func (a *Assembler) absoluteJmpToIPJmp(instTy JmpVariant) (opcode uint32) {
	switch instTy {
	case JMP:
		opcode = a.opCodes.GetBytes(vm.OP_JMPIP)
	case JMPE:
		opcode = a.opCodes.GetBytes(vm.OP_JMPEIP)
	case JMPNE:
		opcode = a.opCodes.GetBytes(vm.OP_JMPNEIP)
	case JMPZ:
		opcode = a.opCodes.GetBytes(vm.OP_JMPZIP)
	case JMPNZ:
		opcode = a.opCodes.GetBytes(vm.OP_JMPNZIP)
	case JMPG:
		opcode = a.opCodes.GetBytes(vm.OP_JMPGIP)
	case JMPGE:
		opcode = a.opCodes.GetBytes(vm.OP_JMPGEIP)
	case JMPL:
		opcode = a.opCodes.GetBytes(vm.OP_JMPLIP)
	case JMPLE:
		opcode = a.opCodes.GetBytes(vm.OP_JMPLEIP)
	case JMPS:
		opcode = a.opCodes.GetBytes(vm.OP_JMPSIP)
	case JMPNS:
		opcode = a.opCodes.GetBytes(vm.OP_JMPNSIP)
	case JMPC:
		opcode = a.opCodes.GetBytes(vm.OP_JMPCIP)
	case JMPNC:
		opcode = a.opCodes.GetBytes(vm.OP_JMPNCIP)
	case JMPO:
		opcode = a.opCodes.GetBytes(vm.OP_JMPOIP)
	case JMPNO:
		opcode = a.opCodes.GetBytes(vm.OP_JMPNOIP)
	}
	return opcode
}

func (a *Assembler) emitJmpIP0R(data InstJmpIP0R, out *[]byte) error {
	opcode := a.absoluteJmpToIPJmp(data.JmpTy)
	var reg byte
	reg = byte(data.OpTy)
	reg <<= 4
	// no second offset register
	reg |= 0x0f
	*out = binary.BigEndian.AppendUint32(*out, uint32(opcode))
	(*out)[len(*out)-4] = reg
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Offset))
	return nil
}
func (a *Assembler) emitJmpIP1R(data InstJmpIP1R, out *[]byte) error {
	opcode := a.absoluteJmpToIPJmp(data.JmpTy)
	var reg byte
	reg = byte(data.OpTy)
	reg <<= 4
	//with second offset register
	reg |= byte(data.Reg.Reg & 0x0f)
	reg |= (0b0000_0011 & data.Reg.Size) << 6
	*out = binary.BigEndian.AppendUint32(*out, uint32(opcode))
	(*out)[len(*out)-4] = reg
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Offset))
	return nil
}
func (a *Assembler) declareLabel(data InstLab) error {
	// this needs to be the address of the function in the virtual address space
	_, posAsInstAddr := a.currentCodePos()
	posAsInstAddr -= decls.INSTRUCTION_SIZE
	if sym, _, ok := a.symbols.GetByName(data.Label); ok {
		switch sym.Vis {
		case aobj.SYM_VEXPORT:
			if sym.Loc != 0 {
				return errors.RedeclaredLabel(data.Label, data.DeclaredAt,
					a.line, a.col)
			} else {
				(*sym).Loc = posAsInstAddr
			}
		case aobj.SYM_VPRIVATE:
			return errors.RedeclaredLabel(data.Label, data.DeclaredAt,
				a.line, a.col)
		default:
			return errors.ImportedSymbolDeclared(data.Label, a.line, a.col)
		}
	} else {
		lab := aobj.SymbolData{
			Ty:   aobj.SYM_TFUNC,
			Vis:  aobj.SYM_VPRIVATE,
			Loc:  posAsInstAddr,
			Name: data.Label,
		}
		a.symbols.AddSymbol(lab)
	}
	return nil
}
func (a *Assembler) patchCall(pos int, address uint64) error {
	opcode := a.opCodes.GetBytes(vm.OP_CALL)
	binary.BigEndian.PutUint32(a.bytecode[pos:], uint32(opcode))
	binary.BigEndian.PutUint64(a.bytecode[pos+decls.OPCODE_SIZE:], uint64(address))
	return nil
}
func (a *Assembler) patchJmp(data PatchJmp, pos int, address uint64) error {
	opcode := a.jmpInstToOpCode(data.Variant)
	binary.BigEndian.PutUint32(a.bytecode[pos:], uint32(opcode))
	binary.BigEndian.PutUint64(a.bytecode[pos+decls.OPCODE_SIZE:], address)
	return nil
}
func (a *Assembler) patchCallIP(sym *aobj.SymbolData, pos int) error {
	opcode := a.opCodes.GetBytes(vm.OP_CALLIP)
	var reg byte
	reg = vm.OP_TADD
	reg <<= 4
	reg |= 0x0f
	posAsInstAddr := uint64(pos + vm.ADDRESSDEADZONE_SIZE)
	diff := int64(sym.Loc) - int64(posAsInstAddr)

	binary.BigEndian.PutUint32(a.bytecode[pos:], uint32(opcode))
	a.bytecode[pos] = reg
	binary.BigEndian.PutUint64(a.bytecode[pos+decls.OPCODE_SIZE:], uint64(diff))
	return nil
}
func (a *Assembler) patchJmpIP(data PatchJmp, sym *aobj.SymbolData, pos int) error {
	opcode := a.absoluteJmpToIPJmp(data.Variant)
	var reg byte
	reg = vm.OP_TADD
	reg <<= 4
	reg |= 0x0f
	posAsInstAddr := uint64(pos + vm.ADDRESSDEADZONE_SIZE)
	diff := int64(sym.Loc) - int64(posAsInstAddr)

	binary.BigEndian.PutUint32(a.bytecode[pos:], uint32(opcode))
	a.bytecode[pos] = reg
	binary.BigEndian.PutUint64(a.bytecode[pos+decls.OPCODE_SIZE:], uint64(diff))
	return nil
}
func (a *Assembler) resolveJumpInsturctions() error {
	for codePos, unresolved := range a.unresolvedJumps {
		sym, symIdx, ok := a.symbols.GetByName(unresolved.Ident)
		if !ok {
			return errors.UnresolvedSymbol(unresolved.Ident)
		}
		switch sym.Vis {
		case aobj.SYM_VPRIVATE:
			fallthrough
		case aobj.SYM_VEXPORT:
			var err error
			switch p := unresolved.PatchTy.(type) {
			case PatchCall:
				err = a.patchCallIP(sym, codePos)
			case PatchJmp:
				if unresolved.Absolute {
					err = a.patchJmp(p, codePos, sym.Loc)
					reloc := aobj.RelocData{
						Loc:       uint64(codePos) + decls.OPCODE_SIZE,
						Ref:       uint64(symIdx),
						PatchSize: 8,
					}
					a.relocations = append(a.relocations, reloc)
				} else {
					err = a.patchJmpIP(p, sym, codePos)
				}
			}
			// if unresolved.PatchTy == PATCH_CALL {
			// 	err = a.patchCallIP(sym, codePos)
			// } else if unresolved.Absolute {
			// 	err = a.patchJmp(unresolved, codePos, sym.Loc)
			// 	reloc := aobj.RelocData{
			// 		Loc:       uint64(codePos) + decls.OPCODE_SIZE,
			// 		Ref:       uint64(symIdx),
			// 		PatchSize: 8,
			// 	}
			// 	a.relocations = append(a.relocations, reloc)
			// } else {
			// 	err = a.patchJmpIP(unresolved, sym, codePos)
			// }
			if err != nil {
				return err
			}
		default:
			var err error
			switch p := unresolved.PatchTy.(type) {
			case PatchCall:
				err = a.patchCall(codePos, 0)
			case PatchJmp:
				err = a.patchJmp(p, codePos, 0)
			}
			// if unresolved.PatchTy == PATCH_CALL {
			// 	err = a.patchCall(codePos, 0)
			// } else {
			// 	err = a.patchJmp(unresolved, codePos, 0)
			// }
			if err != nil {
				return err
			}
			reloc := aobj.RelocData{
				Loc:       uint64(codePos) + decls.OPCODE_SIZE,
				Ref:       uint64(symIdx),
				PatchSize: 8,
			}
			a.relocations = append(a.relocations, reloc)
		}
	}
	return nil
}
func (a *Assembler) resolveSymbols() error {
	for sname, idx := range a.symbols.ByName {
		symbol := a.symbols.InOrder[idx]
		switch symbol.Vis {
		case aobj.SYM_VPRIVATE:
			if symbol.Loc == 0 {
				return errors.UnresolvedSymbol(sname)
			}
		case aobj.SYM_VEXPORT:
			if symbol.Loc == 0 {
				return errors.UnresolvedSymbol(sname)
			}
		}
	}
	return nil
}
func (a *Assembler) emitPushR(data InstPushR, out *[]byte) error {
	push := a.opCodes.GetBytes(vm.OP_PUSHR)
	*out = binary.BigEndian.AppendUint32(*out, uint32(push))
	(*out)[len(*out)-4] = 0b0000_0011 & data.DataSz
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Reg))
	return nil
}
func (a *Assembler) emitPushI(data InstPushI, out *[]byte) error {
	push := a.opCodes.GetBytes(vm.OP_PUSHI)
	*out = binary.BigEndian.AppendUint32(*out, uint32(push))
	(*out)[len(*out)-4] = 0b0000_0011 & data.DataSz
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Imm))
	return nil
}
func (a *Assembler) emitPop(data InstPop, out *[]byte) error {
	pop := a.opCodes.GetBytes(vm.OP_POP)
	*out = binary.BigEndian.AppendUint32(*out, uint32(pop))
	(*out)[len(*out)-4] = 0b0000_0011 & data.DataSz
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Reg))
	return nil
}
func (a *Assembler) emitNop(out *[]byte) error {
	nop := a.opCodes.GetBytes(vm.OP_NOP)
	*out = binary.BigEndian.AppendUint32(*out, uint32(nop))
	*out = binary.BigEndian.AppendUint64(*out, uint64(0))
	return nil
}
func (a *Assembler) emitClr(out *[]byte) error {
	clr := a.opCodes.GetBytes(vm.OP_CLR)
	*out = binary.BigEndian.AppendUint32(*out, uint32(clr))
	*out = binary.BigEndian.AppendUint64(*out, uint64(0))
	return nil
}
func (a *Assembler) emitCallIP0R(data InstCallIP0R, out *[]byte) error {
	call := a.opCodes.GetBytes(vm.OP_CALLIP)
	var reg byte
	reg = byte(data.OpTy)
	reg <<= 4
	reg |= 0x0f
	*out = binary.BigEndian.AppendUint32(*out, uint32(call))
	(*out)[len(*out)-4] = reg
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Offset))
	return nil
}
func (a *Assembler) emitCallIP1R(data InstCallIP1R, out *[]byte) error {
	call := a.opCodes.GetBytes(vm.OP_CALLIP)
	var reg byte
	reg = byte(data.OpTy)
	reg <<= 4
	if data.Reg.Size != vm.SZ_64 {
		return errors.BadRegisterSize(a.line, a.col)
	}
	reg |= byte(data.Reg.Reg & 0x0f)
	*out = binary.BigEndian.AppendUint32(*out, uint32(call))
	(*out)[len(*out)-4] = reg
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Offset))
	return nil
}
func (a *Assembler) emitCall(data InstCall, out *[]byte) error {
	call := a.opCodes.GetBytes(vm.OP_CALL)
	//direct call case
	if data.Addr != 0 {
		*out = binary.BigEndian.AppendUint32(*out, uint32(call))
		*out = binary.BigEndian.AppendUint64(*out, uint64(data.Addr))
	} else {
		position := len(*out)
		a.unresolvedJumps[position] = unresolvedJump{
			Ident:   data.Ident,
			PatchTy: PatchCall{},
		}
		*out = binary.BigEndian.AppendUint32(*out, uint32(a.opCodes.GetBytes(vm.OP_NOP)))
		*out = binary.BigEndian.AppendUint64(*out, 0)
	}
	return nil
}

func (a *Assembler) emitRet(out *[]byte) error {
	ret := a.opCodes.GetBytes(vm.OP_RET)
	*out = binary.BigEndian.AppendUint32(*out, uint32(ret))
	*out = binary.BigEndian.AppendUint64(*out, uint64(0))
	return nil
}
func (a *Assembler) emitExitI(data InstExitI, out *[]byte) error {
	exit := a.opCodes.GetBytes(vm.OP_EXITI)
	*out = binary.BigEndian.AppendUint32(*out, uint32(exit))
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Val))
	return nil
}
func (a *Assembler) emitExitR(data InstExitR, out *[]byte) error {
	exit := a.opCodes.GetBytes(vm.OP_EXITR)
	*out = binary.BigEndian.AppendUint32(*out, uint32(exit))
	param := [8]byte{
		byte(data.Reg.Reg),
		data.Reg.Size,
		0, 0, 0, 0, 0, 0,
	}
	*out = append(*out, param[:]...)
	return nil
}
