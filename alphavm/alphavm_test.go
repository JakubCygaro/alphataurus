package alphavm

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"testing"

	"github.com/JakubCygaro/alphataurus/assembler"
	"github.com/JakubCygaro/alphataurus/internal/vm"
)

func TestMov1(t *testing.T) {
	r0_v, r1_v, r2_v, r3_v := 69, 420, 1.23, 1.23
	asm := fmt.Sprintf(`
	mov r0, %d
	mov r1, %d
	mov r2, %f
	mov r3, r2
	`, r0_v, r1_v, r2_v)
	asmblr := assembler.NewAssembler(*bufio.NewReader(strings.NewReader(asm)))
	mach := vm.CreateVmState(16)
	code, _, err := asmblr.EmitBytecode()
	if err != nil {
		t.Error(err)
	}
	err = mach.Execute(code)
	if err != nil {
		t.Error(err)
	}
	var r0, r1, r2, r3 any
	err = mach.GetGpRXAs(vm.R0_IDX, vm.TY_UINT64, &r0)
	if err != nil {
		t.Error(err)
	}
	err = mach.GetGpRXAs(vm.R1_IDX, vm.TY_UINT64, &r1)
	if err != nil {
		t.Error(err)
	}
	err = mach.GetGpRXAs(vm.R2_IDX, vm.TY_FLOAT64, &r2)
	if err != nil {
		t.Error(err)
	}
	err = mach.GetGpRXAs(vm.R3_IDX, vm.TY_FLOAT64, &r3)
	if err != nil {
		t.Error(err)
	}
	if r0.(uint64) != uint64(r0_v) ||
		r1.(uint64) != uint64(r1_v) ||
		r2.(float64) != r2_v ||
		r3.(float64) != r3_v {
		t.Errorf("Register state not what it was supposed to be")
	}
}
func TestAddRR1(t *testing.T) {
	r0_v, r1_v := 69, 420
	r2_v := r0_v + r1_v
	r3_v := 80.085
	r4_v := 133.7
	r5_v := r3_v + r4_v
	asm := fmt.Sprintf(`
	mov r0, %d
	mov r1, %d
	add UNSIGNED r0, r1
	mov r2, r0
	mov r3, %f
	mov r4, %f
	add FLOAT r3, r4
	mov r5, r3
	`, r0_v, r1_v, r3_v, r4_v)
	asmblr := assembler.NewAssembler(*bufio.NewReader(strings.NewReader(asm)))
	mach := vm.CreateVmState(16)
	code, _, err := asmblr.EmitBytecode()
	if err != nil {
		t.Error(err)
	}
	err = mach.Execute(code)
	if err != nil {
		t.Error(err)
	}
	var r2, r5 any
	err = mach.GetGpRXAs(vm.R2_IDX, vm.TY_UINT64, &r2)
	if err != nil {
		t.Error(err)
	}
	err = mach.GetGpRXAs(vm.R5_IDX, vm.TY_FLOAT64, &r5)
	if err != nil {
		t.Error(err)
	}
	if r2.(uint64) != uint64(r2_v) ||
		r5.(float64) != float64(r5_v) {
		t.Errorf("Register state not what it was supposed to be")
	}
}
func TestAddRR2Signed(t *testing.T) {
	r0_v, r1_v := 69, -420
	r2_v := r0_v + r1_v
	r3_v := 80.085
	r4_v := -133.7
	r5_v := r3_v + r4_v
	asm := fmt.Sprintf(`
	mov r0, %d
	mov r1, %d
	add SIGNED r0, r1
	mov r2, r0
	mov r3, %f
	mov r4, %f
	add FLOAT r3, r4
	mov r5, r3
	`, r0_v, r1_v, r3_v, r4_v)
	asmblr := assembler.NewAssembler(*bufio.NewReader(strings.NewReader(asm)))
	mach := vm.CreateVmState(16)
	code, _, err := asmblr.EmitBytecode()
	if err != nil {
		t.Error(err)
	}
	err = mach.Execute(code)
	if err != nil {
		t.Error(err)
	}
	var r2, r5 any
	err = mach.GetGpRXAs(vm.R2_IDX, vm.TY_UINT64, &r2)
	if err != nil {
		t.Error(err)
	}
	err = mach.GetGpRXAs(vm.R5_IDX, vm.TY_FLOAT64, &r5)
	if err != nil {
		t.Error(err)
	}
	if r2.(uint64) != uint64(r2_v) ||
		r5.(float64) != float64(r5_v) {
		t.Errorf("Register state not what it was supposed to be")
	}
}
func TestAddIR1(t *testing.T) {
	r0_v, r1_v := 69, -420
	r2_v := r0_v + r1_v
	r3_v := 80.085
	r4_v := -133.7
	r5_v := r3_v + r4_v
	asm := fmt.Sprintf(`
	mov r0, %d
	add SIGNED r0, %d
	mov r2, r0
	mov r3, %f
	add FLOAT r3, %f
	mov r5, r3
	`, r0_v, r1_v, r3_v, r4_v)
	asmblr := assembler.NewAssembler(*bufio.NewReader(strings.NewReader(asm)))
	mach := vm.CreateVmState(16)
	code, _, err := asmblr.EmitBytecode()
	if err != nil {
		t.Error(err)
	}
	err = mach.Execute(code)
	if err != nil {
		t.Error(err)
	}
	var r2, r5 any
	err = mach.GetGpRXAs(vm.R2_IDX, vm.TY_UINT64, &r2)
	if err != nil {
		t.Error(err)
	}
	err = mach.GetGpRXAs(vm.R5_IDX, vm.TY_FLOAT64, &r5)
	if err != nil {
		t.Error(err)
	}
	if r2.(uint64) != uint64(r2_v) ||
		r5.(float64) != float64(r5_v) {
		t.Errorf("Register state not what it was supposed to be")
	}
}
func randSign() int {
	return int(math.Ceil(rand.Float64() - 0.5))
}

