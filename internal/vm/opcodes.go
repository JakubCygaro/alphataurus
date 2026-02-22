package vm

import (
	"encoding/binary"
	"fmt"
)

// make it a bigass map

type HandleFunc func(*VmState)

const (
	HANDLE = 0
	NEST   = iota
)

type OpCodeVal uint32

type OpMapVal struct {
	ty     int8
	// handle HandleFunc
	op     OpCodeVal
	nested OpCodeMap
}
type OpCodeMap map[int8]OpMapVal

func handle(op OpCodeVal) OpMapVal {
	return OpMapVal{
		ty:     HANDLE,
		op: op,
	}
}
func nested(nest OpCodeMap) OpMapVal {
	return OpMapVal{
		ty:     NEST,
		nested: nest,
	}
}

var (
	p000X = nested(OpCodeMap{
		0: handle(OP_MOVIR0),
	})
	p00XX = nested(OpCodeMap{
		0: p000X,
	})
	p0XXX = nested(OpCodeMap{
		0: p00XX,
	})
	oPCODE_MAP   = OpCodeMap{
		0: p0XXX,
	}
)

func GetOpcode(opcode uint32) (OpCodeVal, error) {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, opcode)
	ptr := 0
	for {
		v, ok := oPCODE_MAP[int8(bytes[ptr])]
		if !ok {
			return 0, fmt.Errorf("Bad opcode")
		}
		switch v.ty{
		case HANDLE:
			return v.op, nil
		case NEST:
			ptr++
		}
	}

}

const (
	OPCODE_SIZE      = 4                           // opcode size in bytes (32-bits)
	ARGUMENT_SIZE    = 8                           // argument size in bytes (64-bits)
	INSTRUCTION_SIZE = OPCODE_SIZE + ARGUMENT_SIZE // size of a single instruction in bytes

	OP_MOVIR0 = 0x00_01_00_00 // move imediate value to register 0
	OP_MOVIR1 = 0x00_01_00_01 // move imediate value to register 1

	OP_ADDUR0R1 = 0x00_02_00_00 // add register to register and store into second register UNSINGED
	OP_ADDSR0R1 = 0x00_02_00_01 // add register to register and store into second register SINGED

	OP_JMP      = 0x00_10_00_00 // jump to instruction
	OP_JMPE     = 0x00_10_00_01 // jump if equal
	OP_TESTR0R1 = 0x00_11_00_00 // test registers
)
