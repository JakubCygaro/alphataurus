package vm

import (
	"bufio"
	"bytes"
	"encoding/binary"

	sliceutils "github.com/JakubCygaro/alphataurus/internal/pkg/slice_utils"
)
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
func LoadRelocsWithHeader(h ObjFileHeader, obj []byte) (RelocationTable, error) {
	if h.RelocsSize == 0 {
		return make(RelocationTable, 0), nil
	}
	_, data := sliceutils.Chop(obj, 10 + int(h.HeaderSize))
	rels := data[h.RelocsStart:]
	rels, _ = sliceutils.Chop(rels, int(h.RelocsSize))
	return readRelocs(rels)
}
