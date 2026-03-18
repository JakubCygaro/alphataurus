package assembler

import (
	"encoding/binary"
	"fmt"
)

func (a *Assembler) Assemble() ([]byte, error) {
	var ok bool
	var err error
	ok, err = a.parser.ParseNext()
	for ; ok && err == nil; ok, err = a.parser.ParseNext() {
		inst := a.parser.CurrentInst()
		switch inst.Ty {
		case INST_TSECCODE:
			instCount, err := a.EmitBytecode()
			if err != nil {
				return nil, nil
			}
			a.instCount += instCount
		default:
			return nil, fmt.Errorf("Disallowed top level instruction")
		}
	}
	if err := a.resolveJumpInsturctions(&(a.bytecode)); err != nil {
		return nil, err
	}
	return a.writeObjFile()
}

// Obj file header
// AELF 4b
// version 4b
// codeSecStart 8b
// codeSecLen 8b
// staticDataStart 8b
// staticDataSize 8b
// symbolsStart 8b
// symbolsSize 8b
// free space up to 128 bytes (for now)
func (a *Assembler) writeObjFile() ([]byte, error) {
	output := make([]byte, 0, OBJ_FILE_HEADER_SIZE)
	//mag
	output = append(output, OBJ_FILE_MAG...)
	//version
	output = binary.BigEndian.AppendUint32(output, 0x00000001)
	//code start
	output = binary.BigEndian.AppendUint64(output, 0)
	//code size
	output = binary.BigEndian.AppendUint64(output, uint64(len(a.bytecode)))
	//pad with zeros
	output = output[:OBJ_FILE_HEADER_SIZE]
	//dump code
	output = append(output, a.bytecode...)
	fmt.Printf("%+v\n", output)
	return output, nil
}
