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
		case INST_TEXPORT:
			if err := a.handleExport(inst.Data.(InstImportExportData)); err != nil {
				return nil, err
			}
		case INST_TIMPORT:
			if err := a.handleImport(inst.Data.(InstImportExportData)); err != nil {
				return nil, err
			}
		case INST_TSECCODE:
			instCount, err := a.EmitBytecode()
			if err != nil {
				return nil, err
			}
			a.instCount += instCount
		default:
			return nil, fmt.Errorf("Disallowed top level instruction")
		}
	}
	if err != nil {
		return nil, err
	}
	if err := a.resolveJumpInsturctions(&(a.bytecode)); err != nil {
		return nil, err
	}
	if err := a.resolveSymbols(); err != nil {
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
// relocsStart 8b
// relocsEnd 8b
// free space up to 128 bytes (for now)
func (a *Assembler) writeObjFile() ([]byte, error) {
	var syms []byte
	if symbols, err := writeSymbols(&a.symbols); err != nil {
		return nil, err
	} else {
		syms = symbols
	}

	output := make([]byte, 0, OBJ_FILE_HEADER_SIZE)
	//mag
	output = append(output, OBJ_FILE_MAG...)
	//version
	output = binary.BigEndian.AppendUint32(output, 0x00000001)
	//code start
	output = binary.BigEndian.AppendUint64(output, uint64(len(syms)))
	//code size
	output = binary.BigEndian.AppendUint64(output, uint64(len(a.bytecode)))
	//static data
	output = binary.BigEndian.AppendUint64(output, 0)
	output = binary.BigEndian.AppendUint64(output, 0)

	//syms start
	output = binary.BigEndian.AppendUint64(output, 0)
	//syms size
	output = binary.BigEndian.AppendUint64(output, uint64(len(syms)))

	//pad with zeros
	output = output[:OBJ_FILE_HEADER_SIZE]

	//dump symbols
	output = append(output, syms...)
	//dump code
	output = append(output, a.bytecode...)
	return output, nil
}
