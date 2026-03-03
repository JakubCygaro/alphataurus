package assembler

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/JakubCygaro/alphataurus/internal/vm"
)

type labelData struct {
	pos        uint64
	declaredAt string
}
type labelMap map[string]labelData
type unresolvedJumpMap map[int]string

type Assembler struct {
	parser  Parser
	opCodes map[uint32]vm.OpCodeVal
	unresolvedJumps unresolvedJumpMap
	labels  labelMap
}

func NewAssembler(reader bufio.Reader) Assembler {
	return Assembler{
		parser:  NewParser(reader),
		opCodes: vm.GenerateOpcodeMap(),
		labels: make(labelMap),
		unresolvedJumps: make(unresolvedJumpMap),
	}
}

func (a *Assembler) EmitBytecode() ([]byte, int, error) {
	bytecode := make([]byte, 0, 64)
	instCount := 0
	var ok bool
	var err error = nil
	ok, err = a.parser.ParseNext()
	for ; ok && err == nil; ok, err = a.parser.ParseNext() {
		inst := a.parser.CurrentInst()
		switch inst.Ty {
		case INST_TMOVIR:
			err = a.emitMovIR(inst.Data.(InstMovData), &bytecode)
		case INST_TMOVRR:
			err = a.emitMovRR(inst.Data.(InstMovData), &bytecode)
		case INST_TADDRR:
			err = a.emitArthRR(int(inst.Ty), inst.Data.(InstArthData), &bytecode)
		case INST_TSUBRR:
			err = a.emitArthRR(int(inst.Ty), inst.Data.(InstArthData), &bytecode)
		case INST_TMULRR:
			err = a.emitArthRR(int(inst.Ty), inst.Data.(InstArthData), &bytecode)
		case INST_TDIVRR:
			err = a.emitArthRR(int(inst.Ty), inst.Data.(InstArthData), &bytecode)
		case INST_TADDIR:
			err = a.emitArthIR(int(inst.Ty), inst.Data.(InstArthData), &bytecode)
		case INST_TSUBIR:
			err = a.emitArthIR(int(inst.Ty), inst.Data.(InstArthData), &bytecode)
		case INST_TINCR:
			err = a.emitInc(inst.Data.(InstIncDecData), &bytecode)
		case INST_TDECR:
			err = a.emitDec(inst.Data.(InstIncDecData), &bytecode)
		case INST_TCMPRR:
			err = a.emitCmpRR(inst.Data.(InstCmpData), &bytecode)
		case INST_TCMPIR:
			err = a.emitCmpIR(inst.Data.(InstCmpData), &bytecode)
		case INST_TJMP:
			err = a.emitJmp(int(inst.Ty), inst.Data.(InstJmpData), &bytecode)
		case INST_TJMPE:
			err = a.emitJmp(int(inst.Ty), inst.Data.(InstJmpData), &bytecode)
		case INST_TJMPNE:
			err = a.emitJmp(int(inst.Ty), inst.Data.(InstJmpData), &bytecode)
		case INST_TJMPZ:
			err = a.emitJmp(int(inst.Ty), inst.Data.(InstJmpData), &bytecode)
		case INST_TJMPNZ:
			err = a.emitJmp(int(inst.Ty), inst.Data.(InstJmpData), &bytecode)
		case INST_TJMPG:
			err = a.emitJmp(int(inst.Ty), inst.Data.(InstJmpData), &bytecode)
		case INST_TJMPGE:
			err = a.emitJmp(int(inst.Ty), inst.Data.(InstJmpData), &bytecode)
		case INST_TJMPL:
			err = a.emitJmp(int(inst.Ty), inst.Data.(InstJmpData), &bytecode)
		case INST_TJMPLE:
			err = a.emitJmp(int(inst.Ty), inst.Data.(InstJmpData), &bytecode)
		case INST_TLABEL:
			err = a.declareLabel(inst.Data.(InstLabData), &bytecode)
		default:
			pos := a.parser.lexer.CurrentPosition()
			return bytecode, 0, fmt.Errorf("Instruction (%d) WIP %s", INST_TMOVIR, pos)
		}
		if err != nil {
			return bytecode, instCount, err
		}
		instCount++
	}
	if err != nil {
		return bytecode, instCount, err
	}
	err = a.resolveJumpInsturctions(&bytecode)
	return bytecode, instCount, err
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
	*out = append(*out, byte(data.Src))
	*out = append(*out, byte(data.Dest))
	*out = append(*out, byte(0))
	*out = append(*out, byte(data.Ty))
	*out = append(*out, 0, 0, 0, 0)
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
	case INST_TJMPG:
		opcode = a.opCodes[vm.OP_JMPG]
	}
	opPos := len(*out)
	*out = binary.BigEndian.AppendUint32(*out, uint32(opcode))
	if addr, ok := data.Address.(uint64); ok {
		*out = binary.BigEndian.AppendUint64(*out, uint64(addr))
	} else if lab, ok := data.Address.(string); ok {
		if l, ok := a.labels[lab]; ok {
			*out = binary.BigEndian.AppendUint64(*out, uint64(l.pos/vm.INSTRUCTION_SIZE))
		} else {
			a.unresolvedJumps[opPos] = lab
			*out = binary.BigEndian.AppendUint64(*out, uint64(0))
		}
	} else {
		return fmt.Errorf("Jump instruction bad internal assembler data")
	}
	return nil
}
func (a *Assembler) declareLabel(data InstLabData, out *[]byte) error {
	if lab, ok := a.labels[data.Label]; ok {
		return fmt.Errorf("Label '%s' redeclared at %s, first declared at %s",
			data.Label, data.DeclaredAt, lab.declaredAt)
	}
	a.labels[data.Label] = labelData{
		pos:        uint64(len(*out)),
		declaredAt: data.DeclaredAt,
	}
	return nil
}
func (a *Assembler) resolveJumpInsturctions(out *[]byte) error {
	for codePos, destLabel := range a.unresolvedJumps {
		label, ok := a.labels[destLabel]
		if !ok {
			return fmt.Errorf("Could not resolve label '%s'", destLabel)
		}
		binary.BigEndian.PutUint64((*out)[codePos+vm.OPCODE_SIZE:], label.pos/vm.INSTRUCTION_SIZE)
	}
	return nil
}
