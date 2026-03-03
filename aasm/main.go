package main

import (
	"bufio"
	"os"

	"github.com/JakubCygaro/alphataurus/assembler"
	"github.com/JakubCygaro/alphataurus/internal/vm"
	arg "github.com/alexflint/go-arg"
)

var args struct {
	Input  string `arg:"required,positional"`
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
	bytecode, iCount, err := asm.EmitBytecode()
	if err != nil {
		os.Stderr.WriteString(err.Error())
	} else {
		stride := int(len(bytecode) / iCount)
		var out *os.File
		if args.Output != "" {
			out, err = os.Open(args.Output)
		} else {
			out, err = os.Open(file.Name())
		}
		if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(-1)
		}
		defer out.Close()
		for i := 0; i < stride/len(bytecode); i++ {
			if _, err := out.Write(bytecode[i*vm.INSTRUCTION_SIZE : i*vm.INSTRUCTION_SIZE+stride]); err != nil {
				os.Stderr.WriteString(err.Error())
				os.Exit(-1)
			}
		}
	}
}
