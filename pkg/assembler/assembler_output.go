package assembler

import (
	"encoding/binary"

	"github.com/JakubCygaro/alphataurus/pkg/assembler/errors"
	aobj "github.com/JakubCygaro/alphataurus/pkg/vm/obj"
)

func (a *Assembler) Assemble() ([]byte, error) {
	var ok bool
	var err error
	ok, err = a.parser.ParseNext()
	for ; ok && err == nil; ok, err = a.parser.ParseNext() {
		inst := a.parser.CurrentInst()
		a.line = inst.Line
		a.col = inst.Col
		switch i := inst.Data.(type) {
		case InstExport:
			if err := a.handleExport(i); err != nil {
				return nil, err
			}
		case InstImport:
			if err := a.handleImport(i); err != nil {
				return nil, err
			}
		case InstSecCode:
			instCount, err := a.EmitBytecode()
			if err != nil {
				return nil, err
			}
			a.instCount += instCount
		default:
			return nil, errors.DisallowedTopLevelInstruction(a.line, a.col)
		}
	}
	if err != nil {
		return nil, err
	}
	if err := a.resolveJumpInsturctions(); err != nil {
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
func (a *Assembler) writeObjFile() ([]byte, error) {
	var syms, rels []byte
	if symbols, err := aobj.WriteSymbols(&a.symbols); err != nil {
		return nil, err
	} else {
		syms = symbols
	}
	if relocs, err := aobj.WriteRelocs(a.relocations); err != nil {
		return nil, err
	} else {
		rels = relocs
	}
	head := aobj.ObjFileHeader{}
	head.StaticDataSize = 0
	head.SymbolsStart = 0
	head.SymbolsSize = uint64(len(syms))
	head.RelocsStart = head.SymbolsStart + head.SymbolsSize
	head.RelocsSize = uint64(len(rels))
	// code is the last thing in the output
	head.CodeStart = head.RelocsStart + head.RelocsSize
	head.CodeSize = uint64(len(a.bytecode))

	output := make([]byte, 0)

	//mag
	output = append(output, aobj.OBJ_FILE_MAG...)

	//version
	output = binary.BigEndian.AppendUint32(output, 0x00000001)

	//header size
	output = binary.BigEndian.AppendUint16(output, uint16(0))

	preambleEnd := len(output)

	if a.hasEntry {
		//entry point
		output = append(output, aobj.OBJ_FILE_ENTRY)
		output = binary.BigEndian.AppendUint64(output, a.entry)
	}

	output = append(output, aobj.OBJ_FILE_SECCODE)
	//code start
	output = binary.BigEndian.AppendUint64(output, head.CodeStart)
	//code size
	output = binary.BigEndian.AppendUint64(output, head.CodeSize)

	//static data
	output = append(output, aobj.OBJ_FILE_SECSDATA)
	output = binary.BigEndian.AppendUint64(output, head.StaticDataStart)
	output = binary.BigEndian.AppendUint64(output, head.StaticDataSize)

	output = append(output, aobj.OBJ_FILE_SECSYMS)
	//syms start
	output = binary.BigEndian.AppendUint64(output, head.SymbolsStart)
	//syms size
	output = binary.BigEndian.AppendUint64(output, head.SymbolsSize)

	output = append(output, aobj.OBJ_FILE_SECRELS)
	//reloc start
	output = binary.BigEndian.AppendUint64(output, head.RelocsStart)
	//reloc size
	output = binary.BigEndian.AppendUint64(output, head.RelocsSize)

	headerSize := len(output) - preambleEnd
	binary.BigEndian.PutUint16(output[4+4:], uint16(headerSize))

	dataSize := head.CodeStart + head.CodeSize

	dataStartSlice := make([]byte, dataSize)

	//dump code
	codeSlice := dataStartSlice[head.CodeStart : head.CodeStart+head.CodeSize]
	copy(codeSlice, a.bytecode)
	//dump symbols
	symSlice := dataStartSlice[head.SymbolsStart : head.SymbolsStart+head.SymbolsSize]
	copy(symSlice, syms)
	//dump relocs
	relSlice := dataStartSlice[head.RelocsStart : head.RelocsStart+head.RelocsSize]
	copy(relSlice, rels)
	output = append(output, dataStartSlice...)
	return output, nil
}
