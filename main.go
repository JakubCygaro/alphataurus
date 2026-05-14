package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/JakubCygaro/alphataurus/pkg/assembler"
	"github.com/JakubCygaro/alphataurus/pkg/linker"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

const assembly2 = `
export 'atoi'
section '.code'
atoi:
	push bp
	mov bp, sp
	sub r1b, 0x30
	mov r0b, r1b
	pop bp
	ret
`
const assembly = `
import 'atoi'
section '.code'
fail:
	exit 1
@entry
	mov bp, sp
	push BYTE 1
	push BYTE 2
	push BYTE 3
	push BYTE 4
	push BYTE 5
	push BYTE 6
	push BYTE 7
	push BYTE 8

	;; pack into r0

	mov r0b, [bp+1]
	lsh r0, 8
	mov r0b, [bp+2]
	lsh r0, 8
	mov r0b, [bp+3]
	lsh r0, 8
	mov r0b, [bp+4]
	lsh r0, 8
	mov r0b, [bp+5]
	lsh r0, 8
	mov r0b, [bp+6]
	lsh r0, 8
	mov r0b, [bp+7]
	lsh r0, 8
	mov r0b, [bp+8]

	;; check

	mov r1, 8 ;; loop counter

L0:
	cmp r1, 0 ;; loop check
	jle leave
	pop r2b
	cmp r0b, r2b
	jne fail
	rsh r0, 8
	dec r1
	jmp L0

leave:
	exit 0
`

func main() {
	sources := []string{assembly, assembly2}
	objects := make([]linker.LinkerInput, 0)
	for _, s := range sources {
		asm := assembler.NewAssembler(*bufio.NewReader(strings.NewReader(s)))
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
		objects = append(objects, linker.Bytes(bytecode))
	}
	ld := linker.NewLinker()
	elf, err := ld.Link(objects)
	os.WriteFile("dump", elf.Write(), os.FileMode(os.O_TRUNC))
	if err != nil {
		os.Stderr.WriteString("linking error\n")
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(-1)
	}
	mach := vm.CreateVmState(6 * 8)
	err = mach.Execute(elf)
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(-1)
	}
	if rx, err := mach.GetGpRXAsS64(vm.R0_IDX); err == nil {
		fmt.Printf("r0 = %+v\n", rx)
	}
	if rx, err := mach.GetGpRXAsU64(vm.R1_IDX); err == nil {
		fmt.Printf("r1 = %+v\n", rx)
	}
	if rx, err := mach.GetGpRXAsS64(vm.R2_IDX); err == nil {
		fmt.Printf("r2 = %+v\n", rx)
	}
	fmt.Printf("REGISTERS:\n")
	for i, v := range mach.GetRegisters() {
		fmt.Printf("%v\t%+v\n", i, v.ToDisplayString())
	}
	fmt.Printf("%+v\n", mach.GetFlags())
	fmt.Printf("stack:\n%+v\n", mach.GetStack())
	const in = "(-3 + 8)"
	p := assembler.NewParser(*bufio.NewReader(strings.NewReader(in)))
	expr, _ := p.ParseExpression()
	expr, _ = assembler.TryEvaluatePruneExpression(expr)
	fmt.Println(expr.Emit())
	os.Exit(int(mach.GetExitCode()))
}
