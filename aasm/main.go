package main

import (
	"bufio"
	"os"
	"fmt"
	"github.com/JakubCygaro/alphataurus/assembler"
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		os.Exit(-1)
	}
	file, err := os.Open(args[0])
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(-1)
	}
	asm := assembler.NewAssembler(*bufio.NewReader(file))
	bytecode, iCount, err := asm.EmitBytecode()
	if err != nil {
		os.Stderr.WriteString(err.Error())
	} else {
		stride := int(len(bytecode) / iCount)
		for i := 0; i < stride / len(bytecode); i++ {
			fmt.Println(bytecode[i*stride:(i+1)*stride])
		}
	}
}
