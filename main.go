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
@entry
	push BYTE 69
	mov bp, sp
	mov WORD [bp],    0x0000000043434343
	mov HALF [bp],    0x00004545
	mov QUARTER [bp], 0x002c
	mov BYTE [bp],    0x1a
	mov r1b, [bp]
	mov r2q, [bp]
	and r2q, 0x00ff
	mov r3h, [bp]
	and r3h, 0x0000ffff
	mov r4, [bp]
	and r4,  0x00000000ffffffff
	mov r5b, [bp]
	sub r1b, r5b
	exit r1b
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