func TestAddIR2(t *testing.T) {
	r := byte(rand.Int() % vm.GP_REG_MAX)
	reg_v := rand.Uint64();
	reg_add := rand.Uint64();
	asm := fmt.Sprintf(`
	mov r%v, %v
	add UNSIGNED r%v, %v
	`, r, reg_v, r, reg_add)

	asmblr := assembler.NewAssembler(*bufio.NewReader(strings.NewReader(asm)))
	mach := vm.CreateVmState(16)
	code, _, err := asmblr.EmitBytecode()
	if err != nil {
		t.Error(err)
	}
	err = mach.Execute(code)
	if err != nil {
		t.Error(err)
	}
	if v, _ := mach.GetGpRXAsUint64(r); v != reg_v + reg_add {
		t.Errorf("r0 value not what was desired (%v != %v)", v, reg_v + reg_add)
	}
}
func TestAddIR3(t *testing.T) {
	r := byte(rand.Int() % vm.GP_REG_MAX)
	reg_v := int64(rand.Uint64());
	reg_add := int64(rand.Uint64());
	asm := fmt.Sprintf(`
	mov r%v, %v
	add SIGNED r%v, %v
	`, r, reg_v, r, reg_add)

	asmblr := assembler.NewAssembler(*bufio.NewReader(strings.NewReader(asm)))
	mach := vm.CreateVmState(16)
	code, _, err := asmblr.EmitBytecode()
	if err != nil {
		t.Error(err)
	}
	err = mach.Execute(code)
	if err != nil {
		t.Error(err)
	}
	if v, _ := mach.GetGpRXAsInt64(r); v != reg_v + reg_add {
		t.Errorf("r0 value not what was desired (%v != %v)", v, reg_v + reg_add)
	}
}
func TestAddIR4(t *testing.T) {
	r := byte(rand.Int() % vm.GP_REG_MAX)
	reg_v := rand.Float64()
	reg_add := rand.Float64()
	asm := fmt.Sprintf(`
	mov r%v, %v
	add FLOAT r%v, %v
	`, r, reg_v, r, reg_add)

	asmblr := assembler.NewAssembler(*bufio.NewReader(strings.NewReader(asm)))
	mach := vm.CreateVmState(16)
	code, _, err := asmblr.EmitBytecode()
	if err != nil {
		t.Error(err)
	}
	err = mach.Execute(code)
	if err != nil {
		t.Error(err)
	}
	if v, _ := mach.GetGpRXAsFloat64(r); v != reg_v + reg_add {
		t.Errorf("r0 value not what was desired (%v != %v)", v, reg_v + reg_add)
	}
}
