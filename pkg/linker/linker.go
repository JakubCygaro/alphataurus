package linker

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	asm "github.com/JakubCygaro/alphataurus/pkg/assembler"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

type Linker struct {
	objectFiles []asm.ObjFile
}

func NewLinker() Linker {
	return Linker{
		objectFiles: make([]asm.ObjFile, 0),
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
		l.objectFiles = append(l.objectFiles, obj)
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
		l.objectFiles = append(l.objectFiles, obj)
	}
	return nil
}
func (l *Linker) link() (vm.AlphaELFFile, error) {
	ret := vm.AlphaELFFile{}
	ret.Version = l.objectFiles[0].Header.Version
	ret.CodeStart = l.objectFiles[0].Header.CodeStart
	ret.CodeSize = l.objectFiles[0].Header.CodeSize
	ret.Data = l.objectFiles[0].Code
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
