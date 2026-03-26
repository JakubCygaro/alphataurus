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
	Loaded asm.ObjFile
	Raw    []byte
}

type Linker struct {
	objectFiles []objFileData
}

func NewLinker() Linker {
	return Linker{
		objectFiles: make([]objFileData, 0),
	}
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
	ret.Version = l.objectFiles[0].Loaded.Header.Version
	ret.CodeStart = l.objectFiles[0].Loaded.Header.CodeStart
	ret.CodeSize = l.objectFiles[0].Loaded.Header.CodeSize
	ret.StaticDataStart = l.objectFiles[0].Loaded.Header.StaticDataStart
	ret.StaticDataSize = l.objectFiles[0].Loaded.Header.StaticDataSize
	ret.SymbolsStart = l.objectFiles[0].Loaded.Header.SymbolsStart
	ret.SymbolsSize = l.objectFiles[0].Loaded.Header.SymbolsSize
	ret.RelocsStart = l.objectFiles[0].Loaded.Header.RelocsStart
	ret.RelocsSize = l.objectFiles[0].Loaded.Header.RelocsSize
	ret.Data = l.objectFiles[0].Raw
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
