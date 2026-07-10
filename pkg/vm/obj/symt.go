package vm

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"

	sliceutils "github.com/JakubCygaro/alphataurus/internal/pkg/slice_utils"
)

//go:generate stringer -type=SymbolType
type SymbolType byte

const (
	SYM_TFUNC SymbolType = iota
	SYM_TSTATVAR
)

//go:generate stringer -type=SymbolVisibility
type SymbolVisibility byte

const (
	/// defined in file
	SYM_VEXPORT SymbolVisibility = iota
	/// required by the file but not defined
	SYM_VIMPORTSTRONG
	/// weak reference in the file
	SYM_VIMPORTWEAK
	/// defined by the file but not linkable from the outside
	SYM_VPRIVATE
)

type SymbolTable struct {
	InOrder    []*SymbolData
	ByName     map[string]int
	ByLocation map[uint64]*SymbolData
	foreign    map[*SymbolData]struct{}
}

type SymbolData struct {
	Ty  SymbolType
	Vis SymbolVisibility
	// the location is defined with the deadzone added, so any value below the deadzone is treated as invalid
	Loc  uint64
	name string
}

func (sd *SymbolData) String() string {
	return fmt.
		Sprintf(
			"%s %s '%s' at 0x%x",
			sd.Vis.String(),
			sd.Ty.String(),
			sd.name,
			sd.Loc,
		)
}

func (sym *SymbolData) GetName() string {
	return sym.name
}

func DefineSymbol(
	ty SymbolType, vis SymbolVisibility, loc uint64, name string) SymbolData {
	return SymbolData{
		Ty:   ty,
		Vis:  vis,
		Loc:  loc,
		name: name,
	}
}

func NewSymbolTable() SymbolTable {
	table := SymbolTable{
		InOrder:    make([]*SymbolData, 0),
		ByName:     make(map[string]int),
		ByLocation: make(map[uint64]*SymbolData),
		foreign:    make(map[*SymbolData]struct{}),
	}
	return table
}

func (t *SymbolTable) Clear() {
	clear(t.InOrder)
	clear(t.ByName)
	clear(t.ByLocation)
}
func (t *SymbolTable) AddSymbol(def SymbolData) (*SymbolData, error) {
	if i, ok := t.ByName[def.name]; ok {
		return nil,
			fmt.Errorf("TODO: Redefinition of symbol `%s`\n"+
				"Previously defined at: 0x%x", def.name, t.InOrder[i].Loc)
	}
	if sym, ok := t.ByLocation[def.Loc]; ok &&
		def.Vis != SYM_VIMPORTSTRONG &&
		def.Vis != SYM_VIMPORTWEAK {
		return nil,
			fmt.Errorf(
				"TODO: Multiple non-import symbols defined for single location"+
					" within symbol table.\n"+
					"First: %v\nSecond: %v", sym.name, def.name)
	}
	t.InOrder = append(t.InOrder, &def)
	t.ByName[def.name] = len(t.InOrder) - 1
	if def.Vis != SYM_VIMPORTSTRONG &&
		def.Vis != SYM_VIMPORTWEAK {
		t.ByLocation[def.Loc] = &def
	}
	return t.InOrder[len(t.InOrder)-1], nil
}
func (t *SymbolTable) AddForeignSymbol(name string, sym *SymbolData) error {
	if i, ok := t.ByName[name]; ok {
		return fmt.Errorf("TODO: Redefinition of symbol `%s`\n"+
			"Previously defined at: 0x%x", name, t.InOrder[i].Loc)
	}
	t.InOrder = append(t.InOrder, sym)
	t.ByName[name] = len(t.InOrder) - 1
	t.foreign[sym] = struct{}{}
	return nil
}
func (t *SymbolTable) RenameSymbol(oldName, newName string) error {
	if sym, idx, ok := t.GetByName(oldName); ok {
		if _, ok := t.foreign[sym]; ok {
			return fmt.Errorf("TODO: Foreign symbol `%s`cannot be renamed",
				oldName)
		}
		delete(t.ByName, oldName)
		sym.name = newName
		t.ByName[newName] = idx
		return nil
	}
	return fmt.Errorf("TODO: No such symbol `%s`, cannot rename to `%s`",
		oldName, newName)
}

func (t *SymbolTable) GetByName(name string) (*SymbolData, int, bool) {
	if idx, ok := t.ByName[name]; ok {
		return t.InOrder[idx], idx, ok
	}
	return nil, -1, false
}
func (t *SymbolTable) GetByLocation(loc uint64) (*SymbolData, bool) {
	if sym, ok := t.ByLocation[loc]; ok {
		return sym, ok
	}
	return nil, false
}
func (t *SymbolTable) GetName(idx int) (string, bool) {
	for n, i := range t.ByName {
		if i == idx {
			return n, true
		}
	}
	return "", false

}
func readSymbols(symbolSec []byte) (SymbolTable, error) {
	// 1b(TY) 1b(VISIBILITY) 8b(LOC) 4b(NAMELEN) NAMELENb(NAME)
	table := NewSymbolTable()
	reader := bufio.NewReader(bytes.NewReader(symbolSec))
	buf := make([]byte, 64)
	for {
		var vis SymbolVisibility
		var loc uint64
		var namelen uint32
		var name string
		first, err := reader.ReadByte()
		if err != nil {
			break
		}
		ty := SymbolType(first)
		if ty > SYM_TSTATVAR {
			return table, fmt.Errorf("Invalid symbol type")
		}
		if b, err := reader.ReadByte(); err != nil {
			return table, err
		} else {
			vis = SymbolVisibility(b)
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
			return table, fmt.Errorf("Invalid symbol `%s` visibility [%d]", name, vis)
		}
		sym := DefineSymbol(
			ty,
			vis,
			loc,
			name,
		)
		table.AddSymbol(sym)
	}
	return table, nil
}
func writeSymbolDef(sname string, sym SymbolData) []byte {
	// 1b(TY) 1b(VISIBILITY) 8b(LOC) 4b(NAMELEN) NAMELENb(NAME)
	head := make([]byte, 1+1+8+4)
	head[0] = byte(sym.Ty)
	head[1] = byte(sym.Vis)
	binary.BigEndian.PutUint64(head[2:], sym.Loc)
	binary.BigEndian.PutUint32(head[10:], uint32(len(sname)))
	head = append(head, []byte(sname)...)
	return head
}
func WriteSymbols(st *SymbolTable) ([]byte, error) {
	syms := make([]byte, 0, 64)
	for idx, sym := range (*st).InOrder {
		sname, _ := st.GetName(idx)
		syms = append(syms, writeSymbolDef(sname, *sym)...)
	}

	return syms, nil
}
func LoadSymbolsWithHeader(h ObjFileHeader, obj []byte) (SymbolTable, error) {
	if h.SymbolsSize == 0 {
		return NewSymbolTable(), nil
	}
	_, data := sliceutils.Chop(obj, 10+int(h.HeaderSize))
	syms := data[h.SymbolsStart:]
	syms, _ = sliceutils.Chop(syms, int(h.SymbolsSize))
	return readSymbols(syms)
}
