package vm

import (
	"encoding/binary"
	"github.com/JakubCygaro/alphataurus/pkg/vm/errors"
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
	nested opCodeMap
}
type opCodeMap map[int8]OpMapVal

func handle(op OpCodeVal) OpMapVal {
	return OpMapVal{
		ty: HANDLE,
		op: op,
	}
}
func nested(nest opCodeMap) OpMapVal {
	return OpMapVal{
		ty:     NEST,
		nested: nest,
	}
}

var (
	//arth
	p002X = nested(opCodeMap{
		0:  handle(OP_ADDRR),
		1:  handle(OP_INCR),
		2:  handle(OP_SUBRR),
		3:  handle(OP_MULRR),
		4:  handle(OP_DIVRR),
		5:  handle(OP_DECR),
		6:  handle(OP_NOP),
		7:  handle(OP_CLR),
		8:  handle(OP_CALL),
		9:  handle(OP_RET),
		10: handle(OP_NOT),
		11: handle(OP_ORRR),
		12: handle(OP_ANDRR),
		13: handle(OP_XORRR),
		14: handle(OP_LSHRR),
		15: handle(OP_RSHRR),
		16: handle(OP_EXITI),
		17: handle(OP_EXITR),
		19: handle(OP_MOVSB),
		20: handle(OP_MOVREPSB),
		21: handle(OP_MOVSQ),
		22: handle(OP_MOVREPSQ),
		23: handle(OP_MOVSH),
		24: handle(OP_MOVREPSH),
		25: handle(OP_MOVSW),
		26: handle(OP_SDF),
		27: handle(OP_CDF),
		28: handle(OP_NEG),
	})
	//jumps
	//direct jumps
	p030X = nested(opCodeMap{
		0:  handle(OP_JMP),
		1:  handle(OP_JMPE),
		2:  handle(OP_JMPNE),
		3:  handle(OP_JMPZ),
		4:  handle(OP_JMPNZ),
		5:  handle(OP_JMPG),
		6:  handle(OP_JMPGE),
		7:  handle(OP_JMPL),
		8:  handle(OP_JMPLE),
		9:  handle(OP_JMPS),
		10: handle(OP_JMPNS),
		11: handle(OP_JMPC),
		12: handle(OP_JMPNC),
		13: handle(OP_JMPO),
		14: handle(OP_JMPNO),
	})
	p03XX = nested(opCodeMap{
		0: p030X,
		//ip relative jumps
		1:  handle(OP_JMPIP),
		2:  handle(OP_JMPEIP),
		3:  handle(OP_JMPNEIP),
		4:  handle(OP_JMPZIP),
		5:  handle(OP_JMPNZIP),
		6:  handle(OP_JMPGIP),
		7:  handle(OP_JMPGEIP),
		8:  handle(OP_JMPLIP),
		9:  handle(OP_JMPLEIP),
		10: handle(OP_JMPSIP),
		11: handle(OP_JMPNSIP),
		12: handle(OP_JMPCIP),
		13: handle(OP_JMPNCIP),
		14: handle(OP_JMPOIP),
		15: handle(OP_JMPNOIP),
		16: handle(OP_CALLIP),
	})
	p00XX = nested(opCodeMap{
		0:  handle(OP_MOVRR),
		1:  handle(OP_MOVIR),
		2:  p002X,
		4:  handle(OP_CMPRR),
		5:  handle(OP_ADDIR),
		6:  handle(OP_SUBIR),
		7:  handle(OP_MOVDR),
		8:  handle(OP_ORIR),
		9:  handle(OP_ANDIR),
		10: handle(OP_XORIR),
		11: handle(OP_LSHIR),
		12: handle(OP_RSHIR),
		13: handle(OP_CMPIR),
		14: handle(OP_MOVSXRR),
		15: handle(OP_MOVZXRR),
		16: handle(OP_MOVZXIR),
		17: handle(OP_MOVZXDR),
		18: handle(OP_XCHGRR),
		19: handle(OP_XCHGDR),
		20: handle(OP_ROL),
		21: handle(OP_ROR),
	})
	p01XX = nested(opCodeMap{
		// stack manipulation, leave a byte for data size
		0: handle(OP_PUSHR),
		1: handle(OP_PUSHI),
		2: handle(OP_POP),
		//
		3: handle(OP_MOVID),
		4: handle(OP_MOVRD),
	})
	p1XXX = nested(opCodeMap{
		1: handle(OP_MOVDRO1),
		3: handle(OP_MOVRDO1),
		4: handle(OP_MOVIDO1_NO),
		5: handle(OP_MOVIDO1),
		6: handle(OP_MOVZXDRO1),
		7: handle(OP_XCHGDRO1),
	})
	p0XXX = nested(opCodeMap{
		0: p00XX,
		1: p01XX,
		3: p03XX,
	})
	oPCODE_MAP = opCodeMap{
		0: p0XXX,
		1: p1XXX,
		2: handle(OP_MOVIDO2_NO),
		3: handle(OP_MOVIDO2),
		4: handle(OP_MOVDRO2),
		5: handle(OP_MOVZXDRO2),
		6: handle(OP_XCHGDRO2),
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

//go:generate stringer -type=OpCodeVal
const (
	OP_MOVIR      OpCodeVal = iota // move imediate value to register
	OP_MOVRR                       // move register to register
	OP_MOVDR                       // move dereference to register, [<address>]
	OP_MOVDRO1                     // move dereference to register, like [rx + <signed offset>]
	OP_MOVDRO2                     // move dereference to register, like [(rx + rx) +/- <signed offset>]
	OP_MOVID                       // move immediate value to deref
	OP_MOVRD                       // move register value into deref
	OP_MOVRDO1                     // move register value into deref with one offset register
	OP_MOVRDO2                     // move register value into deref with two offset registers
	OP_MOVIDO1_NO                  // move immediate value to deref, offset = 0
	OP_MOVIDO1                     // move immediate value to deref with one offset register
	OP_MOVIDO2_NO                  // move immediate value to deref with two offset registers, offset = 0
	OP_MOVIDO2                     // move immediate value to deref with two offset registers
	// move with sign extend
	OP_MOVSXRR
	// move with zero extend
	OP_MOVZXRR
	OP_MOVZXIR
	OP_MOVZXDR
	OP_MOVZXDRO1
	OP_MOVZXDRO2
	OP_MOVSB
	OP_MOVREPSB
	OP_MOVSQ
	OP_MOVREPSQ
	OP_MOVSH
	OP_MOVREPSH
	OP_MOVSW
	OP_MOVREPSW
	OP_XCHGRR
	OP_XCHGDR
	OP_XCHGDRO1
	OP_XCHGDRO2
	OP_LEAO1
	OP_LEAO2
	OP_ADDRR // add register to register and store into second register, singedness and registers passed in parameter
	OP_ADDIR
	OP_SUBRR
	OP_SUBIR
	OP_MULRR
	OP_DIVRR
	OP_NOT
	OP_NEG
	OP_ROL
	OP_ROR
	OP_ANDRR
	OP_ANDIR
	OP_ORRR
	OP_ORIR
	OP_XORRR
	OP_XORIR
	OP_LSHRR
	OP_LSHIR
	OP_RSHRR
	OP_RSHIR
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
	OP_JMPS
	OP_JMPNS
	OP_JMPC
	OP_JMPNC
	OP_JMPO
	OP_JMPNO
	OP_JMPIP   // jump to instruction
	OP_JMPEIP  // jump if equal
	OP_JMPZIP  // jump if zero
	OP_JMPNEIP // jump if not equal
	OP_JMPNZIP // jump of not zero
	OP_JMPGIP  // jump if greater
	OP_JMPGEIP // jump if greater or equal
	OP_JMPLIP  // jump if less
	OP_JMPLEIP // jump if less or equal
	OP_JMPSIP
	OP_JMPNSIP
	OP_JMPCIP
	OP_JMPNCIP
	OP_JMPOIP
	OP_JMPNOIP
	OP_CMPRR  // compare values in two registers
	OP_CMPIR  // compare value in register and immediate value
	OP_CLR    // clear all flags (set them to false)
	OP_CALL   // call a procedure
	OP_CALLIP // IP relative call
	OP_RET    // return from a procedure
	OP_EXITI  // exit with code
	OP_EXITR
	OP_SDF
	OP_CDF
	OP_NOP // NOP
)

const (
	// last byte for OP_CMP that indicates an immediate value comparasion
	OPLB_CMP_IM = iota
	// last byte for OP_CMP that indicates a register-register cmp
	OPLB_CMP_RR = iota
)

const (
	OP_TADD = iota
	OP_TSUB
	OP_TMUL
	OP_TDIV
)

func recurseIntoOpCodeMap(layer int, opcodes *opCodeMap, bytes []byte, ret *OpCodeMap) {
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

type OpCodeMap map[uint32]OpCodeVal

func GenerateOpcodeMap() OpCodeMap {
	ret := make(OpCodeMap)
	bytes := make([]byte, 4)
	recurseIntoOpCodeMap(3, &oPCODE_MAP, bytes, &ret)
	return ret
}
func (m *OpCodeMap) GetBytes(op OpCodeVal) uint32 {
	return uint32((*m)[uint32(op)])
}
