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
// header_size 2b (excluding version, magic and itself)
// codeSecStart 8b
// codeSecLen 8b
// staticDataStart 8b
// staticDataSize 8b
// symbolsStart 8b
// symbolsSize 8b
// relocsStart 8b
// relocsSize 8b
func (a *Assembler) writeObjFile() ([]byte, error) {
	var syms, rels []byte
	if symbols, err := WriteSymbols(&a.symbols); err != nil {
		return nil, err
	} else {
		syms = symbols
	}
	if relocs, err := WriteRelocs(a.relocations); err != nil {
		return nil, err
	} else {
		rels = relocs
	}
	head := ObjFileHeader{}
	head.StaticDataSize = 0
	head.SymbolsStart = 0
	head.SymbolsSize = uint64(len(syms))
	head.RelocsStart = head.SymbolsStart + head.SymbolsSize
	head.RelocsSize = uint64(len(rels))
	// code is the last thing in the output
	head.CodeStart = head.RelocsStart + head.RelocsSize
	head.CodeSize = uint64(len(a.bytecode))

	const headerSize = 8 * 8

	output := make([]byte, 0, headerSize+head.CodeStart+head.CodeSize)

	//mag
	output = append(output, OBJ_FILE_MAG...)

	//version
	output = binary.BigEndian.AppendUint32(output, 0x00000001)

	//header size
	output = binary.BigEndian.AppendUint16(output, headerSize)

	//code start
	output = binary.BigEndian.AppendUint64(output, head.CodeStart)
	//code size
	output = binary.BigEndian.AppendUint64(output, head.CodeSize)

	//static data
	output = binary.BigEndian.AppendUint64(output, head.StaticDataStart)
	output = binary.BigEndian.AppendUint64(output, head.StaticDataSize)

	//syms start
	output = binary.BigEndian.AppendUint64(output, head.SymbolsStart)
	//syms size
	output = binary.BigEndian.AppendUint64(output, head.StaticDataSize)

	//reloc start
	output = binary.BigEndian.AppendUint64(output, head.RelocsStart)
	//reloc size
	output = binary.BigEndian.AppendUint64(output, head.RelocsSize)
	//pad with zeros
	output = output[:cap(output)]

	dataStartSlice := output[headerSize:]

	//dump code
	codeSlice := dataStartSlice[head.CodeStart : head.CodeStart+head.CodeSize]
	copy(codeSlice, a.bytecode)
	//dump symbols
	symSlice := dataStartSlice[head.SymbolsStart : head.SymbolsStart+head.SymbolsSize]
	copy(symSlice, syms)
	//dump relocs
	relSlice := dataStartSlice[head.RelocsStart : head.RelocsStart+head.RelocsSize]
	copy(relSlice, rels)
	return output, nil
}
