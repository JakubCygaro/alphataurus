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
	mov r0, 69
	mov r1, 420
	add r0, r1
	push r0
	pop r4
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
