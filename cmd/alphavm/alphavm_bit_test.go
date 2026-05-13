package alphavm

import (
	// "fmt"
	"testing"

)

func TestLsh1(t *testing.T) {
	asm := `
	section '.code'
	fail:
		exit 1
	@entry
		mov r7, 1
		lsh r7, 1
		mov r1, 2 ;; divisor
		mov r4, 1 ;; comparer

	L0:
		cmp r7, 1024
		jg leave
		mov r0, r7
		div UNSIGNED WORD
		cmp r2, r4
		jne fail
		lsh r7, 1
		lsh r4, 1
	jmp L0

	leave:
		exit 0
	`
	if mach, err := assembleAndExecute(asm); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if exit := mach.GetExitCode(); exit != 0 {
		t.Error(compilationOfErr(asm))
		t.Error(expectedExitCode(0, exit))
	}
}
