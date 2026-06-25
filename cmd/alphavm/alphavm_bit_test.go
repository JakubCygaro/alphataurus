package alphavm

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	tc "github.com/JakubCygaro/alphataurus/internal/pkg/tests_commons"
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
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
	rA := tc.RandomGpRegisterWord()
	rB := tc.NextRandomGpRegister(rA)
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
		tc.RegStr(rA, vm.SZ_64), initV,
		tc.RegStr(rB, vm.SZ_8), tc.RegStr(rA, vm.SZ_8),
		tc.RegStr(rB, vm.SZ_8),
		tc.RegStr(rB, vm.SZ_8),
		tc.RegStr(rB, vm.SZ_16), tc.RegStr(rA, vm.SZ_16),
		tc.RegStr(rB, vm.SZ_16),
		tc.RegStr(rB, vm.SZ_16),
		tc.RegStr(rB, vm.SZ_32), tc.RegStr(rA, vm.SZ_32),
		tc.RegStr(rB, vm.SZ_32),
		tc.RegStr(rB, vm.SZ_32),
		tc.RegStr(rB, vm.SZ_64), tc.RegStr(rA, vm.SZ_64),
		tc.RegStr(rB, vm.SZ_64),
		tc.RegStr(rB, vm.SZ_64),
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
	rA := tc.RandomGpRegisterWord()
	rB := tc.NextRandomGpRegister(rA)
	rC := tc.NextRandomGpRegister(rB)
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
		tc.RegStr(rA, vm.SZ_64), initV,
		tc.RegStr(rB, vm.SZ_64), sV,
		tc.RegStr(rC, vm.SZ_8), tc.RegStr(rA, vm.SZ_8),
		tc.RegStr(rC, vm.SZ_8), tc.RegStr(rB, vm.SZ_8),
		tc.RegStr(rC, vm.SZ_8),
		tc.RegStr(rC, vm.SZ_16), tc.RegStr(rA, vm.SZ_16),
		tc.RegStr(rC, vm.SZ_16), tc.RegStr(rB, vm.SZ_16),
		tc.RegStr(rC, vm.SZ_16),
		tc.RegStr(rC, vm.SZ_32), tc.RegStr(rA, vm.SZ_32),
		tc.RegStr(rC, vm.SZ_32), tc.RegStr(rB, vm.SZ_32),
		tc.RegStr(rC, vm.SZ_32),
		tc.RegStr(rC, vm.SZ_64), tc.RegStr(rA, vm.SZ_64),
		tc.RegStr(rC, vm.SZ_64), tc.RegStr(rB, vm.SZ_64),
		tc.RegStr(rC, vm.SZ_64),
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
	rA := tc.RandomGpRegisterWord()
	rB := tc.NextRandomGpRegister(rA)
	rC := tc.NextRandomGpRegister(rB)
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
		tc.RegStr(rA, vm.SZ_64), initV,
		tc.RegStr(rB, vm.SZ_64), sV,
		tc.RegStr(rC, vm.SZ_8), tc.RegStr(rA, vm.SZ_8),
		tc.RegStr(rC, vm.SZ_8), tc.RegStr(rB, vm.SZ_8),
		tc.RegStr(rC, vm.SZ_8),
		tc.RegStr(rC, vm.SZ_16), tc.RegStr(rA, vm.SZ_16),
		tc.RegStr(rC, vm.SZ_16), tc.RegStr(rB, vm.SZ_16),
		tc.RegStr(rC, vm.SZ_16),
		tc.RegStr(rC, vm.SZ_32), tc.RegStr(rA, vm.SZ_32),
		tc.RegStr(rC, vm.SZ_32), tc.RegStr(rB, vm.SZ_32),
		tc.RegStr(rC, vm.SZ_32),
		tc.RegStr(rC, vm.SZ_64), tc.RegStr(rA, vm.SZ_64),
		tc.RegStr(rC, vm.SZ_64), tc.RegStr(rB, vm.SZ_64),
		tc.RegStr(rC, vm.SZ_64),
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
func TestXor1(t *testing.T) {
	const stackSize int = 1 + 2 + 4 + 8
	stack := makeTestingStack(stackSize)
	initV := rand.Int63()
	sV := rand.Int63()
	stack.push(uint8(initV) ^ uint8(sV))
	stack.push(uint16(initV) ^ uint16(sV))
	stack.push(uint32(initV) ^ uint32(sV))
	stack.push(uint64(initV) ^ uint64(sV))
	rA := tc.RandomGpRegisterWord()
	rB := tc.NextRandomGpRegister(rA)
	rC := tc.NextRandomGpRegister(rB)
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov %s, %v
		mov %s, %v
		;; XOR SZ_8
		mov %s, %s
		xor %s, %s
		push %s
		;; XOR SZ_16
		mov %s, %s
		xor %s, %s
		push %s
		;; XOR SZ_32
		mov %s, %s
		xor %s, %s
		push %s
		;; XOR SZ_64
		mov %s, %s
		xor %s, %s
		push %s
		exit 0
	`,
		tc.RegStr(rA, vm.SZ_64), initV,
		tc.RegStr(rB, vm.SZ_64), sV,
		tc.RegStr(rC, vm.SZ_8), tc.RegStr(rA, vm.SZ_8),
		tc.RegStr(rC, vm.SZ_8), tc.RegStr(rB, vm.SZ_8),
		tc.RegStr(rC, vm.SZ_8),
		tc.RegStr(rC, vm.SZ_16), tc.RegStr(rA, vm.SZ_16),
		tc.RegStr(rC, vm.SZ_16), tc.RegStr(rB, vm.SZ_16),
		tc.RegStr(rC, vm.SZ_16),
		tc.RegStr(rC, vm.SZ_32), tc.RegStr(rA, vm.SZ_32),
		tc.RegStr(rC, vm.SZ_32), tc.RegStr(rB, vm.SZ_32),
		tc.RegStr(rC, vm.SZ_32),
		tc.RegStr(rC, vm.SZ_64), tc.RegStr(rA, vm.SZ_64),
		tc.RegStr(rC, vm.SZ_64), tc.RegStr(rB, vm.SZ_64),
		tc.RegStr(rC, vm.SZ_64),
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
func TestLogIR1(t *testing.T) {
	into := makeIntoRegistersList()
	lines := make([]string, 0)
	lines = append(lines,
		"section '.code'",
		"@entry",
		"_start:",
	)
	for _, ir := range into {
		if !vm.IsLogRAllowed(byte(ir.Reg)) ||
			ir.Reg == vm.SP_IDX {
			continue
		}
		var a, b uint64
		switch ir.Size {
		case vm.SZ_8:
			a, b = uint64(byte(rand.Int63())), uint64(byte(rand.Int63()))
		case vm.SZ_16:
			a, b = uint64(uint16(rand.Int63())), uint64(uint16(rand.Int63()))
		case vm.SZ_32:
			a, b = uint64(uint32(rand.Int63())), uint64(uint32(rand.Int63()))
		case vm.SZ_64:
			a, b = uint64(rand.Int63()), uint64(rand.Int63())
		}
		lines = append(lines,
			fmt.Sprintf(
				"mov %s, %v",
				tc.RegStr(byte(ir.Reg), ir.Size),
				a,
			),
			fmt.Sprintf(
				"and %s, %v",
				tc.RegStr(byte(ir.Reg), ir.Size),
				b,
			),
			macroAssertEqRI(
				lx.RegisterData(ir),
				a&b,
				UNSIGNED,
			),
			fmt.Sprintf(
				"mov %s, %v",
				tc.RegStr(byte(ir.Reg), ir.Size),
				a,
			),
			fmt.Sprintf(
				"or %s, %v",
				tc.RegStr(byte(ir.Reg), ir.Size),
				b,
			),
			macroAssertEqRI(
				lx.RegisterData(ir),
				a|b,
				UNSIGNED,
			),
			fmt.Sprintf(
				"mov %s, %v",
				tc.RegStr(byte(ir.Reg), ir.Size),
				a,
			),
			fmt.Sprintf(
				"xor %s, %v",
				tc.RegStr(byte(ir.Reg), ir.Size),
				b,
			),
			macroAssertEqRI(
				lx.RegisterData(ir),
				a^b,
				UNSIGNED,
			),
			fmt.Sprintf(
				"mov %s, %v",
				tc.RegStr(byte(ir.Reg), ir.Size),
				a,
			),
			fmt.Sprintf(
				"lsh %s, %v",
				tc.RegStr(byte(ir.Reg), ir.Size),
				b,
			),
			macroAssertEqRI(
				lx.RegisterData(ir),
				a<<b,
				UNSIGNED,
			),
			fmt.Sprintf(
				"mov %s, %v",
				tc.RegStr(byte(ir.Reg), ir.Size),
				a,
			),
			fmt.Sprintf(
				"rsh %s, %v",
				tc.RegStr(byte(ir.Reg), ir.Size),
				b,
			),
			macroAssertEqRI(
				lx.RegisterData(ir),
				a>>b,
				UNSIGNED,
			),
		)
	}
	lines = append(lines,
		"exit 0")
	asm := strings.Join(lines, "\n")
	if mach, err := assembleAndExecute(asm); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if exitCode := mach.GetExitCode(); exitCode != 0 {
		t.Error(compilationOfErr(asm))
		t.Error(expectedExitCode(0, exitCode))
	}
}
func TestLogIRR(t *testing.T) {
	into := makeIntoRegistersList()
	lines := make([]string, 0)
	lines = append(lines,
		"section '.code'",
		"@entry",
		"_start:",
	)
	for _, ar := range into {
		for _, br := range into {
			if !vm.IsLogRAllowed(byte(ar.Reg)) ||
				ar.Reg == vm.SP_IDX ||
				br.Reg == vm.SP_IDX ||
				ar.Size != br.Size ||
				ar.Reg == br.Reg {
				continue
			}
			var a, b uint64
			switch ar.Size {
			case vm.SZ_8:
				a, b = uint64(byte(rand.Int63())), uint64(byte(rand.Int63()))
			case vm.SZ_16:
				a, b = uint64(uint16(rand.Int63())), uint64(uint16(rand.Int63()))
			case vm.SZ_32:
				a, b = uint64(uint32(rand.Int63())), uint64(uint32(rand.Int63()))
			case vm.SZ_64:
				a, b = uint64(rand.Int63()), uint64(rand.Int63())
			}
			lines = append(lines,
				fmt.Sprintf(
					"mov %s, %v",
					tc.RegStr(byte(ar.Reg), ar.Size),
					a,
				),
				fmt.Sprintf(
					"mov %s, %v",
					tc.RegStr(byte(br.Reg), br.Size),
					b,
				),
				fmt.Sprintf(
					"and %s, %s",
					tc.RegStr(byte(ar.Reg), ar.Size),
					tc.RegStr(byte(br.Reg), br.Size),
				),
				macroAssertEqRI(
					lx.RegisterData(ar),
					a&b,
					UNSIGNED,
				),
				fmt.Sprintf(
					"mov %s, %v",
					tc.RegStr(byte(ar.Reg), ar.Size),
					a,
				),
				fmt.Sprintf(
					"or %s, %s",
					tc.RegStr(byte(ar.Reg), ar.Size),
					tc.RegStr(byte(br.Reg), br.Size),
				),
				macroAssertEqRI(
					lx.RegisterData(ar),
					a|b,
					UNSIGNED,
				),
				fmt.Sprintf(
					"mov %s, %v",
					tc.RegStr(byte(ar.Reg), ar.Size),
					a,
				),
				fmt.Sprintf(
					"xor %s, %s",
					tc.RegStr(byte(ar.Reg), ar.Size),
					tc.RegStr(byte(br.Reg), br.Size),
				),
				macroAssertEqRI(
					lx.RegisterData(ar),
					a^b,
					UNSIGNED,
				),
				fmt.Sprintf(
					"mov %s, %v",
					tc.RegStr(byte(ar.Reg), ar.Size),
					a,
				),
				fmt.Sprintf(
					"lsh %s, %s",
					tc.RegStr(byte(ar.Reg), ar.Size),
					tc.RegStr(byte(br.Reg), br.Size),
				),
				macroAssertEqRI(
					lx.RegisterData(ar),
					a<<b,
					UNSIGNED,
				),
				fmt.Sprintf(
					"mov %s, %v",
					tc.RegStr(byte(ar.Reg), ar.Size),
					a,
				),
				fmt.Sprintf(
					"rsh %s, %s",
					tc.RegStr(byte(ar.Reg), ar.Size),
					tc.RegStr(byte(br.Reg), br.Size),
				),
				macroAssertEqRI(
					lx.RegisterData(ar),
					a>>b,
					UNSIGNED,
				),
			)
		}
	}
	lines = append(lines,
		"exit 0")
	asm := strings.Join(lines, "\n")
	if mach, err := assembleAndExecute(asm); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if exitCode := mach.GetExitCode(); exitCode != 0 {
		t.Error(compilationOfErr(asm))
		t.Error(expectedExitCode(0, exitCode))
	}
}
