package assembler

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/JakubCygaro/alphataurus/internal/vm"
)

type Assembler struct {
	parser  Parser
	opCodes map[uint32]vm.OpCodeVal
}

func NewAssembler(reader bufio.Reader) Assembler {
	return Assembler{
		parser:  NewParser(reader),
		opCodes: vm.GenerateOpcodeMap(),
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
			err = a.emitAddRR(inst.Data.(InstAddData), &bytecode)
		default:
			pos := a.parser.lexer.CurrentPosition()
			return bytecode, 0, fmt.Errorf("Instruction (%d) WIP %s", INST_TMOVIR, pos)
		}
		instCount++
	}
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

func (a *Assembler) emitAddRR(data InstAddData, out *[]byte) error {
	add := a.opCodes[vm.OP_ADDRR]
	*out = binary.BigEndian.AppendUint32(*out, uint32(add))
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
