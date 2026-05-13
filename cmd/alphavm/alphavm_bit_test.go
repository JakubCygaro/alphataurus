package alphavm

import (
	// "fmt"
	"testing"

)

func TestLsh1(t *testing.T) {
	asm := `
	section '.code'
	@entry
	start:
		mov r7, 1
		lsh r7, 1
		mov r7, r0
		mov r1, 2
		div UNSIGNED WORD
		cmp r3, 1
		jne fail
		lsh r7, 1
		mov r7, r0
		mov r1, 2
		div UNSIGNED WORD
		cmp r3, 2
		jne fail
		exit 0
	fail:
		exit 1
	`
	if mach, err := assembleAndExecute(asm); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if exit := mach.GetExitCode(); exit != 0 {
		t.Error(compilationOfErr(asm))
		t.Error(expectedExitCode(0, exit))
	}
}
