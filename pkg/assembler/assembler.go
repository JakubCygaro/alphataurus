package assembler

import (
	"bufio"
	"encoding/binary"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

type unresolvedJump struct {
	Ident string
	// what instruction is gonna get patched
	InstTy int
}
type unresolvedJumpMap map[int]unresolvedJump

type Assembler struct {
	parser          Parser
	opCodes         map[uint32]vm.OpCodeVal
	unresolvedJumps unresolvedJumpMap
	//relating to the current instruction
	line, col uint64
	// labels          labelMap
	lastInst    Instruction
	bytecode    []byte
	instCount   int
	symbols     vm.SymbolTable
	relocations vm.RelocationTable
	hasEntry    bool
	entry       uint64
}

func (a *Assembler) InstructionCount() int {
	return a.instCount
}

func NewAssembler(reader bufio.Reader) Assembler {
	return Assembler{
		parser:          NewParser(reader),
		opCodes:         vm.GenerateOpcodeMap(),
		unresolvedJumps: make(unresolvedJumpMap),
		bytecode:        make([]byte, 0, 64),
		symbols:         vm.NewSymbolTable(),
		relocations:     make(vm.RelocationTable, 0),
		hasEntry:        false,
	}
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
		switch inst.Ty {
		case INST_TMOVIR:
			err = a.emitMovIR(inst.Data.(InstMovData), &(a.bytecode))
		case INST_TMOVRR:
			err = a.emitMovRR(inst.Data.(InstMovData), &(a.bytecode))
		case INST_TMOVDRI:
			err = a.emitMovDRI(inst.Data.(InstDerefMovData), &(a.bytecode))
		case INST_TMOVDRO1:
			err = a.emitMovDRO1(inst.Data.(InstDerefMovData), &(a.bytecode))
		case INST_TMOVDRO2:
			err = a.emitMovDRO2(inst.Data.(InstDerefMovData), &(a.bytecode))
		case INST_TMOVID:
			err = a.emitMovID(inst.Data.(InstMovDerefData), &(a.bytecode))
		case INST_TMOVRD:
			err = a.emitMovRD(inst.Data.(InstMovDerefData), &(a.bytecode))
		case INST_TMOVIDO1:
			err = a.emitMovIDO1(inst.Data.(InstMovDerefData), &(a.bytecode))
		case INST_TMOVRDO1:
			err = a.emitMovRDO1(inst.Data.(InstMovDerefData), &(a.bytecode))
		case INST_TMOVIDO2:
			err = a.emitMovIDO2(inst.Data.(InstMovDerefData), &(a.bytecode))
		case INST_TMOVRDO2:
			err = a.emitMovRDO2(inst.Data.(InstMovDerefData), &(a.bytecode))
		case INST_TADDRR:
			err = a.emitArthRR(int(inst.Ty), inst.Data.(InstArthData), &(a.bytecode))
		case INST_TSUBRR:
			err = a.emitArthRR(int(inst.Ty), inst.Data.(InstArthData), &(a.bytecode))
		case INST_TMULRR:
			err = a.emitArthRR(int(inst.Ty), inst.Data.(InstArthData), &(a.bytecode))
		case INST_TDIVRR:
			err = a.emitArthRR(int(inst.Ty), inst.Data.(InstArthData), &(a.bytecode))
		case INST_TADDIR:
			err = a.emitArthIR(int(inst.Ty), inst.Data.(InstArthData), &(a.bytecode))
		case INST_TSUBIR:
			err = a.emitArthIR(int(inst.Ty), inst.Data.(InstArthData), &(a.bytecode))
		case INST_TNOT:
			err = a.emitNot(int(inst.Ty), inst.Data.(InstLogicalData), &(a.bytecode))
		case INST_TANDRR:
			err = a.emitLogRR(int(inst.Ty), inst.Data.(InstLogicalData), &(a.bytecode))
		case INST_TORRR:
			err = a.emitLogRR(int(inst.Ty), inst.Data.(InstLogicalData), &(a.bytecode))
		case INST_TXORRR:
			err = a.emitLogRR(int(inst.Ty), inst.Data.(InstLogicalData), &(a.bytecode))
		case INST_TLSHRR:
			err = a.emitLogRR(int(inst.Ty), inst.Data.(InstLogicalData), &(a.bytecode))
		case INST_TRSHRR:
			err = a.emitLogRR(int(inst.Ty), inst.Data.(InstLogicalData), &(a.bytecode))
		case INST_TANDIR:
			err = a.emitLogIR(int(inst.Ty), inst.Data.(InstLogicalData), &(a.bytecode))
		case INST_TORIR:
			err = a.emitLogIR(int(inst.Ty), inst.Data.(InstLogicalData), &(a.bytecode))
		case INST_TXORIR:
			err = a.emitLogIR(int(inst.Ty), inst.Data.(InstLogicalData), &(a.bytecode))
		case INST_TLSHIR:
			err = a.emitLogIR(int(inst.Ty), inst.Data.(InstLogicalData), &(a.bytecode))
		case INST_TRSHIR:
			err = a.emitLogIR(int(inst.Ty), inst.Data.(InstLogicalData), &(a.bytecode))
		case INST_TINCR:
			err = a.emitInc(inst.Data.(InstIncDecData), &(a.bytecode))
		case INST_TDECR:
			err = a.emitDec(inst.Data.(InstIncDecData), &(a.bytecode))
		case INST_TCMPRR:
			err = a.emitCmpRR(inst.Data.(InstCmpData), &(a.bytecode))
		case INST_TCMPIR:
			err = a.emitCmpIR(inst.Data.(InstCmpData), &(a.bytecode))
		case INST_TJMP:
			err = a.emitJmp(int(inst.Ty), inst.Data.(InstJmpData), &(a.bytecode))
		case INST_TJMPE:
			err = a.emitJmp(int(inst.Ty), inst.Data.(InstJmpData), &(a.bytecode))
		case INST_TJMPNE:
			err = a.emitJmp(int(inst.Ty), inst.Data.(InstJmpData), &(a.bytecode))
		case INST_TJMPZ:
			err = a.emitJmp(int(inst.Ty), inst.Data.(InstJmpData), &(a.bytecode))
		case INST_TJMPNZ:
			err = a.emitJmp(int(inst.Ty), inst.Data.(InstJmpData), &(a.bytecode))
		case INST_TJMPG:
			err = a.emitJmp(int(inst.Ty), inst.Data.(InstJmpData), &(a.bytecode))
		case INST_TJMPGE:
			err = a.emitJmp(int(inst.Ty), inst.Data.(InstJmpData), &(a.bytecode))
		case INST_TJMPL:
			err = a.emitJmp(int(inst.Ty), inst.Data.(InstJmpData), &(a.bytecode))
		case INST_TJMPLE:
			err = a.emitJmp(int(inst.Ty), inst.Data.(InstJmpData), &(a.bytecode))
		case INST_TJMPIP0R:
			err = a.emitJmpIP(int(inst.Ty), inst.Data.(InstJmpIPData), &(a.bytecode))
		case INST_TJMPIP1R:
			err = a.emitJmpIP(int(inst.Ty), inst.Data.(InstJmpIPData), &(a.bytecode))
		case INST_TLABEL:
			err = a.declareLabel(inst.Data.(InstLabData))
			instCount--
		case INST_TPUSHR:
			err = a.emitPushR(inst.Data.(InstPushPopData), &(a.bytecode))
		case INST_TPUSHI:
			err = a.emitPushI(inst.Data.(InstPushPopData), &(a.bytecode))
		case INST_TPOP:
			err = a.emitPop(inst.Data.(InstPushPopData), &(a.bytecode))
		case INST_TNOP:
			err = a.emitNop(&(a.bytecode))
		case INST_TCALL:
			err = a.emitCall(inst.Data.(InstCallData), &(a.bytecode))
		case INST_TCALLIP0R:
			err = a.emitCallIP(int(inst.Ty), inst.Data.(InstCallIPData), &(a.bytecode))
		case INST_TCALLIP1R:
			err = a.emitCallIP(int(inst.Ty), inst.Data.(InstCallIPData), &(a.bytecode))
		case INST_TRET:
			err = a.emitRet(&(a.bytecode))
		case INST_TEXIT:
			err = a.emitExit(inst.Data.(InstExitData), &(a.bytecode))
		case INST_TATTRENTRY:
			if a.hasEntry {
				err = errors.MultipleEntry(inst.Line, inst.Col)
			} else {
				a.hasEntry = true
				_, ent := a.currentCodePos()
				a.entry = ent
			}

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

func (a *Assembler) emitMovIR(data InstMovData, out *[]byte) error {
	mov := a.opCodes[vm.OP_MOVIR]
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	lastByte :=
		(0b0000_1111 & byte(data.Dest)) |
			(0b0011_0000 & (byte(0) << 4)) |
			(0b1100_0000 & (byte(data.DataSize) << 6))
	(*out)[len(*out)-4] = lastByte
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Imm))
	return nil
}

func (a *Assembler) emitMovRR(data InstMovData, out *[]byte) error {
	mov := a.opCodes[vm.OP_MOVRR]
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	destsrc := 0b00001111 & byte(data.Dest)
	destsrc |= (0b00001111 & byte(data.Src)) << 4
	dataSz := (0b0000_0011 & data.DataSize)
	(*out)[len(*out)-4] = dataSz
	*out = binary.BigEndian.AppendUint64(*out, uint64(0))
	(*out)[len(*out)-1] = destsrc
	return nil
}

func (a *Assembler) emitMovDRI(data InstDerefMovData, out *[]byte) error {
	mov := a.opCodes[vm.OP_MOVDRI]
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	lastByte := (0b0000_1111 & byte(data.Dest.Reg))
	lastByte |= (0b0000_0011 & data.Dest.Size) << 4
	(*out)[len(*out)-4] = lastByte
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Offset))
	return nil
}
func (a *Assembler) emitMovDRO1(data InstDerefMovData, out *[]byte) error {
	mov := a.opCodes[vm.OP_MOVDRO1]
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
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
func (a *Assembler) emitMovDRO2(data InstDerefMovData, out *[]byte) error {
	mov := a.opCodes[vm.OP_MOVDRO2]
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	lastByte := (0b0000_1111 & byte(data.Dest.Reg)) << 4
	lastByte |= (0b0000_1111 & byte(data.OReg1.Reg))
	penultByte := (0b0000_1111 & byte(data.OReg2.Reg)) << 4
	penultByte |= (0b0000_0011 & byte(data.OReg1.Size)) << 2
	penultByte |= (0b0000_0011 & byte(data.OpTy))
	(*out)[len(*out)-4] = lastByte
	(*out)[len(*out)-3] = penultByte
	param := uint64(data.Offset) & 0x3f_ff_ff_ff
	param |= uint64(data.Dest.Size&0b0000_0011) << 62
	*out = binary.BigEndian.AppendUint64(*out, param)
	return nil
}
func (a *Assembler) emitMovID(data InstMovDerefData, out *[]byte) error {
	mov := a.opCodes[vm.OP_MOVID]
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	lastByte := 0b0000_0011 & data.DataSize
	(*out)[len(*out)-4] = lastByte
	*out = binary.BigEndian.AppendUint32(*out, uint32(data.Offset))
	*out = binary.BigEndian.AppendUint32(*out, uint32(data.Imm))
	return nil
}
func (a *Assembler) emitMovRD(data InstMovDerefData, out *[]byte) error {
	mov := a.opCodes[vm.OP_MOVRD]
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	lastByte := 0b0000_1111 & byte(data.SourceReg.Reg)
	lastByte |= (0b0000_0011 & data.SourceReg.Size) << 4
	(*out)[len(*out)-4] = lastByte
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Offset))
	return nil
}
func (a *Assembler) emitMovIDO1(data InstMovDerefData, out *[]byte) error {
	mov := a.opCodes[vm.OP_MOVIDO1]
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	// lastByte := (0b0000_1111 & byte(data.Dest.Reg)) << 4
	lastByte := byte(0)
	lastByte |= (0b0000_1111 & byte(data.OReg1.Reg))
	penultByte := (0b0000_0011 & byte(data.DataSize)) << 4
	penultByte |= (0b0000_0011 & byte(data.OReg1.Size)) << 2
	penultByte |= (0b0000_0011 & byte(data.OpTy))
	(*out)[len(*out)-4] = lastByte
	(*out)[len(*out)-3] = penultByte
	*out = binary.BigEndian.AppendUint32(*out, uint32(data.Offset))
	*out = binary.BigEndian.AppendUint32(*out, uint32(data.Imm))
	return nil
}
func (a *Assembler) emitMovRDO1(data InstMovDerefData, out *[]byte) error {
	mov := a.opCodes[vm.OP_MOVRDO1]
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	lastByte := (0b0000_1111 & byte(data.SourceReg.Reg)) << 4
	lastByte |= (0b0000_1111 & byte(data.OReg1.Reg))
	penultByte := (0b0000_0011 & byte(data.SourceReg.Size)) << 4
	penultByte |= (0b0000_0011 & byte(data.OReg1.Size)) << 2
	penultByte |= (0b0000_0011 & byte(data.OpTy))
	(*out)[len(*out)-4] = lastByte
	(*out)[len(*out)-3] = penultByte
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Offset))
	return nil
}
func (a *Assembler) emitMovIDO2(data InstMovDerefData, out *[]byte) error {
	if data.OReg1.Size != data.OReg2.Size {
		return errors.MismatchedRegisterSizes(
			a.parser.currentInst.Line,
			a.parser.currentInst.Col,
		)
	}
	mov := a.opCodes[vm.OP_MOVIDO2]
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	lastByte := (0b0000_1111 & byte(0)) << 4
	lastByte |= (0b0000_1111 & byte(data.OReg1.Reg))
	penultByte := (0b0000_1111 & byte(data.OReg2.Reg)) << 4
	penultByte |= (0b0000_0011 & byte(data.OReg1.Size)) << 2
	penultByte |= (0b0000_0011 & byte(data.OpTy))
	(*out)[len(*out)-4] = lastByte
	(*out)[len(*out)-3] = penultByte
	*out = binary.BigEndian.AppendUint32(*out, uint32(data.Offset))
	*out = binary.BigEndian.AppendUint32(*out, uint32(data.Imm))
	(*out)[len(*out)-8] = (0b0000_0011 & data.DataSize) << 6
	return nil
}
func (a *Assembler) emitMovRDO2(data InstMovDerefData, out *[]byte) error {
	if data.OReg1.Size != data.OReg2.Size ||
		data.SourceReg.Size < data.OReg1.Size {
		return errors.MismatchedRegisterSizes(
			a.parser.currentInst.Line,
			a.parser.currentInst.Col,
		)
	}
	mov := a.opCodes[vm.OP_MOVRDO1]
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	lastByte := (0b0000_1111 & byte(data.SourceReg.Reg)) << 4
	lastByte |= (0b0000_1111 & byte(data.OReg1.Reg))
	penultByte := (0b0000_1111 & byte(data.OReg2.Reg)) << 4
	penultByte |= (0b0000_0011 & byte(data.OReg1.Size)) << 2
	penultByte |= (0b0000_0011 & byte(data.OpTy))
	(*out)[len(*out)-4] = lastByte
	(*out)[len(*out)-3] = penultByte
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Offset))
	(*out)[len(*out)-8] = (0b0000_0011 & data.SourceReg.Size) << 6
	return nil
}
func (a *Assembler) emitArthRR(op int, data InstArthData, out *[]byte) error {
	var opCode vm.OpCodeVal
	sized := false
	switch op {
	case INST_TADDRR:
		opCode = a.opCodes[vm.OP_ADDRR]
	case INST_TSUBRR:
		opCode = a.opCodes[vm.OP_SUBRR]
	case INST_TDIVRR:
		opCode = a.opCodes[vm.OP_DIVRR]
		sized = true
	case INST_TMULRR:
		opCode = a.opCodes[vm.OP_MULRR]
		sized = true
	}
	*out = binary.BigEndian.AppendUint32(*out, uint32(opCode))
	tySizesByte := (0b0000_0011 & byte(data.Ty))

	if sized {
		tySizesByte |= (0b000_0011 & byte(data.DataSize)) << 2
		tySizesByte |= (0b000_0011 & byte(data.DataSize)) << 4
	} else {
		tySizesByte |= (0b000_0011 & byte(data.Source.Size)) << 2
		tySizesByte |= (0b000_0011 & byte(data.Dest.Size)) << 4
	}
	param := [8]byte{
		byte(data.Source.Reg),
		byte(data.Dest.Reg),
		byte(tySizesByte),
		0,
		0,
		0,
		0,
		0,
	}
	*out = append(*out, param[:]...)
	// *out = append(*out, byte(data.Source))
	// *out = append(*out, byte(data.Dest))
	// *out = append(*out, byte(0))
	// *out = append(*out, byte(data.Ty))
	// *out = append(*out, 0, 0, 0, 0)
	return nil
}
func (a *Assembler) emitLogRR(op int, data InstLogicalData, out *[]byte) error {
	if data.First.Size < data.Second.Size {
		return errors.MismatchedRegisterSizes(
			a.parser.currentInst.Line,
			a.parser.currentInst.Col,
		)
	}
	var opCode vm.OpCodeVal
	switch op {
	case INST_TANDRR:
		opCode = a.opCodes[vm.OP_ANDRR]
	case INST_TORRR:
		opCode = a.opCodes[vm.OP_ORRR]
	case INST_TXORRR:
		opCode = a.opCodes[vm.OP_XORRR]
	case INST_TLSHRR:
		opCode = a.opCodes[vm.OP_LSHRR]
	case INST_TRSHRR:
		opCode = a.opCodes[vm.OP_RSHRR]
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
func (a *Assembler) emitLogIR(ty int, data InstLogicalData, out *[]byte) error {
	var opCode vm.OpCodeVal
	switch ty {
	case INST_TANDIR:
		opCode = a.opCodes[vm.OP_ANDIR]
	case INST_TORIR:
		opCode = a.opCodes[vm.OP_ORIR]
	case INST_TXORIR:
		opCode = a.opCodes[vm.OP_XORIR]
	case INST_TLSHIR:
		opCode = a.opCodes[vm.OP_LSHIR]
	case INST_TRSHIR:
		opCode = a.opCodes[vm.OP_RSHIR]
	}
	*out = binary.BigEndian.AppendUint32(*out, uint32(opCode))
	lastByte := byte(0)
	lastByte |= 0b0000_1111 & byte(data.First.Reg)
	lastByte |= (0b0000_0011 & byte(data.First.Size)) << 4
	(*out)[len(*out)-4] = lastByte
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Imm))
	return nil
}
func (a *Assembler) emitArthIR(ty int, data InstArthData, out *[]byte) error {
	var opCode vm.OpCodeVal
	switch ty {
	case INST_TADDIR:
		opCode = a.opCodes[vm.OP_ADDIR]
	case INST_TSUBIR:
		opCode = a.opCodes[vm.OP_SUBIR]
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
func (a *Assembler) emitNot(ty int, data InstLogicalData, out *[]byte) error {
	opCode := a.opCodes[vm.OP_NOT]
	*out = binary.BigEndian.AppendUint32(*out, uint32(opCode))
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.First.Reg))
	return nil
}
func (a *Assembler) emitInc(data InstIncDecData, out *[]byte) error {
	inc := a.opCodes[vm.OP_INCR]
	*out = binary.BigEndian.AppendUint32(*out, uint32(inc))
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Reg.Reg))
	return nil
}
func (a *Assembler) emitDec(data InstIncDecData, out *[]byte) error {
	dec := a.opCodes[vm.OP_DECR]
	*out = binary.BigEndian.AppendUint32(*out, uint32(dec))
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Reg.Reg))
	return nil
}
func (a *Assembler) emitCmpRR(data InstCmpData, out *[]byte) error {
	cmp := a.opCodes[vm.OP_CMPRR]
	*out = binary.BigEndian.AppendUint32(*out, uint32(cmp))
	regs := (0b0000_1111 & byte(data.Sub.Reg)) << 4
	regs |= 0b0000_1111 & byte(data.Min.Reg)
	// regs |= 0b0000_0011 & byte(data.Min.Size) << 2
	(*out)[len(*out)-4] = regs
	*out = append(*out, byte(data.Ty), byte(data.Min.Size), 0, 0, 0, 0, 0, 0)
	return nil
}
func (a *Assembler) emitCmpIR(data InstCmpData, out *[]byte) error {
	cmp := a.opCodes[vm.OP_CMPIR]
	*out = binary.BigEndian.AppendUint32(*out, uint32(cmp))
	// subtrahend |= (lastByte & 0xf0) >> 4
	// minuend |= (lastByte & 0x0f)
	regs := (0b00001111 & byte(data.Min.Reg)) << 4
	regs |= (0b00001111 & byte(data.Min.Size)) << 2
	if data.ImmIsFloat {
		regs |= 0b0000_0001
	}
	(*out)[len(*out)-4] = regs
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Imm))
	return nil
}
func (a *Assembler) jmpInstToOpCode(ty int) vm.OpCodeVal {
	var opcode vm.OpCodeVal
	switch ty {
	case INST_TJMP:
		opcode = a.opCodes[vm.OP_JMP]
	case INST_TJMPE:
		opcode = a.opCodes[vm.OP_JMPE]
	case INST_TJMPNE:
		opcode = a.opCodes[vm.OP_JMPNE]
	case INST_TJMPZ:
		opcode = a.opCodes[vm.OP_JMPZ]
	case INST_TJMPNZ:
		opcode = a.opCodes[vm.OP_JMPNZ]
	case INST_TJMPG:
		opcode = a.opCodes[vm.OP_JMPG]
	case INST_TJMPGE:
		opcode = a.opCodes[vm.OP_JMPGE]
	case INST_TJMPL:
		opcode = a.opCodes[vm.OP_JMPL]
	case INST_TJMPLE:
		opcode = a.opCodes[vm.OP_JMPLE]
	}
	return opcode
}
func (a *Assembler) emitJmp(ty int, data InstJmpData, out *[]byte) error {
	opcode := a.jmpInstToOpCode(ty)
	position := len(*out)
	if addr, ok := data.Address.(uint64); ok {
		*out = binary.BigEndian.AppendUint32(*out, uint32(opcode))
		*out = binary.BigEndian.AppendUint64(*out, uint64(addr))
	} else {
		a.unresolvedJumps[position] = unresolvedJump{
			Ident:  data.Address.(string),
			InstTy: ty,
		}
		*out = binary.BigEndian.AppendUint32(*out, uint32(a.opCodes[vm.OP_NOP]))
		*out = binary.BigEndian.AppendUint64(*out, uint64(0))
	}
	return nil
}

