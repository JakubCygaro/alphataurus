package main

import (
	"bufio"
	"fmt"
	// "math/rand"
	"os"
	"strings"

	"github.com/JakubCygaro/alphataurus/assembler"
	"github.com/JakubCygaro/alphataurus/internal/vm"
)

const assembly = `
	mov r4, 118856529506480456
	sub SIGNED r4, -8413980612100018402
`

func main() {
	asm := assembler.NewAssembler(*bufio.NewReader(strings.NewReader(assembly)))
	bytecode, iCount, err := asm.EmitBytecode()
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(-1)
	} else {
		// stride := int(len(bytecode) / iCount)
		// fmt.Printf("stride %d \n", stride)
		// for i := 0; i < len(bytecode)/stride; i++ {
		// 	fmt.Println(bytecode[i*stride : (i+1)*stride])
		// }
	}
	fmt.Printf("emitted bytecode size: %d\n", len(bytecode))
	fmt.Printf("emitted %d instructions\n", iCount)
	mach := vm.CreateVmState(1024)
	err = mach.Execute(bytecode)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", mach.GetRegisters())
	fmt.Printf("%+v\n", mach.GetFlags())
}
