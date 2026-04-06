package linker

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	asm "github.com/JakubCygaro/alphataurus/pkg/assembler"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

type objFileData struct {
	IsInMemory bool
	Path       string
	Loaded     asm.ObjFile
	Raw        []byte
}
type objFileIdx int
type globalSymbolTable struct {
	symbols asm.SymbolTable
	files   map[*asm.SymbolData]objFileIdx
}
type SourcePath string
type Bytes []byte
type inputMetadata struct {
	IsFile   bool
	FilePath string
}
type fileReloc struct {
	CodeSecOff, StaticDataSecOff uint64
}
type LinkerInput interface {
	ToBytes() (Bytes, error)
	GetMetadata() inputMetadata
}

func (s SourcePath) ToBytes() (Bytes, error) {
	file, err := os.ReadFile(string(s))
	if err != nil {
		return nil, err
	}
	return file, nil
}
func (s SourcePath) GetMetadata() inputMetadata {
	return inputMetadata{
		IsFile:   true,
		FilePath: string(s),
	}
}
func (b Bytes) ToBytes() (Bytes, error) {
	return b, nil
}
func (b Bytes) GetMetadata() inputMetadata {
	return inputMetadata{
		IsFile: false,
	}
}

func newGlobalSymbolTable() globalSymbolTable {
	return globalSymbolTable{
		symbols: asm.NewSymbolTable(),
		files:   make(map[*asm.SymbolData]objFileIdx),
	}
}

func (t *globalSymbolTable) AddSymbol(file objFileIdx, name string, sym *asm.SymbolData) bool {
	if ok := t.symbols.AddForeignSymbol(name, sym); !ok {
		return false
	}
	t.files[sym] = file
	return true
}
func (t *globalSymbolTable) GetSymbol(name string) (file objFileIdx, inTable int, sym *asm.SymbolData, ok bool) {
	if sym, idx, ok := t.symbols.GetByName(name); !ok {
		return 0, 0, nil, false
	} else {
		return t.files[sym], idx, sym, ok
	}
}

type Linker struct {
	objectFiles []objFileData
	globals     globalSymbolTable
	relocations map[objFileIdx]fileReloc
}

func NewLinker() Linker {
	return Linker{
		objectFiles: make([]objFileData, 0),
		globals:     newGlobalSymbolTable(),
		relocations: make(map[objFileIdx]fileReloc),
	}
}
func (l *Linker) readGlobalSymbols(objidx objFileIdx, obj *asm.ObjFile) error {
	for name, idx := range obj.Symbols.ByName {
		sym := obj.Symbols.InOrder[idx]
		if sym.Vis != asm.SYM_VEXPORT {
			continue
		}
		if ok := l.globals.AddSymbol(objidx, name, sym); !ok {
			return fmt.Errorf("multiple definitions of symbol '%s'", name)
		}
	}
	return nil
}

func (l *Linker) collectSources(sources []LinkerInput) error {
	for _, src := range sources {
		meta := src.GetMetadata()
		source, err := src.ToBytes()
		if err != nil {
			return err
		}
		if err := l.collect(source, meta); err != nil {
			return err
		}
	}
	return nil
}

func (l *Linker) collect(b Bytes, meta inputMetadata) error {
	header, err := asm.LoadObjFileHeader(bufio.NewReader(bytes.NewReader(b)))
	if err != nil {
		return err
	}
	obj, err := asm.LoadObjFile(header, b)
	if err != nil {
		return err
	}
	l.objectFiles = append(l.objectFiles, objFileData{
		Loaded:     obj,
		Raw:        b,
		IsInMemory: !meta.IsFile,
		Path:       meta.FilePath,
	})
	return nil
}
func (l *Linker) link() (vm.AlphaEXEFile, error) {
	ret := vm.AlphaEXEFile{}
	data := make([]byte, 0)
	baseOff := uint64(0)
	for idx, obj := range l.objectFiles {
		if err := l.readGlobalSymbols(objFileIdx(idx), &(l.objectFiles[idx].Loaded)); err != nil {
			return ret, err
		}
		baseOff += uint64(len(data))
		data = append(data, obj.Loaded.Code...)
		ret.CodeSize += uint64(len(obj.Loaded.Code))
		l.relocations[objFileIdx(idx)] = fileReloc{
			CodeSecOff: baseOff,
		}
		if obj.Loaded.Header.HasEntry && !ret.HasEntry {
			ret.HasEntry = true
			ret.Entry = obj.Loaded.Header.Entry + baseOff
		} else if obj.Loaded.Header.HasEntry {
			return ret, fmt.Errorf("Multiple entry points defined")
		}
	}

	ret.Version = 0x00000001

	for idx, obj := range l.objectFiles {
		thisObjRels := l.relocations[objFileIdx(idx)]
		syms := obj.Loaded.Symbols
		for _, rel := range obj.Loaded.Relocs {
			realLoc := rel.Loc + thisObjRels.CodeSecOff
			symInFile := syms.InOrder[rel.Ref]
			realRef := uint64(0)
			// if this is an import symbol
			if symInFile.Loc == 0 {
				// find the symbol
				f, _, s, ok := l.globals.GetSymbol(symInFile.Name)
				if !ok {
					return ret, fmt.Errorf("Unresolved symbol '%s'", symInFile.Name)
				}
				relocated := l.relocations[objFileIdx(f)]
				realRef = + s.Loc + (relocated.CodeSecOff / vm.INSTRUCTION_SIZE)
			} else {
				realRef = + symInFile.Loc + (thisObjRels.CodeSecOff / vm.INSTRUCTION_SIZE)
			}
			// now apply the patch
			switch rel.PatchSize {
			case 8:
				binary.BigEndian.PutUint64(data[realLoc:], realRef)
			case 4:
				binary.BigEndian.PutUint32(data[realLoc:], uint32(realRef))
			case 2:
				binary.BigEndian.PutUint16(data[realLoc:], uint16(realRef))
			case 1:
				data[realLoc] = byte(realRef)
			default:
				return ret, fmt.Errorf("Bad patch size of %d", rel.PatchSize)
			}
		}
	}

	ret.CodeStart = 0
	ret.CodeSize = uint64(len(data))
	ret.Data = data
	// ret.CodeSize = l.objectFiles[0].Loaded.Header.CodeSize
	// ret.StaticDataStart = l.objectFiles[0].Loaded.Header.StaticDataStart
	// ret.StaticDataSize = l.objectFiles[0].Loaded.Header.StaticDataSize
	// ret.SymbolsStart = l.objectFiles[0].Loaded.Header.SymbolsStart
	// ret.SymbolsSize = l.objectFiles[0].Loaded.Header.SymbolsSize
	// ret.RelocsStart = l.objectFiles[0].Loaded.Header.RelocsStart
	// ret.RelocsSize = l.objectFiles[0].Loaded.Header.RelocsSize
	// ret.HeaderSize = l.objectFiles[0].Loaded.Header.HeaderSize
	// ret.Data = l.objectFiles[0].Raw
	// ret.Entry = l.objectFiles[0].Loaded.Header.Entry
	// ret.HasEntry = l.objectFiles[0].Loaded.Header.HasEntry
	return ret, nil
}

func (l *Linker) Link(sources []LinkerInput) (vm.AlphaEXEFile, error) {
	ret := vm.AlphaEXEFile{}
	if err := l.collectSources(sources); err != nil {
		return ret, err
	}
	return l.link()
}
