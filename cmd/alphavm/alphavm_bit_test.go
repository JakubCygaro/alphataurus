package alphavm

import (
	// "fmt"
	"fmt"
	"math/rand"
	"testing"

	"github.com/JakubCygaro/alphataurus/pkg/vm"
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
func TestNot1(t *testing.T) {
	const stackSize int = 1 + 2 + 4 + 8
	stack := makeTestingStack(stackSize)
	initV := rand.Int63()
	stack.push(^uint8(initV))
	stack.push(^uint16(initV))
	stack.push(^uint32(initV))
	stack.push(^uint64(initV))
	rA := randomGpRegisterWord()
	rB := nextRandomGpRegister(rA)
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov %s, %v
		;; not as SZ_8
		mov %s, %s
		not %s
		push %s
		;; not as SZ_16
		mov %s, %s
		not %s
		push %s
		;; not as SZ_32
		mov %s, %s
		not %s
		push %s
		;; not as SZ_64
		mov %s, %s
		not %s
		push %s
		exit 0
	`,
		regStr(rA, vm.SZ_64), initV,
		regStr(rB, vm.SZ_8), regStr(rA, vm.SZ_8),
		regStr(rB, vm.SZ_8),
		regStr(rB, vm.SZ_8),
		regStr(rB, vm.SZ_16), regStr(rA, vm.SZ_16),
		regStr(rB, vm.SZ_16),
		regStr(rB, vm.SZ_16),
		regStr(rB, vm.SZ_32), regStr(rA, vm.SZ_32),
		regStr(rB, vm.SZ_32),
		regStr(rB, vm.SZ_32),
		regStr(rB, vm.SZ_64), regStr(rA, vm.SZ_64),
		regStr(rB, vm.SZ_64),
		regStr(rB, vm.SZ_64),
	)
	if elf, err := assembleAndLink(asm); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if mach, err := executeStackSize(elf, uint64(stackSize)); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if err := expectStack(&mach, vm.VmStack(stack)); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	}
}
func TestAnd1(t *testing.T) {
	const stackSize int = 1 + 2 + 4 + 8
	stack := makeTestingStack(stackSize)
	initV := rand.Int63()
	sV := rand.Int63()
	stack.push(uint8(initV) & uint8(sV))
	stack.push(uint16(initV) & uint16(sV))
	stack.push(uint32(initV) & uint32(sV))
	stack.push(uint64(initV) & uint64(sV))
	rA := randomGpRegisterWord()
	rB := nextRandomGpRegister(rA)
	rC := nextRandomGpRegister(rB)
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov %s, %v
		mov %s, %v
		;; AND SZ_8
		mov %s, %s
		and %s, %s
		push %s
		;; AND SZ_16
		mov %s, %s
		and %s, %s
		push %s
		;; AND SZ_32
		mov %s, %s
		and %s, %s
		push %s
		;; AND SZ_64
		mov %s, %s
		and %s, %s
		push %s
		exit 0
	`,
		regStr(rA, vm.SZ_64), initV,
		regStr(rB, vm.SZ_64), sV,
		regStr(rC, vm.SZ_8), regStr(rA, vm.SZ_8),
		regStr(rC, vm.SZ_8), regStr(rB, vm.SZ_8),
		regStr(rC, vm.SZ_8),
		regStr(rC, vm.SZ_16), regStr(rA, vm.SZ_16),
		regStr(rC, vm.SZ_16), regStr(rB, vm.SZ_16),
		regStr(rC, vm.SZ_16),
		regStr(rC, vm.SZ_32), regStr(rA, vm.SZ_32),
		regStr(rC, vm.SZ_32), regStr(rB, vm.SZ_32),
		regStr(rC, vm.SZ_32),
		regStr(rC, vm.SZ_64), regStr(rA, vm.SZ_64),
		regStr(rC, vm.SZ_64), regStr(rB, vm.SZ_64),
		regStr(rC, vm.SZ_64),
	)
	if elf, err := assembleAndLink(asm); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if mach, err := executeStackSize(elf, uint64(stackSize)); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if err := expectStack(&mach, vm.VmStack(stack)); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	}
}
func TestOr1(t *testing.T) {
	const stackSize int = 1 + 2 + 4 + 8
	stack := makeTestingStack(stackSize)
	initV := rand.Int63()
	sV := rand.Int63()
	stack.push(uint8(initV) | uint8(sV))
	stack.push(uint16(initV) | uint16(sV))
	stack.push(uint32(initV) | uint32(sV))
	stack.push(uint64(initV) | uint64(sV))
	rA := randomGpRegisterWord()
	rB := nextRandomGpRegister(rA)
	rC := nextRandomGpRegister(rB)
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov %s, %v
		mov %s, %v
		;; OR SZ_8
		mov %s, %s
		or %s, %s
		push %s
		;; OR SZ_16
		mov %s, %s
		or %s, %s
		push %s
		;; OR SZ_32
		mov %s, %s
		or %s, %s
		push %s
		;; OR SZ_64
		mov %s, %s
		or %s, %s
		push %s
		exit 0
	`,
		regStr(rA, vm.SZ_64), initV,
		regStr(rB, vm.SZ_64), sV,
		regStr(rC, vm.SZ_8), regStr(rA, vm.SZ_8),
		regStr(rC, vm.SZ_8), regStr(rB, vm.SZ_8),
		regStr(rC, vm.SZ_8),
		regStr(rC, vm.SZ_16), regStr(rA, vm.SZ_16),
		regStr(rC, vm.SZ_16), regStr(rB, vm.SZ_16),
		regStr(rC, vm.SZ_16),
		regStr(rC, vm.SZ_32), regStr(rA, vm.SZ_32),
		regStr(rC, vm.SZ_32), regStr(rB, vm.SZ_32),
		regStr(rC, vm.SZ_32),
		regStr(rC, vm.SZ_64), regStr(rA, vm.SZ_64),
		regStr(rC, vm.SZ_64), regStr(rB, vm.SZ_64),
		regStr(rC, vm.SZ_64),
	)
	if elf, err := assembleAndLink(asm); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if mach, err := executeStackSize(elf, uint64(stackSize)); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if err := expectStack(&mach, vm.VmStack(stack)); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	}
}
