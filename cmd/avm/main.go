package main

import (
	"fmt"
	"os"
	// "path/filepath"
	// "strings"

	vm "github.com/JakubCygaro/alphataurus/pkg/vm"
	aelf "github.com/JakubCygaro/alphataurus/pkg/vm/aelf"
	arg "github.com/alexflint/go-arg"
)

var args struct {
	Input             string `arg:"positional,required"`
	StackSize         uint   `arg:"-s,--stack-size" default:"1024"`
	DumpRegisters     bool   `arg:"--dump-registers" default:"false"`
	TraceInstructions bool   `arg:"-t,--trace-instructions" default:"false"`
}

func main() {
	arg.MustParse(&args)
	if bytes, err := os.ReadFile(args.Input); err != nil {
		fmt.Fprintf(os.Stderr, "Could not read file `%s`\n", args.Input)
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	} else if elf, err := aelf.LoadAELF(bytes); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	} else {
		mach := vm.CreateVmState(uint64(args.StackSize))
		if args.TraceInstructions {
			mach.SetOpCodeTrace(func(inst vm.OpCodeVal) {
				printlnTrace("[%s]", inst.String())
			})
		}
		var exit int
		if err := mach.Execute(elf); err != nil {
			os.Stderr.WriteString(err.Error())
			os.Stderr.WriteString("\n")
			exit = 1
		} else {
			exit = int(mach.GetExitCode())
		}
		if args.DumpRegisters {
			dumpRegisters(&mach)
		}
		os.Exit(exit)
	}
}
func dumpRegisters(mach *vm.VmState) {
	printlnPostExecution("REGISTERS")
	for i, v := range mach.GetRegisters() {
		printlnPostExecution("%v\t%+v", i, v.ToDisplayString())
	}
	printlnPostExecution("FLAGS")
	printlnPostExecution("[%+v]", mach.GetFlags())
}

func printlnPostExecution(format string, a ...any) {
	fmt.Fprintf(os.Stdout, "[PE]==> %s\n", fmt.Sprintf(format, a...))
}
func printlnTrace(format string, a ...any) {
	fmt.Fprintf(os.Stdout, "[Trace]~> %s\n", fmt.Sprintf(format, a...))
}
