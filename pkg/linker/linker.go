package linker

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	aelf "github.com/JakubCygaro/alphataurus/pkg/vm/aelf"
	"github.com/JakubCygaro/alphataurus/pkg/vm/decls"
	"github.com/JakubCygaro/alphataurus/pkg/vm/obj"
)

type Option[T any] struct {
	val T
	has bool
}

// func (opt *Option[T]) HasValue() bool {
// 	return opt.has
// }
// func (opt *Option[T]) Set(val T) {
// 	opt.val, opt.has = val, true
// }
// func (opt *Option[T]) Get() T {
// 	if !opt.has {
// 		panic("called Get on empty Option[T]")
// 	}
// 	return opt.val
// }

type LinkingOptions struct {
	EntryFile string
	EntryLab  string
}

func DefaultLinkingOpts() LinkingOptions {
	return LinkingOptions{}
}

type objFileData struct {
	IsInMemory bool
	Path       string
	Loaded     vm.ObjFile
	Raw        []byte
}
type objFileIdx int
type globalSymbolTable struct {
	symbols vm.SymbolTable
	files   map[*vm.SymbolData]objFileIdx
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
		symbols: vm.NewSymbolTable(),
		files:   make(map[*vm.SymbolData]objFileIdx),
	}
}

func (t *globalSymbolTable) AddSymbol(
	file objFileIdx, name string, sym *vm.SymbolData) error {
	if err := t.symbols.AddForeignSymbol(name, sym); err != nil {
		return err
	}
	t.files[sym] = file
	return nil
}
func (t *globalSymbolTable) GetSymbol(name string) (
	file objFileIdx, inTable int, sym *vm.SymbolData, ok bool) {
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
	opts        LinkingOptions
}

func NewLinker() Linker {
	return Linker{
		objectFiles: make([]objFileData, 0),
		globals:     newGlobalSymbolTable(),
		relocations: make(map[objFileIdx]fileReloc),
	}
}
func (l *Linker) readGlobalSymbols(objidx objFileIdx, obj *vm.ObjFile) error {
	for name, idx := range obj.Symbols.ByName {
		sym := obj.Symbols.InOrder[idx]
		if sym.Vis != vm.SYM_VEXPORT {
			continue
		}
		if err := l.globals.AddSymbol(objidx, name, sym); err != nil {
			return fmt.Errorf("%s -- symbol '%s'", err, name)
		}
	}
	return nil
}

func (l *Linker) collectSources(sources []LinkerInput) error {
	foundEntry := l.opts.EntryFile == ""
	for _, src := range sources {
		meta := src.GetMetadata()
		if !foundEntry && meta.IsFile && meta.FilePath == l.opts.EntryFile {
			foundEntry = true
		}
		source, err := src.ToBytes()
		if err != nil {
			return err
		}
		if err := l.collect(source, meta); err != nil {
			return err
		}
	}
	if !foundEntry {
		return fmt.
			Errorf("TODO: specified entry point file `%s` not found", l.opts.EntryFile)
	}
	return nil
}

func (l *Linker) collect(b Bytes, meta inputMetadata) error {
	header, err := vm.LoadObjFileHeader(bufio.NewReader(bytes.NewReader(b)))
	if err != nil {
		return err
	}
	obj, err := vm.LoadObjFile(header, b)
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
func (l *Linker) link() (aelf.AlphaELFFile, error) {
	ret := aelf.AlphaELFFile{}
	collectedCode := make([]byte, 0)
	codeBaseOff := uint64(0)
	overrideEntry := l.opts.EntryFile != ""
	for idx, obj := range l.objectFiles {
		if err := l.readGlobalSymbols(objFileIdx(idx),
			&(l.objectFiles[idx].Loaded)); err != nil {

			return ret, err
		}
		collectedCode = append(collectedCode, obj.Loaded.Code...)
		objCodeSize := uint64(len(obj.Loaded.Code))
		ret.CodeSize += objCodeSize
		l.relocations[objFileIdx(idx)] = fileReloc{
			CodeSecOff: codeBaseOff,
		}
		if overrideEntry && obj.Path == l.opts.EntryFile {
			if l.opts.EntryLab == "" && !obj.Loaded.Header.HasEntry {
				return ret, fmt.
					Errorf("TODO: File `%s` does not declare an entry point",
						l.opts.EntryFile)
			} else if obj.Loaded.Header.HasEntry {
				ret.HasEntry = true
				ret.Entry = obj.Loaded.Header.Entry + codeBaseOff
			} else {
				if s, _, ok :=
					obj.Loaded.Symbols.GetByName(l.opts.EntryLab); !ok {
					return ret, fmt.
						Errorf("TODO: File `%s` does not declare symbol `%s`",
							l.opts.EntryFile,
							l.opts.EntryLab)
				} else {
					ret.HasEntry = true
					ret.Entry = s.Loc + codeBaseOff + decls.INSTRUCTION_SIZE
				}
			}
		} else if obj.Loaded.Header.HasEntry && !ret.HasEntry {
			ret.HasEntry = true
			ret.Entry = obj.Loaded.Header.Entry + codeBaseOff
		} else if obj.Loaded.Header.HasEntry {
			return ret, fmt.Errorf("Multiple entry points defined")
		}
		codeBaseOff += objCodeSize
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
				f, _, symInGlobals, ok := l.globals.GetSymbol(symInFile.GetName())
				if !ok {
					return ret, fmt.Errorf("Unresolved symbol '%s'", symInFile.GetName())
				}
				relocated := l.relocations[objFileIdx(f)]
				realRef = +symInGlobals.Loc + relocated.CodeSecOff
			} else {
				realRef = +symInFile.Loc + thisObjRels.CodeSecOff
			}
			// now apply the patch
			switch rel.PatchSize {
			case 8:
				binary.BigEndian.PutUint64(collectedCode[realLoc:], realRef)
			case 4:
				binary.BigEndian.PutUint32(collectedCode[realLoc:], uint32(realRef))
			case 2:
				binary.BigEndian.PutUint16(collectedCode[realLoc:], uint16(realRef))
			case 1:
				collectedCode[realLoc] = byte(realRef)
			default:
				return ret, fmt.Errorf("Bad patch size of %d", rel.PatchSize)
			}
		}
	}
	ret.CodeStart = 0
	ret.CodeSize = uint64(len(collectedCode))
	ret.PostHeaderData = collectedCode
	return ret, nil
}

func (l *Linker) Link(
	sources []LinkerInput,
	opts LinkingOptions,
) (aelf.AlphaELFFile, error) {
	ret := aelf.AlphaELFFile{}
	l.opts = opts
	if err := l.collectSources(sources); err != nil {
		return ret, err
	}
	return l.link()
}
