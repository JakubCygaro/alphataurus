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
	OBJ_FILE_MAG = "AOBJ"
)

const (
	OBJ_FILE_SECCODE = 1
	OBJ_FILE_SECSDATA = 1 + iota
	OBJ_FILE_SECSYMS
	OBJ_FILE_SECRELS
)
// The header starts with a magic number OBJ_FILE_MAG
// then comes the version 4b(MAJOR, MINOR, TWEAK, PATCH)
// followed by the size of the header 2b (the sum size of all following header data)
// then all other data comes in the format 1b(TAG) xb(DATA)
type ObjFileHeader struct {
	Version                         uint32
	HeaderSize                      uint16
	CodeStart, CodeSize             uint64
	StaticDataStart, StaticDataSize uint64
	SymbolsStart, SymbolsSize       uint64
	RelocsStart, RelocsSize         uint64
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

type SymbolTable struct {
	InOrder []*SymbolData
	ByName  map[string]int
}

type SymbolData struct {
	Ty  byte
	Vis byte
	// the location is defined with the deadzone added, so any value below the deadzone is treated as invalid
	Loc uint64
}

func NewSymbolTable() SymbolTable {
	table := SymbolTable{
		InOrder: make([]*SymbolData, 0),
		ByName:  make(map[string]int),
	}
	return table
}

func (t *SymbolTable) AddSymbol(name string, def SymbolData) (*SymbolData, bool) {
	if _, ok := t.ByName[name]; ok {
		return nil, false
	}
	t.InOrder = append(t.InOrder, &def)
	t.ByName[name] = len(t.InOrder) - 1
	return t.InOrder[len(t.InOrder)-1], true
}

func (t *SymbolTable) GetByName(name string) (*SymbolData, int, bool) {
	if idx, ok := t.ByName[name]; ok {
		return t.InOrder[idx], idx, ok
	}
	return nil, -1, false
}

const (
	RELOC_TINVALID = 0
)

// 8b(LOC) 8b(REF) 1b(PATCHSIZE)
type RelocData struct {
	// where that symbol is referenced in the code
	Loc uint64
	// what symbol is being referenced
	Ref       uint64
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

func readSecStartSize(r *bufio.Reader) (start, size uint64, err error) {
	buf := [16]byte{}
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, 0, fmt.Errorf("Object file too short")
	}
	return binary.BigEndian.Uint64(buf[:8]), binary.BigEndian.Uint64(buf[8:]), nil
}

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

	if _, err := io.ReadFull(reader, buf[:2]); err != nil {
		return ret, fmt.Errorf("Object file too short")
	}
	ret.HeaderSize = binary.BigEndian.Uint16(buf)

	i := 0
	for ; i < int(ret.HeaderSize); i++ {
		tag, err := reader.ReadByte()
		if err != nil {
			return ret, err
		}
		switch tag {
		case OBJ_FILE_SECCODE:
			if s, sz, err := readSecStartSize(reader); err != nil{
				return ret, err
			} else {
				ret.CodeStart, ret.CodeSize = s, sz
				i += 16
			}
		case OBJ_FILE_SECSDATA:
			if s, sz, err := readSecStartSize(reader); err != nil{
				return ret, err
			} else {
				ret.StaticDataStart, ret.StaticDataSize = s, sz
				i += 16
			}
		case OBJ_FILE_SECSYMS:
			if s, sz, err := readSecStartSize(reader); err != nil{
				return ret, err
			} else {
				ret.SymbolsStart, ret.SymbolsSize = s, sz
				i += 16
			}
		case OBJ_FILE_SECRELS:
			if s, sz, err := readSecStartSize(reader); err != nil{
				return ret, err
			} else {
				ret.RelocsStart, ret.RelocsSize = s, sz
				i += 16
			}
		}

	}
	//
	// if _, err := io.ReadFull(reader, buf); err != nil {
	// 	return ret, fmt.Errorf("Object file too short")
	// }
	// ret.CodeStart = binary.BigEndian.Uint64(buf)
	//
	// if _, err := io.ReadFull(reader, buf); err != nil {
	// 	return ret, fmt.Errorf("Object file too short")
	// }
	// ret.CodeSize = binary.BigEndian.Uint64(buf)
	//
	// if _, err := io.ReadFull(reader, buf); err != nil {
	// 	return ret, fmt.Errorf("Object file too short")
	// }
	// ret.StaticDataStart = binary.BigEndian.Uint64(buf)
	//
	// if _, err := io.ReadFull(reader, buf); err != nil {
	// 	return ret, fmt.Errorf("Object file too short")
	// }
	// ret.StaticDataSize = binary.BigEndian.Uint64(buf)
	//
	// if _, err := io.ReadFull(reader, buf); err != nil {
	// 	return ret, fmt.Errorf("Object file too short")
	// }
	// ret.SymbolsStart = binary.BigEndian.Uint64(buf)
	//
	// if _, err := io.ReadFull(reader, buf); err != nil {
	// 	return ret, fmt.Errorf("Object file too short")
	// }
	// ret.SymbolsSize = binary.BigEndian.Uint64(buf)
	//
	// if _, err := io.ReadFull(reader, buf); err != nil {
	// 	return ret, fmt.Errorf("Object file too short")
	// }
	// ret.RelocsStart = binary.BigEndian.Uint64(buf)
	//
	// if _, err := io.ReadFull(reader, buf); err != nil {
	// 	return ret, fmt.Errorf("Object file too short")
	// }
	// ret.RelocsSize = binary.BigEndian.Uint64(buf)

	return ret, nil
}
func readSymbols(symbolSec []byte) (SymbolTable, error) {
	// 1b(TY) 1b(VISIBILITY) 8b(LOC) 4b(NAMELEN) NAMELENb(NAME)
	table := NewSymbolTable()
	reader := bufio.NewReader(bytes.NewReader(symbolSec))
	buf := make([]byte, 64)
	for {
		var vis byte
		var loc uint64
		var namelen uint32
		var name string
		first, err := reader.ReadByte()
		if err != nil {
			break
		}
		ty := first
		if ty > SYM_TSTATVAR {
			return table, fmt.Errorf("Invalid symbol type")
		}
		if b, err := reader.ReadByte(); err != nil {
			return table, err
		} else {
			vis = b
		}
		if _, err := reader.Read(buf[:8]); err != nil {
			return table, err
		} else {
			loc = binary.BigEndian.Uint64(buf[:8])
		}
		if _, err := reader.Read(buf[:4]); err != nil {
			return table, err
		} else {
			namelen = binary.BigEndian.Uint32(buf[:4])
		}
		if extendBy := int(namelen) - len(buf); extendBy > 0 {
			buf = make([]byte, len(buf)+extendBy)
		}
		if _, err := reader.Read(buf[:namelen]); err != nil {
			return table, err
		} else {
			name = string(buf[:namelen])
		}
		if vis > SYM_VPRIVATE {
			return table, fmt.Errorf("Invalid symbol `%s` visibility", name)
		}
		sym := SymbolData{
			Ty:  ty,
			Vis: vis,
			Loc: loc,
		}
		table.AddSymbol(name, sym)
	}
	return table, nil

}

