package linker

import (
	"bufio"
	"bytes"
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
		IsFile: true,
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
}

func NewLinker() Linker {
	return Linker{
		objectFiles: make([]objFileData, 0),
		globals:     newGlobalSymbolTable(),
	}
}
func (l *Linker) readGlobalSymbols(objidx objFileIdx, obj *asm.ObjFile) error {
	for name, idx := range obj.Symbols.ByName {
		sym := obj.Symbols.InOrder[idx]
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
		l.collect(source, meta)
	}
	return nil
}

func (l *Linker) collect(b Bytes, meta inputMetadata) error {
	headerBytes := b[:vm.AELF_FILE_HEADER_SIZE]
	header, err := asm.LoadObjFileHeader(bufio.NewReader(bytes.NewReader(headerBytes[:])))
	if err != nil {
		return err
	}
	obj, err := asm.LoadObjFile(header, b)
	if err != nil {
		return err
	}
	l.objectFiles = append(l.objectFiles, objFileData{
		Loaded: obj,
		Raw: b,
		IsInMemory: !meta.IsFile,
		Path: meta.FilePath,
	})
	return nil
}
func (l *Linker) link() (vm.AlphaELFFile, error) {
	ret := vm.AlphaELFFile{}
	for idx := range l.objectFiles {
		if err := l.readGlobalSymbols(objFileIdx(idx), &(l.objectFiles[idx].Loaded)); err != nil {
			return ret, err
		}
	}
	// for idx, obj := range l.objectFiles {
	// }
	ret.Version = l.objectFiles[0].Loaded.Header.Version
	ret.CodeStart = l.objectFiles[0].Loaded.Header.CodeStart
	ret.CodeSize = l.objectFiles[0].Loaded.Header.CodeSize
	ret.StaticDataStart = l.objectFiles[0].Loaded.Header.StaticDataStart
	ret.StaticDataSize = l.objectFiles[0].Loaded.Header.StaticDataSize
	ret.SymbolsStart = l.objectFiles[0].Loaded.Header.SymbolsStart
	ret.SymbolsSize = l.objectFiles[0].Loaded.Header.SymbolsSize
	ret.RelocsStart = l.objectFiles[0].Loaded.Header.RelocsStart
	ret.RelocsSize = l.objectFiles[0].Loaded.Header.RelocsSize
	ret.HeaderSize = l.objectFiles[0].Loaded.Header.HeaderSize
	ret.Data = l.objectFiles[0].Raw
	ret.Entry = l.objectFiles[0].Loaded.Header.Entry
	ret.HasEntry = l.objectFiles[0].Loaded.Header.HasEntry
	return ret, nil
}

func (l *Linker) Link(sources []LinkerInput) (vm.AlphaELFFile, error) {
	ret := vm.AlphaELFFile{}
	if len(sources) > 1 {
		return ret, fmt.Errorf("Multiple object file linking TODO")
	}
	if err := l.collectSources(sources); err != nil {
		return ret, err
	}
	return l.link()
}
