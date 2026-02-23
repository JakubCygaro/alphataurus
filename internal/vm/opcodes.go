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
	ty int8
	// handle HandleFunc
	op     OpCodeVal
	nested OpCodeMap
}
type OpCodeMap map[int8]OpMapVal

func handle(op OpCodeVal) OpMapVal {
	return OpMapVal{
		ty: HANDLE,
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
	//arth
	p002X = nested(OpCodeMap{
		0: handle(OP_ADDRR),
	})
	//jumps
	p003X = nested(OpCodeMap{
		0: handle(OP_JMP),
		1: handle(OP_JMPE),
	})
	p00XX = nested(OpCodeMap{
		0: handle(OP_MOVRR),
		1: handle(OP_MOVIR),
		2: p002X,
		3: p003X,
	})
	p0XXX = nested(OpCodeMap{
		0: p00XX,
	})
	oPCODE_MAP = OpCodeMap{
		0: p0XXX,
	}
)

func GetOpcode(opcodebytes []byte) (OpCodeVal, error) {
	ptr := 3
	opmap := oPCODE_MAP
	for {
		v, ok := opmap[int8(opcodebytes[ptr])]
		if !ok {
			cd := binary.BigEndian.Uint32(opcodebytes)
			return 0, fmt.Errorf("Bad opcode 0x%08x (%032b)", cd, cd)
		}
		switch v.ty {
		case HANDLE:
			return v.op, nil
		case NEST:
			ptr--
			opmap = v.nested
		}
	}

}

const (
	OPCODE_SIZE      = 4                           // opcode size in bytes (32-bits)
	ARGUMENT_SIZE    = 8                           // argument size in bytes (64-bits)
	INSTRUCTION_SIZE = OPCODE_SIZE + ARGUMENT_SIZE // size of a single instruction in bytes
)
const (
	OP_MOVIR = 0 // move imediate value to register
	OP_MOVRR = iota // move register to register

	OP_ADDRR = iota // add register to register and store into second register, singedness and registers passed in parameter

	OP_JMP      = iota // jump to instruction
	OP_JMPE     = iota // jump if equal
	OP_TESTR0R1 = iota // test registers
)

func recurseIntoOpCodeMap(layer int, opcodes *OpCodeMap, bytes []byte, ret *map[uint32]OpCodeVal) {
		current := opcodes
		for k, v := range *current {
			bytes[layer] = byte(k)
			if v.ty == HANDLE {
				(*ret)[uint32(v.op)] = OpCodeVal(binary.BigEndian.Uint32(bytes))
			} else {
				recurseIntoOpCodeMap(layer-1, &v.nested, bytes, ret)
			}
		}
	}


func GenerateOpcodeMap() map[uint32]OpCodeVal {
	ret := make(map[uint32]OpCodeVal)
	bytes := make([]byte, 4)
	recurseIntoOpCodeMap(3, &oPCODE_MAP, bytes, &ret)
	return ret
}
