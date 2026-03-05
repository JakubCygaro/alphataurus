package vm

import (
	"encoding/binary"
	"github.com/JakubCygaro/alphataurus/internal/vm/errors"
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
		1: handle(OP_INCR),
		2: handle(OP_SUBRR),
		3: handle(OP_MULRR),
		4: handle(OP_DIVRR),
		5: handle(OP_DECR),
	})
	//jumps
	p03XX = nested(OpCodeMap{
		0: handle(OP_JMP),
		1: handle(OP_JMPE),
		2: handle(OP_JMPG),
	})
	p00XX = nested(OpCodeMap{
		0: handle(OP_MOVRR),
		1: handle(OP_MOVIR),
		2: p002X,
		4: handle(OP_CMP),
		5: handle(OP_ADDIR),
		6: handle(OP_SUBIR),
		7: handle(OP_MOVDRI),
	})
	//stack
	p011X = nested(OpCodeMap{
		0: handle(OP_PUSHR),
		1: handle(OP_PUSHI),
		2: handle(OP_POP),
	})
	p01XX = nested(OpCodeMap{
		1: p011X,
	})
	p1XXX = nested(OpCodeMap{
		0: handle(OP_MOVDRO1),
	})
	p0XXX = nested(OpCodeMap{
		0: p00XX,
		1: p01XX,
		3: p03XX,
	})
	oPCODE_MAP = OpCodeMap{
		0: p0XXX,
		1: p1XXX,
	}
)

func (state *VmState) GetOpcode(opcodebytes []byte) (OpCodeVal, error) {
	ptr := 3
	opmap := oPCODE_MAP
	for {
		v, ok := opmap[int8(opcodebytes[ptr])]
		if !ok {
			cd := binary.BigEndian.Uint32(opcodebytes)
			return 0, errors.BadOpcode(cd, state.byteCodePos)
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
	OP_MOVIR   = 0    // move imediate value to register
	OP_MOVRR   = iota // move register to register
	OP_MOVDRI         // move dereference to register, [<address>]
	OP_MOVDRO1        // move dereference to register, like [rx + <signed offset>]
	OP_MOVDRO2        // move dereference to register, like [rx <op> rx + <signed offset>]
	OP_ADDRR          // add register to register and store into second register, singedness and registers passed in parameter
	OP_ADDIR
	OP_SUBRR
	OP_SUBIR
	OP_MULRR
	OP_DIVRR
	OP_INCR
	OP_DECR

	OP_PUSHR
	OP_PUSHI
	OP_POP

	OP_JMP   // jump to instruction
	OP_JMPE  // jump if equal
	OP_JMPZ  // jump if zero
	OP_JMPNE // jump if not equal
	OP_JMPNZ // jump of not zero
	OP_JMPG  // jump if greater
	OP_JMPGE // jump if greater or equal
	OP_JMPL  // jump if less
	OP_JMPLE // jump if less or equal
	OP_CMP   // test registers
)

const (
	// last byte for OP_CMP that indicates an immediate value comparasion
	OPLB_CMP_IM = iota
	// last byte for OP_CMP that indicates a register-register cmp
	OPLB_CMP_RR = iota
)

const (
	OP_TADD = iota
	OP_TSUBRI
	OP_TSUBIR
	OP_TMUL
	OP_TDIVRI
	OP_TDIVIR
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
