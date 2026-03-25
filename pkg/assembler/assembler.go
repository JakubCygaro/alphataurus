package assembler

import (
	"bufio"
	"encoding/binary"
	"fmt"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

type labelData struct {
	pos        uint64
	declaredAt string
}
type labelMap map[string]labelData
type unresolvedJumpMap map[int]string

type Assembler struct {
	parser          Parser
	opCodes         map[uint32]vm.OpCodeVal
	unresolvedJumps unresolvedJumpMap
	labels          labelMap
	lastInst        Instruction
	bytecode        []byte
	instCount       int
	symbols         SymbolTable
	relocations     RelocationTable
}

func (a *Assembler) InstructionCount() int {
	return a.instCount
}

func NewAssembler(reader bufio.Reader) Assembler {
	return Assembler{
		parser:          NewParser(reader),
		opCodes:         vm.GenerateOpcodeMap(),
		labels:          make(labelMap),
		unresolvedJumps: make(unresolvedJumpMap),
		bytecode:        make([]byte, 0, 64),
		symbols:         NewSymbolTable(),
		relocations:     make(RelocationTable, 0),
	}
}

func (a *Assembler) EmitBytecode() (int, error) {
	instCount := 0
	var ok bool
	var err error = nil
	ok, err = a.parser.ParseNext()
	for ; ok && err == nil; ok, err = a.parser.ParseNext() {
		inst := a.parser.CurrentInst()
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
			err = a.declareLabel(inst.Data.(InstLabData), &(a.bytecode))
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
		case INST_TRET:
			err = a.emitRet(&(a.bytecode))
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
	dest := 0b00001111 & byte(data.Dest)
	(*out)[len(*out)-4] = dest
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Imm))
	return nil
}

func (a *Assembler) emitMovRR(data InstMovData, out *[]byte) error {
	mov := a.opCodes[vm.OP_MOVRR]
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	destsrc := 0b00001111 & byte(data.Dest)
	destsrc |= (0b00001111 & byte(data.Src)) << 4
	(*out)[len(*out)-4] = destsrc
	*out = binary.BigEndian.AppendUint64(*out, uint64(0))
	return nil
}

