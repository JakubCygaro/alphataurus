package assembler

import "encoding/binary"

func (a *Assembler) Assemble() ([]byte, error) {
	var ok bool
	var err error
	ok, err = a.parser.ParseNext()
	for ; ok && err == nil; ok, err = a.parser.ParseNext() {
		inst := a.parser.CurrentInst()
		switch inst.Ty {
		case INST_TSECCODE:
			a.EmitBytecode()
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
	output = append(output, OBJ_FILE_MAG...)
	binary.BigEndian.AppendUint32(output, 0x00000001)
	binary.BigEndian.AppendUint64(output, OBJ_FILE_HEADER_SIZE)
	binary.BigEndian.AppendUint64(output, uint64(len(a.bytecode)))
	output = output[:OBJ_FILE_HEADER_SIZE]
	output = append(output, a.bytecode...)
	return output, nil
}
