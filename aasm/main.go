package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/JakubCygaro/alphataurus/assembler"
	arg "github.com/alexflint/go-arg"
)

var args struct {
	Input  string `arg:"positional,required"`
	Output string `arg:"-o,--output" help:"output file path"`
}

func main() {
	arg.MustParse(&args)
	file, err := os.Open(args.Input)
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(-1)
	}
	defer file.Close()
	asm := assembler.NewAssembler(*bufio.NewReader(file))
	bytecode, _, err := asm.EmitBytecode()
	if err != nil {
		os.Stderr.WriteString(err.Error())
	} else {
		var out *os.File
		if args.Output != "" {
			out, err = os.Create(args.Output)
		} else {
			name := file.Name()
			out, err = os.Create(fmt.Sprintf("%s.ao", strings.TrimSuffix(name, filepath.Ext(name))))
		}
		if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(-1)
		}
		defer out.Close()
		if _, err := out.Write(bytecode); err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(-1)
		}
	}
}
