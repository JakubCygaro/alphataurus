package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/JakubCygaro/alphataurus/assembler"
	"github.com/JakubCygaro/alphataurus/internal/vm"
	"github.com/JakubCygaro/alphataurus/linker"
)

const assembly = `
section '.code'
	mov r0, 420
`
func main() {
	asm := assembler.NewAssembler(*bufio.NewReader(strings.NewReader(assembly)))
	bytecode, err := asm.Assemble()
	if err != nil {
		os.Stderr.WriteString("assembling error\n")
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(-1)
	}
	iCount := asm.InstructionCount()
	fmt.Printf("emitted bytecode size: %d\n", len(bytecode))
	fmt.Printf("emitted %d instructions\n", iCount)
	ld := linker.NewLinker()
	elf, err := ld.LinkBytes([]linker.Bytes{bytecode})
	if err != nil {
		os.Stderr.WriteString("linking error\n")
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(-1)
	}
	mach := vm.CreateVmState(16)
	err = mach.Execute(elf)
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(-1)
	}
	if rx, err := mach.GetGpRXAsUint64(vm.R0_IDX); err == nil {
		fmt.Printf("r0 = %064b\n", rx)
	}
	if rx, err := mach.GetGpRXAsUint64(vm.R1_IDX); err == nil {
		fmt.Printf("r1 = %+v\n", rx)
	}
	if rx, err := mach.GetGpRXAsUint64(vm.R2_IDX); err == nil {
		fmt.Printf("r2 = %+v\n", rx)
	}
	fmt.Printf("%+v\n", mach.GetRegisters())
	fmt.Printf("%+v\n", mach.GetFlags())
	fmt.Printf("stack:\n%+v\n", mach.GetStack())
	const in = "(-3 + 8)"
	p := assembler.NewParser(*bufio.NewReader(strings.NewReader(in)))
	expr, _ := p.ParseExpression()
	expr, _ = assembler.TryEvaluatePruneExpression(expr)
	fmt.Println(expr.Emit())
	
}
