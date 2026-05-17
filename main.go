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
export 'add'
section '.code'
add:
	push bp
	mov bp, sp
	add r1b, r2b
	mov r0b, r1b
	pop bp
	ret
`
const assembly = `
import 'add'
section '.code'
@entry
_start:
	mov bp, sp
	add sp, 2
	mov BYTE [bp+1], 17
	mov BYTE [bp+2], 193
	mov r1b, [bp+1]
	mov r2b, [bp+2]
	call add
	exit r0b
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
	mach := vm.CreateVmState(24)
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
	// os.Exit(int(mach.GetExitCode()))
}
