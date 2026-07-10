package vm

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/JakubCygaro/alphataurus/pkg/vm/decls"
)

const (
	OBJ_FILE_MAG = "AOBJ"
)

const (
	OBJ_FILE_SECCODE  = 1
	OBJ_FILE_SECSDATA = 1 + iota
	OBJ_FILE_SECSYMS
	OBJ_FILE_SECRELS
	OBJ_FILE_ENTRY
)

// The header starts with a magic number OBJ_FILE_MAG
// then comes the version 4b(MAJOR, MINOR, TWEAK, PATCH)
// followed by the size of the header 2b (the sum size of all following header data)
// then all other data comes in the format 1b(TAG) xb(DATA)
type ObjFileHeader struct {
	Version                         uint32
	HeaderSize                      uint16
	HasEntry                        bool
	Entry                           uint64
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
			if s, sz, err := readSecStartSize(reader); err != nil {
				return ret, err
			} else {
				ret.CodeStart, ret.CodeSize = s, sz
				i += 16
			}
		case OBJ_FILE_SECSDATA:
			if s, sz, err := readSecStartSize(reader); err != nil {
				return ret, err
			} else {
				ret.StaticDataStart, ret.StaticDataSize = s, sz
				i += 16
			}
		case OBJ_FILE_SECSYMS:
			if s, sz, err := readSecStartSize(reader); err != nil {
				return ret, err
			} else {
				ret.SymbolsStart, ret.SymbolsSize = s, sz
				i += 16
			}
		case OBJ_FILE_SECRELS:
			if s, sz, err := readSecStartSize(reader); err != nil {
				return ret, err
			} else {
				ret.RelocsStart, ret.RelocsSize = s, sz
				i += 16
			}
		case OBJ_FILE_ENTRY:
			buf := [8]byte{}
			if _, err := io.ReadFull(reader, buf[:]); err != nil {
				return ret, err
			} else {
				ret.HasEntry = true
				ret.Entry = binary.BigEndian.Uint64(buf[:])
				i += 8
			}
		}

	}
	return ret, nil
}
func LoadObjFile(h ObjFileHeader, binary []byte) (ObjFile, error) {
	ret := ObjFile{}
	h.HeaderSize += 10
	ret.Header = h
	if int(h.CodeSize) == 0 {
		return ret, fmt.Errorf("Bad header code sec data")
	}
	if (int(h.CodeStart)+int(h.CodeSize)-int(h.CodeStart))%decls.INSTRUCTION_SIZE != 0 {
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
