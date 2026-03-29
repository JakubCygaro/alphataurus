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
	Name       string
	Loaded     asm.ObjFile
	Raw        []byte
}
type objFileIdx int
type globalSymbolTable struct {
	symbols asm.SymbolTable
	files   map[*asm.SymbolData]objFileIdx
}
func  newGlobalSymbolTable() globalSymbolTable {
	return globalSymbolTable{
		symbols: asm.NewSymbolTable(),
		files: make(map[*asm.SymbolData]objFileIdx),
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
	globals globalSymbolTable
}

func NewLinker() Linker {
	return Linker{
		objectFiles: make([]objFileData, 0),
		globals: newGlobalSymbolTable(),
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

func (l *Linker) collectFiles(sources []string) error {
	for _, src := range sources {
		file, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		headerBytes := file[:vm.AELF_FILE_HEADER_SIZE]
		header, err := asm.LoadObjFileHeader(bufio.NewReader(bytes.NewReader(headerBytes[:])))
		if err != nil {
			return err
		}
		obj, err := asm.LoadObjFile(header, file)
		if err != nil {
			return err
		}
		l.objectFiles = append(l.objectFiles, objFileData{Loaded: obj, Raw: file})
	}
	return nil
}
func (l *Linker) collectBytes(sources []Bytes) error {
	for _, file := range sources {
		if len(file) < vm.AELF_FILE_HEADER_SIZE {
			return fmt.Errorf("not a valid aobj file, header was too small")
		}
		headerBytes := file[:vm.AELF_FILE_HEADER_SIZE]
		header, err := asm.LoadObjFileHeader(bufio.NewReader(bytes.NewReader(headerBytes[:])))
		if err != nil {
			return err
		}
		obj, err := asm.LoadObjFile(header, file)
		if err != nil {
			return err
		}
		l.objectFiles = append(l.objectFiles, objFileData{Loaded: obj, Raw: file})
	}
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

type Bytes []byte

func (l *Linker) LinkBytes(sources []Bytes) (vm.AlphaELFFile, error) {
	ret := vm.AlphaELFFile{}
	if len(sources) > 1 {
		return ret, fmt.Errorf("Multiple object file linking TODO")
	}
	if err := l.collectBytes(sources); err != nil {
		return ret, err
	}
	return l.link()
}

func (l *Linker) LinkFiles(files []string) (vm.AlphaELFFile, error) {
	ret := vm.AlphaELFFile{}
	if len(files) > 1 {
		return ret, fmt.Errorf("Multiple object file linking TODO")
	}
	if err := l.collectFiles(files); err != nil {
		return ret, err
	}
	return l.link()
}
