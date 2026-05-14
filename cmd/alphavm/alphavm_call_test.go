package alphavm

import (
	"fmt"
	"math/rand"
	"testing"

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
	} else if exit := mach.GetExitCode(); exit != uint64(a + b) {
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
	} else if exit := mach.GetExitCode(); exit != uint64(a + b) {
		t.Error(compilationOfErr(asm))
		t.Error(expectedExitCode(0, exit))
	}
}
