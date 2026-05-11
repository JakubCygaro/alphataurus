package alphavm

import (
	"fmt"
	"math"
	"regexp"
	"testing"

	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func TestMov1(t *testing.T) {
	r0_v, r1_v, r2_v, r3_v := 69, 420, 1.23, 1.23
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov r0, %d
		mov r1, %d
		mov r2, %f
		mov r3, r2
	`, r0_v, r1_v, r2_v)
	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R0_IDX: vm.RegisterWithValue(uint64(r0_v)),
		vm.R1_IDX: vm.RegisterWithValue(uint64(r1_v)),
		vm.R2_IDX: vm.RegisterWithValue(math.Float64bits(r2_v)),
		vm.R3_IDX: vm.RegisterWithValue(math.Float64bits(r3_v)),
	}); err != nil {
		t.Errorf(err.Error())
	}
}
func TestMov2(t *testing.T) {
	r0_v, r1_v, r2_v, r3_v := 69, 420, 123, 123
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov r0b, %d
		mov r1q, %d
		mov r2h, %d
		mov r3, r2
	`, r0_v, r1_v, r2_v)
	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R0_IDX: vm.RegisterWithValueSized(uint64(r0_v), vm.SZ_8),
		vm.R1_IDX: vm.RegisterWithValueSized(uint64(r1_v), vm.SZ_16),
		vm.R2_IDX: vm.RegisterWithValueSized(uint64(r2_v), vm.SZ_32),
		vm.R3_IDX: vm.RegisterWithValue(uint64(r3_v)),
	}); err != nil {
		t.Errorf(err.Error())
	}
}
func TestMov3(t *testing.T) {
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		push WORD 0
		mov bp, sp
		mov WORD [bp], 0x1111111100000000
		mov HALF [bp], 0x4d3c0000
		mov QUARTER [bp], 0x2b00
		mov BYTE [bp], 0x1a
	`)
	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	// if err := expectGpRegisters(asm, &mach, ExpMap{
	// 	vm.R0_IDX: vm.RegisterWithValueSized(uint64(r0_v), vm.SZ_8),
	// 	vm.R1_IDX: vm.RegisterWithValueSized(uint64(r1_v), vm.SZ_16),
	// 	vm.R2_IDX: vm.RegisterWithValueSized(uint64(r2_v), vm.SZ_32),
	// 	vm.R3_IDX: vm.RegisterWithValue(uint64(r3_v)),
	// }); err != nil {
	// 	t.Errorf(err.Error())
	// }
}
func TestMov1F(t *testing.T) {
	rA, rAsz := randomGpRegisterWord(), byte(vm.SZ_8)
	rB, rBsz := nextRandomGpRegister(rA), byte(vm.SZ_64)
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov %s, %s
	`, regStr(rA, rAsz), regStr(rB, rBsz))
	if _, err := assemble(asm); err != nil {
		if ok, _ := regexp.MatchString("Mismatched register sizes", err.Error()); !ok {
			t.Error(compilationOfErr(asm))
			t.Errorf("Expected mismatched register sizes assembling error, " +
				"got a different error")
		}
	} else {
		t.Error(compilationOfErr(asm))
		t.Error(assemblingErrorExpected())
	}
}
