package vm

import (
	"encoding/binary"
)

const (
	AELF_FILE_MAG = "AELF"
)

type AlphaELFFile struct {
	Version                         uint32
	HeaderSize                      uint16
	HasEntry                        bool
	Entry                           uint64
	CodeStart, CodeSize             uint64
	StaticDataStart, StaticDataSize uint64
	SymbolsStart, SymbolsSize       uint64
	RelocsStart, RelocsSize         uint64
	Data                            []byte
}

func writeSeg(out []byte, segStart, segSize uint64, segTy byte) {
	out = append(out, segTy)
	out = binary.BigEndian.AppendUint64(out, segStart)
	out = binary.BigEndian.AppendUint64(out, segSize)
}

func (aelf AlphaELFFile) Write() []byte {
	out := make([]byte, 0)
	out = append(out, AELF_FILE_MAG...)
	out = binary.BigEndian.AppendUint32(out, aelf.Version)
	out = binary.BigEndian.AppendUint16(out, 0) // to be patched later
	prembleSize := len(out)

	if aelf.HasEntry {
		out = append(out, OBJ_FILE_ENTRY)
		out = binary.BigEndian.AppendUint64(out, aelf.Entry)
	}
	if aelf.CodeSize != 0 {
		writeSeg(out, aelf.CodeStart, aelf.CodeSize, OBJ_FILE_SECCODE)
	}
	if aelf.StaticDataSize != 0 {
		writeSeg(out, aelf.StaticDataStart, aelf.StaticDataSize, OBJ_FILE_SECSDATA)
	}
	if aelf.SymbolsSize != 0 {
		writeSeg(out, aelf.SymbolsStart, aelf.SymbolsSize, OBJ_FILE_SECSYMS)
	}
	if aelf.RelocsSize != 0 {
		writeSeg(out, aelf.RelocsStart, aelf.RelocsSize, OBJ_FILE_SECRELS)
	}

	headerSize := uint16(len(out) - prembleSize)
	binary.BigEndian.PutUint16(out[:8], headerSize)

	out = append(out, aelf.Data...)

	return out
}
