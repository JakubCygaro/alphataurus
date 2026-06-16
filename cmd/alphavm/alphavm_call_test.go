package alphavm

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	tc "github.com/JakubCygaro/alphataurus/internal/pkg/tests_commons"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
	// "github.com/JakubCygaro/alphataurus/pkg/vm"
)

func TestCall1(t *testing.T) {
	a, b := byte(rand.Int()), byte(rand.Int())
	asm := fmt.Sprintf(`
	section '.code'
	@entry
	_start:
		mov bp, sp
		add sp, 2
		mov BYTE [bp+1], %v
		mov BYTE [bp+2], %v
		mov r1b, [bp+1]
		mov r2b, [bp+2]
		call add
		exit r0b
	add:
		push bp
		mov bp, sp
		add r1b, r2b
		mov r0b, r1b
		pop bp
		ret
	`, a, b)
	if mach, err := assembleAndExecute(asm); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if exit := mach.GetExitCode(); exit != uint64(a+b) {
		t.Error(compilationOfErr(asm))
		t.Error(expectedExitCode(0, exit))
	}
}
func TestCall2(t *testing.T) {
	a, b := byte(rand.Int()), byte(rand.Int())
	asm := fmt.Sprintf(`
	import 'add'
	section '.code'
	@entry
	_start:
		mov bp, sp
		add sp, 2
		mov BYTE [bp+1], %v
		mov BYTE [bp+2], %v
		mov r1b, [bp+1]
		mov r2b, [bp+2]
		call add
		exit r0b
	`, a, b)
	asm2 := `
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
	if elf, err := assembleAndLink(asm, asm2); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if mach, err := executeStackSize(elf, 32); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if exit := mach.GetExitCode(); exit != uint64(a+b) {
		t.Error(compilationOfErr(asm))
		t.Error(expectedExitCode(0, exit))
	}
}
func TestCall3(t *testing.T) {
	a := byte(rand.Int())
	asm := `
	import 'foo'
	section '.code'
	@entry
	_start:
		call foo
		exit r0b
	`
	asm2 := `
	export 'foo'
	import 'bar'
	section '.code'
	foo:
		push bp
		mov bp, sp
		call bar
		pop bp
		ret
	`
	asm3 := fmt.Sprintf(`
	export 'bar'
	section '.code'
	bar:
		push bp
		mov bp, sp
		mov r0b, %v
		pop bp
		ret
	`, a)
	if elf, err := assembleAndLink(asm, asm2, asm3); err != nil {
		t.Error(compilationOfErr(asm, asm2, asm3))
		t.Error(err)
	} else if mach, err := executeStackSize(elf, 32); err != nil {
		t.Error(compilationOfErr(asm, asm2, asm3))
		t.Error(err)
	} else if exit := mach.GetExitCode(); exit != uint64(a) {
		t.Error(compilationOfErr(asm, asm2, asm3))
		t.Error(expectedExitCode(uint64(a), exit))
	}
}
func TestCall4(t *testing.T) {
	type asmFn struct {
		asm, fn string
		incr    int
	}
	rA := tc.RandomGpRegisterWord()
	res := 0
	functions := make([]asmFn, 0)
	for i := range rand.Intn(10) {
		fn := fmt.Sprintf("autogen_func_%d", i)
		incr := rand.Intn(5000) - 5000/2
		asm := fmt.Sprintf(`
		export '%s'
		section '.code'
		%s:
			add SIGNED %s, %v
			ret
		`,
			fn,
			fn,
			tc.RegStr(rA, vm.SZ_64), incr,
		)
		res += incr
		functions = append(functions, asmFn{asm, fn, incr})
	}
	entryLines := make([]string, 0)
	sources := make([]string, 0)
	for _, f := range functions {
		entryLines = append(entryLines,
			fmt.Sprintf("import '%s'", f.fn))
		sources = append(sources, f.asm)
	}
	entryLines = append(entryLines,
		"section '.code'",
		"@entry",
		"_start:",
	)
	for _, f := range functions {
		entryLines = append(entryLines,
			fmt.Sprintf("call %s", f.fn))
	}
	entryLines = append(entryLines,
		fmt.Sprintf("exit %s", tc.RegStr(rA, vm.SZ_64)),
	)
	entry := strings.Join(entryLines, "\n")
	sources = append(sources, entry)
	if elf, err := assembleAndLink(sources...); err != nil {
		t.Error(compilationOfErr(sources...))
		t.Error(err)
	} else if mach, err := executeStackSize(elf, 32); err != nil {
		t.Error(compilationOfErr(sources...))
		t.Error(err)
	} else if exit := mach.GetExitCode(); exit != uint64(res) {
		t.Error(compilationOfErr(sources...))
		t.Error(expectedExitCode(uint64(res), exit))
	}
}