func (a *Assembler) absoluteJmpToIPJmp(instTy int) (opcode vm.OpCodeVal) {
	switch instTy {
	case INST_TJMP:
		opcode = a.opCodes[vm.OP_JMPIP]
	case INST_TJMPE:
		opcode = a.opCodes[vm.OP_JMPEIP]
	case INST_TJMPNE:
		opcode = a.opCodes[vm.OP_JMPNEIP]
	case INST_TJMPZ:
		opcode = a.opCodes[vm.OP_JMPZIP]
	case INST_TJMPNZ:
		opcode = a.opCodes[vm.OP_JMPNZIP]
	case INST_TJMPG:
		opcode = a.opCodes[vm.OP_JMPGIP]
	case INST_TJMPGE:
		opcode = a.opCodes[vm.OP_JMPGEIP]
	case INST_TJMPL:
		opcode = a.opCodes[vm.OP_JMPLIP]
	case INST_TJMPLE:
		opcode = a.opCodes[vm.OP_JMPLEIP]
	}
	return opcode
}

func (a *Assembler) emitJmpIP(ty int, data InstJmpIPData, out *[]byte) error {
	opcode := a.absoluteJmpToIPJmp(data.JmpTy)
	var reg byte
	reg = byte(data.OpTy)
	reg <<= 4
	// no second offset register
	if ty == INST_TJMPIP0R {
		reg |= 0x0f
	} else {
		//with second offset register
		reg |= byte(data.Reg.Reg & 0x0f)
		reg |= (0b0000_0011 & data.Reg.Size) << 6
	}
	*out = binary.BigEndian.AppendUint32(*out, uint32(opcode))
	(*out)[len(*out)-4] = reg
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Offset))
	return nil
}
func (a *Assembler) declareLabel(data InstLabData) error {
	// this needs to be the address of the function in the virtual address space
	_, posAsInstAddr := a.currentCodePos()
	posAsInstAddr -= vm.INSTRUCTION_SIZE
	if sym, _, ok := a.symbols.GetByName(data.Label); ok {
		switch sym.Vis {
		case vm.SYM_VEXPORT:
			if sym.Loc != 0 {
				return errors.RedeclaredLabel(data.Label, data.DeclaredAt,
					a.line, a.col)
			} else {
				(*sym).Loc = posAsInstAddr
			}
		case vm.SYM_VPRIVATE:
			return errors.RedeclaredLabel(data.Label, data.DeclaredAt,
				a.line, a.col)
		default:
			return errors.ImportedSymbolDeclared(data.Label, a.line, a.col)
		}
	} else {
		a.symbols.AddSymbol(vm.SymbolData{
			Ty:   vm.SYM_TFUNC,
			Vis:  vm.SYM_VPRIVATE,
			Loc:  posAsInstAddr,
			Name: data.Label,
		})
	}
	return nil
}
func (a *Assembler) patchCall(pos int, address uint64) error {
	opcode := a.opCodes[vm.OP_CALL]
	binary.BigEndian.PutUint32(a.bytecode[pos:], uint32(opcode))
	binary.BigEndian.PutUint64(a.bytecode[pos+vm.OPCODE_SIZE:], uint64(address))
	return nil
}
func (a *Assembler) patchJmp(data unresolvedJump, pos int, address uint64) error {
	opcode := a.jmpInstToOpCode(data.InstTy)
	binary.BigEndian.PutUint32(a.bytecode[pos:], uint32(opcode))
	binary.BigEndian.PutUint64(a.bytecode[pos+vm.OPCODE_SIZE:], address)
	return nil
}
func (a *Assembler) patchCallIP(data unresolvedJump, sym *vm.SymbolData, pos int) error {
	opcode := a.opCodes[vm.OP_CALLIP]
	var reg byte
	reg = vm.OP_TADD
	reg <<= 4
	reg |= 0x0f
	posAsInstAddr := uint64(pos + vm.ADDRESSDEADZONE_SIZE)
	diff := int64(sym.Loc) - int64(posAsInstAddr)

	binary.BigEndian.PutUint32(a.bytecode[pos:], uint32(opcode))
	a.bytecode[pos] = reg
	binary.BigEndian.PutUint64(a.bytecode[pos+vm.OPCODE_SIZE:], uint64(diff))
	return nil
}
func (a *Assembler) patchJmpIP(data unresolvedJump, sym *vm.SymbolData, pos int) error {
	opcode := a.absoluteJmpToIPJmp(data.InstTy)
	var reg byte
	reg = vm.OP_TADD
	reg <<= 4
	reg |= 0x0f
	posAsInstAddr := uint64(pos + vm.ADDRESSDEADZONE_SIZE)
	diff := int64(sym.Loc) - int64(posAsInstAddr)

	binary.BigEndian.PutUint32(a.bytecode[pos:], uint32(opcode))
	a.bytecode[pos] = reg
	binary.BigEndian.PutUint64(a.bytecode[pos+vm.OPCODE_SIZE:], uint64(diff))
	return nil
}
func (a *Assembler) resolveJumpInsturctions() error {
	for codePos, unresolved := range a.unresolvedJumps {
		sym, symIdx, ok := a.symbols.GetByName(unresolved.Ident)
		if !ok {
			return errors.UnresolvedSymbol(unresolved.Ident)
		}
		switch {
		case sym.Vis == vm.SYM_VPRIVATE || sym.Vis == vm.SYM_VEXPORT:
			var err error
			if unresolved.InstTy == INST_TCALL {
				err = a.patchCallIP(unresolved, sym, codePos)
			} else {
				err = a.patchJmpIP(unresolved, sym, codePos)
			}
			if err != nil {
				return err
			}
		default:
			var err error
			if unresolved.InstTy == INST_TCALL {
				err = a.patchCall(codePos, 0)
			} else {
				err = a.patchJmp(unresolved, codePos, 0)
			}
			if err != nil {
				return err
			}
			reloc := vm.RelocData{
				Loc:       uint64(codePos) + vm.OPCODE_SIZE,
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
		case vm.SYM_VPRIVATE:
			if symbol.Loc == 0 {
				return errors.UnresolvedSymbol(sname)
			}
		case vm.SYM_VEXPORT:
			if symbol.Loc == 0 {
				return errors.UnresolvedSymbol(sname)
			}
		}
	}
	return nil
}
func (a *Assembler) emitPushR(data InstPushPopData, out *[]byte) error {
	push := a.opCodes[vm.OP_PUSHR]
	*out = binary.BigEndian.AppendUint32(*out, uint32(push))
	(*out)[len(*out)-4] = 0b0000_0011 & data.DataSz
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Reg))
	return nil
}
func (a *Assembler) emitPushI(data InstPushPopData, out *[]byte) error {
	push := a.opCodes[vm.OP_PUSHI]
	*out = binary.BigEndian.AppendUint32(*out, uint32(push))
	(*out)[len(*out)-4] = 0b0000_0011 & data.DataSz
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Imm))
	return nil
}
func (a *Assembler) emitPop(data InstPushPopData, out *[]byte) error {
	pop := a.opCodes[vm.OP_POP]
	*out = binary.BigEndian.AppendUint32(*out, uint32(pop))
	(*out)[len(*out)-4] = 0b0000_0011 & data.DataSz
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Reg))
	return nil
}
func (a *Assembler) emitNop(out *[]byte) error {
	nop := a.opCodes[vm.OP_NOP]
	*out = binary.BigEndian.AppendUint32(*out, uint32(nop))
	*out = binary.BigEndian.AppendUint64(*out, uint64(0))
	return nil
}
func (a *Assembler) emitCallIP(ty int, data InstCallIPData, out *[]byte) error {
	call := a.opCodes[vm.OP_CALLIP]
	var reg byte
	reg = byte(data.OpTy)
	reg <<= 4
	if ty == INST_TCALLIP0R {
		reg |= 0x0f
	} else {
		if data.Reg.Size != vm.SZ_64 {
			return errors.BadRegisterSize(a.line, a.col)
		}
		reg |= byte(data.Reg.Reg & 0x0f)
	}
	*out = binary.BigEndian.AppendUint32(*out, uint32(call))
	(*out)[len(*out)-4] = reg
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Offset))
	return nil
}
func (a *Assembler) emitCall(data InstCallData, out *[]byte) error {
	call := a.opCodes[vm.OP_CALL]
	//direct call case
	if data.Addr != 0 {
		*out = binary.BigEndian.AppendUint32(*out, uint32(call))
		*out = binary.BigEndian.AppendUint64(*out, uint64(data.Addr))
	} else {
		position := len(*out)
		a.unresolvedJumps[position] = unresolvedJump{Ident: data.Ident, InstTy: INST_TCALL}
		*out = binary.BigEndian.AppendUint32(*out, uint32(a.opCodes[vm.OP_NOP]))
		*out = binary.BigEndian.AppendUint64(*out, 0)
	}
	return nil
}

func (a *Assembler) emitRet(out *[]byte) error {
	ret := a.opCodes[vm.OP_RET]
	*out = binary.BigEndian.AppendUint32(*out, uint32(ret))
	*out = binary.BigEndian.AppendUint64(*out, uint64(0))
	return nil
}
func (a *Assembler) emitExit(data InstExitData, out *[]byte) error {
	ret := a.opCodes[vm.OP_EXIT]
	*out = binary.BigEndian.AppendUint32(*out, uint32(ret))
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Val))
	return nil
}
