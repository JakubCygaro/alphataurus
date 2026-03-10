package main

import (
	"bufio"
	"fmt"
	"github.com/JakubCygaro/alphataurus/assembler"
	"github.com/JakubCygaro/alphataurus/internal/vm"
	"os"
	"strings"
)

const assembly = `
	push bp
	mov bp, sp
	mov [bp+1], 420
	mov [bp+2], 69
	mov r1, [bp+1]
	mov r2, [bp+2]
	add r1, r2
	mov [bp+3], r1
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
	mach := vm.CreateVmState(16)
	err = mach.Execute(bytecode)
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(-1)
	}
	if rx, err := mach.GetGpRXAsUint64(vm.R0_IDX); err == nil {
		fmt.Printf("r0 = %+v\n", rx)
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
	// const in = "[1+(bp+r3)+0+1]"
	// p := assembler.NewParser(*bufio.NewReader(strings.NewReader(in)))
	// expr, _ := p.ParseExpression()
	// expr, _ = assembler.TryEvaluatePruneExpression(expr)
	// fmt.Println(expr.Emit())
	// assembler.PruneExpression(expr, nil)
	// expr, _ = assembler.TryEvaluateExpression(expr)
	// fmt.Println(expr.Emit())
	
}