func (a *Assembler) emitMovDRI(data InstDerefMovData, out *[]byte) error {
	mov := a.opCodes[vm.OP_MOVDRI]
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	(*out)[len(*out)-4] = byte(data.Dest)
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Offset))
	return nil
}
func (a *Assembler) emitMovDRO1(data InstDerefMovData, out *[]byte) error {
	mov := a.opCodes[vm.OP_MOVDRO1]
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	param := byte(data.Dest) << 4
	param |= (byte(data.OReg1) & 0x0f)
	(*out)[len(*out)-4] = param
	(*out)[len(*out)-3] = byte(data.OpTy)
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Offset))
	return nil
}
func (a *Assembler) emitMovDRO2(data InstDerefMovData, out *[]byte) error {
	mov := a.opCodes[vm.OP_MOVDRO2]
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	param := (byte(data.Dest) & 0x0f) << 4
	param |= (byte(data.OReg1) & 0x0f)
	(*out)[len(*out)-4] = param
	param = 0
	param = (byte(data.OReg2) & 0x0f) << 4
	param |= (byte(data.OpTy) & 0x0f)
	(*out)[len(*out)-3] = param
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Offset))
	return nil
}
func (a *Assembler) emitMovID(data InstMovDerefData, out *[]byte) error {
	mov := a.opCodes[vm.OP_MOVID]
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	*out = binary.BigEndian.AppendUint32(*out, uint32(data.Offset))
	*out = binary.BigEndian.AppendUint32(*out, uint32(data.Imm))
	return nil
}
func (a *Assembler) emitMovRD(data InstMovDerefData, out *[]byte) error {
	mov := a.opCodes[vm.OP_MOVRD]
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	param := data.Offset << 8
	param |= int64(byte(data.SourceReg))
	*out = binary.BigEndian.AppendUint64(*out, uint64(param))
	return nil
}
func (a *Assembler) emitMovIDO1(data InstMovDerefData, out *[]byte) error {
	mov := a.opCodes[vm.OP_MOVIDO1]
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	param := (byte(0) & 0x0f) << 4
	param |= (byte(data.OReg1) & 0x0f)
	(*out)[len(*out)-4] = param
	param = 0
	param = (byte(0) & 0x0f) << 4
	param |= (byte(data.OpTy) & 0x0f)
	(*out)[len(*out)-3] = param
	*out = binary.BigEndian.AppendUint32(*out, uint32(data.Offset))
	*out = binary.BigEndian.AppendUint32(*out, uint32(data.Imm))
	return nil
}
func (a *Assembler) emitMovRDO1(data InstMovDerefData, out *[]byte) error {
	mov := a.opCodes[vm.OP_MOVRDO1]
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	param := (byte(data.SourceReg) & 0x0f) << 4
	param |= (byte(data.OReg1) & 0x0f)
	(*out)[len(*out)-4] = param
	param = 0
	param = (byte(0) & 0x0f) << 4
	param |= (byte(data.OpTy) & 0x0f)
	(*out)[len(*out)-3] = param
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Offset))
	return nil
}
func (a *Assembler) emitMovIDO2(data InstMovDerefData, out *[]byte) error {
	mov := a.opCodes[vm.OP_MOVIDO2]
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	param := (byte(0) & 0x0f) << 4
	param |= (byte(data.OReg1) & 0x0f)
	(*out)[len(*out)-4] = param
	param = 0
	param = (byte(data.OReg2) & 0x0f) << 4
	param |= (byte(data.OpTy) & 0x0f)
	(*out)[len(*out)-3] = param
	*out = binary.BigEndian.AppendUint32(*out, uint32(data.Offset))
	*out = binary.BigEndian.AppendUint32(*out, uint32(data.Imm))
	return nil
}
func (a *Assembler) emitMovRDO2(data InstMovDerefData, out *[]byte) error {
	mov := a.opCodes[vm.OP_MOVRDO1]
	*out = binary.BigEndian.AppendUint32(*out, uint32(mov))
	param := (byte(data.SourceReg) & 0x0f) << 4
	param |= (byte(data.OReg1) & 0x0f)
	(*out)[len(*out)-4] = param
	param = 0
	param = (byte(data.OReg2) & 0x0f) << 4
	param |= (byte(data.OpTy) & 0x0f)
	(*out)[len(*out)-3] = param
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Offset))
	return nil
}
func (a *Assembler) emitArthRR(op int, data InstArthData, out *[]byte) error {
	var opCode vm.OpCodeVal
	switch op {
	case INST_TADDRR:
		opCode = a.opCodes[vm.OP_ADDRR]
	case INST_TSUBRR:
		opCode = a.opCodes[vm.OP_SUBRR]
	case INST_TDIVRR:
		opCode = a.opCodes[vm.OP_DIVRR]
	case INST_TMULRR:
		opCode = a.opCodes[vm.OP_MULRR]
	}
	*out = binary.BigEndian.AppendUint32(*out, uint32(opCode))
	// src = param[0]
	// dest = param[1]
	// ty = param[3]
	*out = append(*out, byte(data.Source))
	*out = append(*out, byte(data.Dest))
	*out = append(*out, byte(0))
	*out = append(*out, byte(data.Ty))
	*out = append(*out, 0, 0, 0, 0)
	return nil
}
func (a *Assembler) emitLogRR(op int, data InstLogicalData, out *[]byte) error {
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
	// first = param[0]
	// second = param[1]
	*out = append(*out, byte(data.First))
	*out = append(*out, byte(data.Second))
	*out = append(*out, byte(0))
	*out = append(*out, byte(0))
	*out = append(*out, 0, 0, 0, 0)
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
	(*out)[len(*out)-4] = byte(data.First)
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
	destTy := 0b00001111 & byte(data.Dest)
	destTy |= (0b00001111 & byte(data.Ty)) << 4
	(*out)[len(*out)-4] = destTy
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Imm))
	return nil
}
func (a *Assembler) emitNot(ty int, data InstLogicalData, out *[]byte) error {
	opCode := a.opCodes[vm.OP_NOT]
	*out = binary.BigEndian.AppendUint32(*out, uint32(opCode))
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.First))
	return nil
}
func (a *Assembler) emitInc(data InstIncDecData, out *[]byte) error {
	inc := a.opCodes[vm.OP_INCR]
	*out = binary.BigEndian.AppendUint32(*out, uint32(inc))
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Reg))
	return nil
}
func (a *Assembler) emitDec(data InstIncDecData, out *[]byte) error {
	dec := a.opCodes[vm.OP_DECR]
	*out = binary.BigEndian.AppendUint32(*out, uint32(dec))
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Reg))
	return nil
}
func (a *Assembler) emitCmpRR(data InstCmpData, out *[]byte) error {
	cmp := a.opCodes[vm.OP_CMP]
	*out = binary.BigEndian.AppendUint32(*out, uint32(cmp))
	// subtrahend |= (lastByte & 0xf0) >> 4
	// minuend |= (lastByte & 0x0f)
	regs := (0b00001111 & byte(data.Sub)) << 4
	regs |= 0b00001111 & byte(data.Min)
	(*out)[len(*out)-4] = regs
	*out = append(*out, byte(data.Ty), 0, 0, 0, 0, 0, 0, 0)
	return nil
}
func (a *Assembler) emitCmpIR(data InstCmpData, out *[]byte) error {
	cmp := a.opCodes[vm.OP_CMP]
	*out = binary.BigEndian.AppendUint32(*out, uint32(cmp))
	// subtrahend |= (lastByte & 0xf0) >> 4
	// minuend |= (lastByte & 0x0f)
	regs := (0b00001111 & (byte(data.Ty) + 0b00001000)) << 4
	regs |= 0b00001111 & byte(data.Min)
	(*out)[len(*out)-4] = regs
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Imm))
	return nil
}
func (a *Assembler) emitJmp(ty int, data InstJmpData, out *[]byte) error {
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
	opPos := len(*out)
	*out = binary.BigEndian.AppendUint32(*out, uint32(opcode))
	if addr, ok := data.Address.(uint64); ok {
		*out = binary.BigEndian.AppendUint64(*out, uint64(addr))
	} else if lab, ok := data.Address.(string); ok {
		if l, ok := a.labels[lab]; ok {
			*out = binary.BigEndian.AppendUint64(*out, uint64(l.pos/vm.INSTRUCTION_SIZE)+vm.ADDRESSDEADZONE_SIZE-1)
		} else {
			a.unresolvedJumps[opPos] = lab
			*out = binary.BigEndian.AppendUint64(*out, uint64(0))
		}
	} else {
		return fmt.Errorf("Jump instruction bad internal assembler data")
	}
	return nil
}
func (a *Assembler) emitJmpIP(ty int, data InstJmpIPData, out *[]byte) error {
	var opcode vm.OpCodeVal
	switch data.JmpTy {
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
	var reg byte
	reg = byte(data.OpTy)
	reg <<= 4
	if ty == INST_TJMPIP0R {
		reg |= 0x0f
	} else {
		reg |= byte(data.Reg & 0x0f)
	}
	*out = binary.BigEndian.AppendUint32(*out, uint32(opcode))
	(*out)[len(*out)-4] = reg
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Offset))
	return nil
}
func (a *Assembler) declareLabel(data InstLabData, out *[]byte) error {
	if lab, ok := a.labels[data.Label]; ok {
		return errors.RedeclaredLabel(data.Label, data.DeclaredAt,
			lab.declaredAt, a.parser.lexer.line, a.parser.lexer.col)
	}
	a.labels[data.Label] = labelData{
		pos:        uint64(len(*out)),
		declaredAt: data.DeclaredAt,
	}
	if esym, _, ok := a.symbols.GetByName(data.Label); ok && (esym.Vis == SYM_VEXPORT || esym.Vis == SYM_VPRIVATE) {
		(*esym).Loc = a.labels[data.Label].pos
	} else if ok && (esym.Vis == SYM_VIMPORTWEAK || esym.Vis == SYM_VIMPORTSTRONG) {
		return errors.ImportedSymbolDeclared(data.Label, a.parser.lexer.line, a.parser.lexer.col)
	}
	return nil
}
func (a *Assembler) resolveJumpInsturctions(out *[]byte) error {
	for codePos, destLabel := range a.unresolvedJumps {
		label, ok := a.labels[destLabel]
		if !ok {
			return errors.UnresolvedLabel(destLabel)
		}
		sym, idx, ok := a.symbols.GetByName(destLabel)
		if !ok {
			return errors.UnresolvedSymbol(destLabel)
		}
		if sym.Vis != SYM_VPRIVATE && sym.Vis != SYM_VEXPORT {
			return errors.UnresolvedSymbol(destLabel)
		}

		binary.BigEndian.PutUint64((*out)[codePos+vm.OPCODE_SIZE:],
			uint64(label.pos/vm.INSTRUCTION_SIZE)+vm.ADDRESSDEADZONE_SIZE-1)
		reloc := RelocData {
			Loc: uint64(codePos)+vm.OPCODE_SIZE,
			Ref: uint64(idx),
			PatchSize: 8,
		}
		a.relocations = append(a.relocations, reloc)
	}
	return nil
}
func (a *Assembler) resolveSymbols() error {
	for sname, idx := range a.symbols.ByName {
		symbol := a.symbols.InOrder[idx]
		switch symbol.Vis {
		case SYM_VPRIVATE:
			if symbol.Loc == 0 {
				return errors.UnresolvedSymbol(sname)
			}
		case SYM_VEXPORT:
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
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Reg))
	return nil
}
func (a *Assembler) emitPushI(data InstPushPopData, out *[]byte) error {
	push := a.opCodes[vm.OP_PUSHI]
	*out = binary.BigEndian.AppendUint32(*out, uint32(push))
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Imm))
	return nil
}
func (a *Assembler) emitPop(data InstPushPopData, out *[]byte) error {
	pop := a.opCodes[vm.OP_POP]
	*out = binary.BigEndian.AppendUint32(*out, uint32(pop))
	*out = binary.BigEndian.AppendUint64(*out, uint64(data.Reg))
	return nil
}
func (a *Assembler) emitNop(out *[]byte) error {
	nop := a.opCodes[vm.OP_NOP]
	*out = binary.BigEndian.AppendUint32(*out, uint32(nop))
	*out = binary.BigEndian.AppendUint64(*out, uint64(0))
	return nil
}
func (a *Assembler) emitCall(data InstCallData, out *[]byte) error {
	call := a.opCodes[vm.OP_CALL]
	*out = binary.BigEndian.AppendUint32(*out, uint32(call))
	if data.Addr != 0 {
		*out = binary.BigEndian.AppendUint64(*out, uint64(data.Addr))
	} else {
		if lab, ok := a.labels[data.Ident]; ok {
			*out = binary.BigEndian.AppendUint64(*out, uint64(lab.pos/vm.INSTRUCTION_SIZE)+vm.ADDRESSDEADZONE_SIZE-1)
		} else {
			opPos := len(*out)
			a.unresolvedJumps[opPos] = data.Ident
			*out = binary.BigEndian.AppendUint64(*out, uint64(0))
		}
	}
	return nil
}
func (a *Assembler) emitRet(out *[]byte) error {
	ret := a.opCodes[vm.OP_RET]
	*out = binary.BigEndian.AppendUint32(*out, uint32(ret))
	*out = binary.BigEndian.AppendUint64(*out, uint64(0))
	return nil
}
