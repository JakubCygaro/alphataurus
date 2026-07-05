package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/JakubCygaro/alphataurus/pkg/assembler"
	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
	"github.com/JakubCygaro/alphataurus/pkg/linker"
	vm "github.com/JakubCygaro/alphataurus/pkg/vm"
)

const assembly3 = `
export 'bar'
section '.code'
bar:
	push bp
	mov bp, sp
	mov r0b, 184
	pop bp
	ret
`

const assembly2 = `
export 'foo'
import 'bar'
section '.code'
foo:
	push bp
	mov bp, sp
	call bar
	mov r2, 14
	pop bp
	ret
`
const assembly = `
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

func logWarnings(awd assembler.AssemblerWarningData) {
	fmt.Fprintf(os.Stdout,
		"Assembler warning at (%v:%v): %s.\n",
		awd.Line,
		awd.Col,
		awd.Message,
	)
}

func main() {
	sources := []string{assembly}
	objects := make([]linker.LinkerInput, 0)
	for _, s := range sources {
		asm := assembler.NewAssembler(bufio.NewReader(strings.NewReader(s)))
		asm.WarningSink = logWarnings
		bytecode, err := asm.Assemble()
		if err != nil {
			os.Stderr.WriteString("Assembler error: ")
			os.Stderr.WriteString(err.Error())
			os.Stderr.WriteString("\n")
			os.Exit(-1)
		}
		iCount := asm.InstructionCount()
		fmt.Printf("Emitted bytecode size: %d\n", len(bytecode))
		fmt.Printf("Emitted %d instructions\n", iCount)
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
	mach := vm.CreateVmState(64)
	mach.SetOpCodeTrace(func(op vm.OpCodeVal) {
		fmt.Printf("[%s]\n", op.String())
	})
	var execError error
	err = mach.Execute(elf)
	if err != nil {
		execError = err
	}
	if rx, err := mach.GetGpRXAsS64(vm.R0_IDX); err == nil {
		fmt.Printf("r0 = %+v\n", int8(rx))
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
	const in = "r1 + r1 - (1 + 2)"
	var ex *pr.Expr
	p := pr.NewParser(bufio.NewReader(strings.NewReader(in)))
	if expr, err := p.ParseExpression(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Println(expr.Emit())
		ex = expr
	}
	ev := assembler.ExpressionEvaluator{
		Ctx: assembler.EvaluationContext{
			VarProvider: nil,
		},
	}
	if eval, err := ev.TryEvaluateExpression(ex); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else if em, err := eval.Emit(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Println(em)
	}
	// expr, _ = pr.TryEvaluatePruneExpression(expr)
	// fmt.Println((((6594007686923535256 / (6948242974143 - 1987936890282)) / 3106268) / 1758))
	if execError != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	}
	os.Exit(int(mach.GetExitCode()))
}
