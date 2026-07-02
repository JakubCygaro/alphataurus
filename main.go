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
@entry
_start:
    mov r1q, chuj
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
	err = mach.Execute(elf)
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(-1)
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
			Variables: make(map[string]any),
		},
	}
	if eval, ok := ev.TryEvaluateExpression(ex); !ok {
		fmt.Fprintln(os.Stderr, ok)
	} else if em, err := eval.Emit(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Println(em)
	}
	// expr, _ = pr.TryEvaluatePruneExpression(expr)
	// fmt.Println((((6594007686923535256 / (6948242974143 - 1987936890282)) / 3106268) / 1758))
	// os.Exit(int(mach.GetExitCode()))
}
