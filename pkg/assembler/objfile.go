package assembler

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

const (
	OBJ_FILE_MAG         = "AOBJ"
	OBJ_FILE_HEADER_SIZE = 128
)

type ObjFileHeader struct {
	Version                         uint32
	CodeStart, CodeSize             uint64
	StaticDataStart, StaticDataSize uint64
	SymbolsStart, SymbolsSize       uint64
}

type StaticDataTable map[string]StaticData

type StaticData struct {
	Ty  int
	Val []uint64
}

const (
	SYM_TFUNC = iota
	SYM_TSTATVAR
)

const (
	/// defined in file
	SYM_VEXPORT = iota
	/// required by the file but not defined
	SYM_VIMPORTSTRONG
	/// weak reference in the file
	SYM_VIMPORTWEAK
	/// defined by the file but not linkable from the outside
	SYM_VPRIVATE
)

type SymbolTable map[string]SymbolData

type SymbolData struct {
	Ty  byte
	Vis byte
	// the location is defined with the deadzone added, so any value below the deadzone is treated as invalid
	Loc uint64
}

// reloc
// 8b(LOC) 8b(REF) 1b(SIZE)

const (
	RELOC_TINVALID = 0
)

type RelocData struct {
	// where that symbol is referenced in the code
	Loc uint64
	// what symbol is being referenced
	Ref uint64
	PatchSize byte
}

type RelocationTable []RelocData

type ObjFile struct {
	Header     ObjFileHeader
	Code       []byte
	StaticData StaticDataTable
	Symbols    SymbolTable
	Relocs     RelocationTable
}

// Obj file header
// AOBJ 4b
// version 4b
// codeSecStart 8b
// codeSecLen 8b
// staticDataStart 8b
// staticDataSize 8b
// symbolsStart 8b
// symbolsSize 8b
// free space up to 128 bytes (for now)

func LoadObjFileHeader(reader *bufio.Reader) (ObjFileHeader, error) {
	ret := ObjFileHeader{}
	mag := [4]byte{}
	if n, err := io.ReadFull(reader, mag[:]); err != nil {
		return ret, err
	} else if n != 4 {
		return ret, fmt.Errorf("Object file too short")
	} else if string(mag[:]) != OBJ_FILE_MAG {
		return ret, fmt.Errorf("Not an object file")
	}
	buf := make([]byte, 8)
	if _, err := io.ReadFull(reader, buf[:4]); err != nil {
		return ret, fmt.Errorf("Object file too short")
	}
	ret.Version = binary.BigEndian.Uint32(buf)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return ret, fmt.Errorf("Object file too short")
	}
	ret.CodeStart = binary.BigEndian.Uint64(buf)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return ret, fmt.Errorf("Object file too short")
	}
	ret.CodeSize = binary.BigEndian.Uint64(buf)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return ret, fmt.Errorf("Object file too short")
	}
	ret.StaticDataStart = binary.BigEndian.Uint64(buf)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return ret, fmt.Errorf("Object file too short")
	}
	ret.StaticDataSize = binary.BigEndian.Uint64(buf)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return ret, fmt.Errorf("Object file too short")
	}
	ret.SymbolsStart = binary.BigEndian.Uint64(buf)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return ret, fmt.Errorf("Object file too short")
	}
	ret.SymbolsSize = binary.BigEndian.Uint64(buf)
	return ret, nil
}
func loadSymbols(symbolSec []byte) (SymbolTable, error) {
	// 1b(TY) 1b(VISIBILITY) 8b(LOC) 4b(NAMELEN) NAMELENb(NAME)
	table := make(SymbolTable, 0)
	reader := bufio.NewReader(bytes.NewReader(symbolSec))
	buf := make([]byte, 0, 64)
	for {
		var vis byte
		var loc uint64
		var namelen uint32
		var name string
		first, err := reader.ReadByte()
		if err != nil {
			return table, nil
		}
		ty := first
		if ty > SYM_TSTATVAR {
			return nil, fmt.Errorf("Invalid symbol type")
		}
		if b, err := reader.ReadByte(); err != nil {
			return nil, err
		} else {
			vis = b
		}
		if _, err := reader.Read(buf[:8]); err != nil {
			return nil, err
		} else {
			loc = binary.BigEndian.Uint64(buf[:8])
		}
		if _, err := reader.Read(buf[:4]); err != nil {
			return nil, err
		} else {
			namelen = binary.BigEndian.Uint32(buf[:4])
		}
		if extendBy := int(namelen) - len(buf); extendBy > 0 {
			buf = make([]byte, len(buf)+extendBy)
		}
		if _, err := reader.Read(buf[:namelen]); err != nil {
			return nil, err
		} else {
			name = string(buf[:namelen])
		}
		if vis > SYM_VPRIVATE {
			return nil, fmt.Errorf("Invalid symbol `%s` visibility", name)
		}
		table[name] = SymbolData{
			Ty:  ty,
			Vis: vis,
			Loc: loc,
		}
	}

	return table, nil

}
func writeSymbolDef(sname string, sym SymbolData) []byte {
	// 1b(TY) 1b(VISIBILITY) 8b(LOC) 4b(NAMELEN) NAMELENb(NAME)
	head := make([]byte, 1+1+8+4+len(sname))
	head[0] = sym.Ty
	head[1] = sym.Vis
	binary.BigEndian.PutUint64(head[2:], sym.Loc)
	binary.BigEndian.PutUint32(head[10:], uint32(len(sname)))
	head = append(head, []byte(sname)...)
	return head
}
func writeSymbols(st *SymbolTable) ([]byte, error) {
	syms := make([]byte, 0, 64)

	for sname, sym := range *st {
		syms = append(syms, writeSymbolDef(sname, sym)...)
	}

	return syms, nil
}

func LoadObjFile(h ObjFileHeader, binary []byte) (ObjFile, error) {
	ret := ObjFile{}
	ret.Header = h
	if int(h.CodeSize) == 0 {
		return ret, fmt.Errorf("Bad header code sec data")
	}
	if (int(h.CodeStart)+int(h.CodeSize)-int(h.CodeStart))%vm.INSTRUCTION_SIZE != 0 {
		return ret, fmt.Errorf("Bad code section size")
	}
	ret.Code = binary[uint64(OBJ_FILE_HEADER_SIZE)+h.CodeStart : uint64(OBJ_FILE_HEADER_SIZE)+h.CodeStart+h.CodeSize]

	if h.SymbolsSize != 0 {
		symSec := binary[uint64(OBJ_FILE_HEADER_SIZE)+h.SymbolsStart : uint64(OBJ_FILE_HEADER_SIZE)+h.SymbolsStart+h.SymbolsSize]
		if syms, err := loadSymbols(symSec); err != nil {
			return ret, err
		} else {
			ret.Symbols = syms
		}
	}
	if h.StaticDataSize != 0 {
		return ret, fmt.Errorf("sdata todo")
	}
	return ret, nil
}
func writeRelocDef(rel RelocData) []byte {
	head := make([]byte, 8+8+1)
	binary.BigEndian.PutUint64(head[0:], rel.Loc)
	binary.BigEndian.PutUint64(head[8:], rel.Ref)
	head[len(head)-1] = rel.PatchSize
	return head
}
func writeRelocs(rel RelocationTable) ([]byte, error) {
	relocs := make([]byte, 0, 64)

	for _, reloc := range rel {
		relocs = append(relocs, writeRelocDef(reloc)...)
	}
	return relocs, nil
}
