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
start:
	push 69
	add sp, 1
	push 420
	jmp start
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
	mach := vm.CreateVmState(4)
	err = mach.Execute(bytecode)
	if err != nil {
		fmt.Println(err)
	}
	r6, _ := mach.GetGpRXAsFloat64(vm.R6_IDX)
	fmt.Printf("r6 = %+v\n", r6)
	fmt.Printf("%+v\n", mach.GetRegisters())
	fmt.Printf("%+v\n", mach.GetFlags())

	expr := assembler.MakeDeref(
		assembler.MakeArth(
			assembler.MakeConstexprI64(2),
			assembler.MakeConstexprI64(2),
			assembler.ARTHEXPR_TADD,
		),
	)
	em, _ := expr.Emit()
	ev, _ := assembler.TryEvaluateExpression(&expr)
	fmt.Printf("%+v\n", em)
	v, _ := ev.Emit()
	fmt.Printf("%+v\n", v)
}
