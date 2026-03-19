package assembler

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/JakubCygaro/alphataurus/internal/vm"
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

type SymbolTable map[string]SymbolData

type SymbolData struct {
	Ty  int
	Val []uint64
}

type ObjFile struct {
	Header     ObjFileHeader
	Code       []byte
	StaticData StaticDataTable
	Symbols    SymbolData
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

func LoadObjFile(h ObjFileHeader, binary []byte) (ObjFile, error) {
	ret := ObjFile{}
	ret.Header = h
	if len(binary) < int(h.CodeStart)+int(h.CodeSize) {
		return ret, fmt.Errorf("Bad header code sec data")
	}
	if (int(h.CodeStart)+int(h.CodeSize)-int(h.CodeStart))%vm.INSTRUCTION_SIZE != 0 {
		return ret, fmt.Errorf("Bad code section size")
	}
	ret.Code = binary[uint64(OBJ_FILE_HEADER_SIZE)+h.CodeStart : uint64(OBJ_FILE_HEADER_SIZE)+h.CodeStart+h.CodeSize]

	if h.StaticDataSize != 0 {
		return ret, fmt.Errorf("sdata todo")
	}
	if h.SymbolsSize != 0 {
		return ret, fmt.Errorf("sym todo")
	}
	return ret, nil
}