// 8b(LOC) 8b(REF) 1b(PATCHSIZE)
func readRelocs(relocSec []byte) (RelocationTable, error) {
	relocs := make(RelocationTable, 0)
	reader := bufio.NewReader(bytes.NewReader(relocSec))
	buf := make([]byte, 0, 64)
	for {
		var loc, ref uint64
		var patchSize byte
		_, err := reader.Read(buf[:8])
		if err != nil {
			break
		}
		loc = binary.BigEndian.Uint64(buf[:8])
		if _, err := reader.Read(buf[:8]); err != nil {
			return relocs, err
		} else {
			ref = binary.BigEndian.Uint64(buf[:8])
		}
		if b, err := reader.ReadByte(); err != nil {
			return relocs, err
		} else {
			patchSize = b
		}
		reloc := RelocData{
			Loc:       loc,
			Ref:       ref,
			PatchSize: patchSize,
		}
		relocs = append(relocs, reloc)
	}
	return relocs, nil
}
func writeSymbolDef(sname string, sym SymbolData) []byte {
	// 1b(TY) 1b(VISIBILITY) 8b(LOC) 4b(NAMELEN) NAMELENb(NAME)
	head := make([]byte, 1+1+8+4)
	head[0] = sym.Ty
	head[1] = sym.Vis
	binary.BigEndian.PutUint64(head[2:], sym.Loc)
	binary.BigEndian.PutUint32(head[10:], uint32(len(sname)))
	head = append(head, []byte(sname)...)
	return head
}
func WriteSymbols(st *SymbolTable) ([]byte, error) {
	syms := make([]byte, 0, 64)

	for sname, idx := range (*st).ByName {
		sym := (*st).InOrder[idx]
		syms = append(syms, writeSymbolDef(sname, *sym)...)
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
	ret.Code = binary[uint64(h.HeaderSize)+h.CodeStart : uint64(h.HeaderSize)+h.CodeStart+h.CodeSize]

	if h.SymbolsSize != 0 {
		symSec := binary[uint64(h.HeaderSize)+h.SymbolsStart : uint64(h.HeaderSize)+h.SymbolsStart+h.SymbolsSize]
		if syms, err := readSymbols(symSec); err != nil {
			return ret, err
		} else {
			ret.Symbols = syms
		}
	}
	if h.RelocsSize != 0 {
		relocSec := binary[uint64(h.HeaderSize)+h.RelocsStart : uint64(h.HeaderSize)+h.RelocsStart+h.RelocsSize]
		if relocs, err := readRelocs(relocSec); err != nil {
			return ret, err
		} else {
			ret.Relocs = relocs
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
func WriteRelocs(rel RelocationTable) ([]byte, error) {
	relocs := make([]byte, 0, 64)

	for _, reloc := range rel {
		relocs = append(relocs, writeRelocDef(reloc)...)
	}
	return relocs, nil
}
