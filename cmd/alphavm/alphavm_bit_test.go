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
// this tests packs 8 bytes from the stack into r0 register and then compares it
// as r0b to these values on the stack
func TestLshRsh1(t *testing.T) {
	asm := `
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
	if mach, err := assembleAndExecute(asm); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if exit := mach.GetExitCode(); exit != 0 {
		t.Error(compilationOfErr(asm))
		t.Error(expectedExitCode(0, exit))
	}
}
