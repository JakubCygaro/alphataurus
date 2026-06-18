package vm

import (
	"encoding/binary"
	"fmt"

	sutil "github.com/JakubCygaro/alphataurus/internal/pkg/slice_utils"
	obj "github.com/JakubCygaro/alphataurus/pkg/vm/obj"
)

const (
	AELF_FILE_MAG = "AELF"
)
const (
	AELF_FILE_SECCODE  = obj.OBJ_FILE_SECCODE
	AELF_FILE_SECSDATA = obj.OBJ_FILE_SECSDATA
	AELF_FILE_SECSYMS  = obj.OBJ_FILE_SECSYMS
	AELF_FILE_SECRELS  = obj.OBJ_FILE_SECRELS
	AELF_FILE_ENTRY    = obj.OBJ_FILE_ENTRY
)

// the AELF file is laid out like so:
//
// === PREAMBLE ===
//
// mag: 4b
//
// ver: 4b
//
// header size: 2b
//
// === HEADER ===
//
// field tag: 1b
//
// field: var b
type AlphaELFFile struct {
	Version uint32
	// this excludes the preamble, because it is of constant size
	//
	// the header starts right after the header size
	HeaderSize                      uint16
	HasEntry                        bool
	Entry                           uint64
	CodeStart, CodeSize             uint64
	StaticDataStart, StaticDataSize uint64
	SymbolsStart, SymbolsSize       uint64
	RelocsStart, RelocsSize         uint64
	// All bytes that come after the header section of the file
	PostHeaderData                  []byte
}

func writeSeg(out []byte, segStart, segSize uint64, segTy byte) (res []byte) {
	out = append(out, segTy)
	out = binary.BigEndian.AppendUint64(out, segStart)
	out = binary.BigEndian.AppendUint64(out, segSize)
	return out
}

func (aelf AlphaELFFile) Write() []byte {
	out := make([]byte, 0, 64)
	out = append(out, []byte(AELF_FILE_MAG)...)
	out = binary.BigEndian.AppendUint32(out, aelf.Version)
	out = binary.BigEndian.AppendUint16(out, 0) // header size to be patched later
	prembleSize := len(out)

	if aelf.HasEntry {
		out = append(out, obj.OBJ_FILE_ENTRY)
		out = binary.BigEndian.AppendUint64(out, aelf.Entry)
	}
	if aelf.CodeSize != 0 {
		out = writeSeg(out, aelf.CodeStart, aelf.CodeSize, AELF_FILE_SECCODE)
	}
	if aelf.StaticDataSize != 0 {
		out = writeSeg(out, aelf.StaticDataStart, aelf.StaticDataSize, AELF_FILE_SECSDATA)
	}
	if aelf.SymbolsSize != 0 {
		out = writeSeg(out, aelf.SymbolsStart, aelf.SymbolsSize, AELF_FILE_SECSYMS)
	}
	if aelf.RelocsSize != 0 {
		out = writeSeg(out, aelf.RelocsStart, aelf.RelocsSize, AELF_FILE_SECRELS)
	}

	headerSize := uint16(len(out) - prembleSize)
	binary.BigEndian.PutUint16(out[8:], headerSize)

	out = append(out, aelf.PostHeaderData...)

	return out
}
func LoadAELF(bytes []byte) (AlphaELFFile, error) {
	full := bytes
	ret := AlphaELFFile{}
	mag, bytes := sutil.Chop(bytes, 4)
	if mag == nil {
		return ret, notAnAELFFILEError("file too short")
	}
	for i := range 4 {
		if mag[i] != AELF_FILE_MAG[i] {
			return ret, notAnAELFFILEError("bad file identifier")
		}
	}
	ver, bytes := sutil.Chop(bytes, 4)
	if ver == nil {
		return ret, notAnAELFFILEError("bad version section")
	}
	ret.Version = binary.BigEndian.Uint32(ver[:])
	headSz, bytes := sutil.Chop(bytes, 2)
	if headSz == nil {
		return ret, notAnAELFFILEError("header size not given")
	}
	ret.HeaderSize = binary.BigEndian.Uint16(headSz[:])
	if len(full) < int(ret.HeaderSize) {
		return ret, notAnAELFFILEError("bad header size")
	}
	headerBytes := full[10 : 10+ret.HeaderSize]
	if err := readHeader(headerBytes, &ret); err != nil {
		return ret, err
	}
	ret.PostHeaderData = full[10+ret.HeaderSize:]
	return ret, nil
}
func readHeader(header []byte, out *AlphaELFFile) error {
	for len(header) > 0 {
		var tag []byte
		tag, header = sutil.Chop(header, 1)
		switch tag[0] {
		case AELF_FILE_ENTRY:
			if out.HasEntry {
				return malformedAELFError("multiple entry point declarations")
			}
			var entry []byte
			entry, header = sutil.Chop(header, 8)
			if entry == nil {
				return malformedAELFError("bad entry point data")
			}
			e := binary.BigEndian.Uint64(entry[:])
			out.Entry = e
			out.HasEntry = true
		case AELF_FILE_SECCODE:
			var buf []byte
			buf, header = sutil.Chop(header, 8)
			if buf == nil {
				return malformedAELFError("bad code section start data")
			}
			out.CodeStart = binary.BigEndian.Uint64(buf[:])
			buf, header = sutil.Chop(header, 8)
			if buf == nil {
				return malformedAELFError("bad code section size data")
			}
			out.CodeSize = binary.BigEndian.Uint64(buf[:])
		default:
			return malformedAELFError("unknown AELF header tag [0x%x]", tag[0])
		}
	}
	return nil
}

func malformedAELFError(format string, a ...any) error {
	return fmt.
		Errorf("Malformed AELF header: %s", fmt.Sprintf(format, a...))
}
func notAnAELFFILEError(format string, a ...any) error {
	return fmt.
		Errorf("Not an AELF file: %s", fmt.Sprintf(format, a...))
}
