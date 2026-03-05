package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"github.com/JakubCygaro/alphataurus/assembler"
	"github.com/JakubCygaro/alphataurus/internal/vm"
)

const assembly = `
	push 1337
	push 80085
	mov bp, sp
	push 69
	push 420
	push 2137
	pop
	pop
	pop
	mov r0, [bp+1]
	mov r1, [bp+2]
	mov r2, [bp+3]
`

func main() {
	asm := assembler.NewAssembler(*bufio.NewReader(strings.NewReader(assembly)))
	bytecode, iCount, err := asm.EmitBytecode()
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(-1)
	}
	fmt.Printf("emitted bytecode size: %d\n", len(bytecode))
	fmt.Printf("emitted %d instructions\n", iCount)
	mach := vm.CreateVmState(64)
	err = mach.Execute(bytecode)
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
	}
	if rx, err := mach.GetGpRXAsUint64(vm.R0_IDX); err == nil{
		fmt.Printf("r0 = %+v\n", rx)
	}
	if rx, err := mach.GetGpRXAsUint64(vm.R1_IDX); err == nil{
		fmt.Printf("r1 = %+v\n", rx)
	}
	if rx, err := mach.GetGpRXAsUint64(vm.R2_IDX); err == nil{
		fmt.Printf("r2 = %+v\n", rx)
	}
	fmt.Printf("%+v\n", mach.GetRegisters())
	fmt.Printf("%+v\n", mach.GetFlags())
}
