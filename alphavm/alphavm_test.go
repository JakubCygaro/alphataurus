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

func assembleAndExecute(source string) (vm.VmState, error) {
	asmblr := assembler.NewAssembler(*bufio.NewReader(strings.NewReader(source)))
	mach := vm.CreateVmState(16)
	code, _, err := asmblr.EmitBytecode()
	if err != nil {
		return mach, err
	}
	err = mach.Execute(code)
	if err != nil {
		return mach, err
	}
	return mach, nil
}
type ExpMap map[byte]uint64
func expectGpRegisters(t *testing.T, asm string, mach *vm.VmState, regStates ExpMap) {
	err := false
	for k, v := range regStates {
		if r, _ := mach.GetGpRXAsUint64(k); r != v {
			t.Errorf("State of general purpose register r%v was different from expected", k)
			t.Errorf("uint64  (%v != %v)", v, r)
			t.Errorf("int64   (%v != %v)", int64(v), int64(r))
			t.Errorf("float64 (%v != %v)", math.Float64frombits(v), math.Float64frombits(r))
			err = true
		}
	}
	if err {
		t.Errorf("Compilation of:%s", asm)
	}
}

func TestMov1(t *testing.T) {
	r0_v, r1_v, r2_v, r3_v := 69, 420, 1.23, 1.23
	asm := fmt.Sprintf(`
	mov r0, %d
	mov r1, %d
	mov r2, %f
	mov r3, r2
	`, r0_v, r1_v, r2_v)
	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	expectGpRegisters(t, asm, &mach, ExpMap{
		vm.R0_IDX: uint64(r0_v),
		vm.R1_IDX: uint64(r1_v),
		vm.R2_IDX: math.Float64bits(r2_v),
		vm.R3_IDX: math.Float64bits(r3_v),
	})
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
	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	expectGpRegisters(t, asm, &mach, ExpMap{
		vm.R2_IDX: uint64(r2_v),
		vm.R5_IDX: math.Float64bits(r5_v),
	})
}
func randSign() int {
	return int(math.Ceil(rand.Float64() - 0.5))
}

func TestAddIR2(t *testing.T) {
	r := byte(rand.Int() % vm.GP_REG_MAX)
	reg_v := rand.Uint64()
	reg_add := rand.Uint64()
	asm := fmt.Sprintf(`
	mov r%v, %v
	add UNSIGNED r%v, %v
	`, r, reg_v, r, reg_add)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	expectGpRegisters(t, asm, &mach, ExpMap{
		r: uint64(reg_v+reg_add),
	})
}
func TestAddIR3(t *testing.T) {
	r := byte(rand.Int() % vm.GP_REG_MAX)
	reg_v := int64(rand.Uint64())
	reg_add := int64(rand.Uint64())
	asm := fmt.Sprintf(`
	mov r%v, %v
	add SIGNED r%v, %v
	`, r, reg_v, r, reg_add)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	expectGpRegisters(t, asm, &mach, ExpMap{
		r: uint64(reg_v+reg_add),
	})
}
func TestAddIR4(t *testing.T) {
	r := byte(rand.Int() % vm.GP_REG_MAX)
	reg_v := rand.Float64() * 1000
	reg_add := rand.Float64() * 1000
	asm := fmt.Sprintf(`
	mov r%v, %v
	add FLOAT r%v, %v
	`, r, reg_v, r, reg_add)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	expectGpRegisters(t, asm, &mach, ExpMap{
		r: math.Float64bits(reg_v+reg_add),
	})
}
func TestSubIR1(t *testing.T) {
	r := byte(rand.Int() % vm.GP_REG_MAX)
	reg_v := (rand.Uint64())
	reg_sub := (rand.Uint64())
	asm := fmt.Sprintf(`
	mov r%v, %v
	sub UNSIGNED r%v, %v
	`, r, reg_v, r, reg_sub)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	expectGpRegisters(t, asm, &mach, ExpMap{
		r: uint64(reg_v-reg_sub),
	})
}
func TestSubIR2(t *testing.T) {
	r := byte(rand.Int() % vm.GP_REG_MAX)
	reg_v := int64(rand.Uint64())
	reg_sub := int64(rand.Uint64())
	asm := fmt.Sprintf(`
	mov r%v, %v
	sub SIGNED r%v, %v
	`, r, reg_v, r, reg_sub)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	expectGpRegisters(t, asm, &mach, ExpMap{
		r: uint64(reg_v-reg_sub),
	})
}
func TestSubIR3(t *testing.T) {
	r := byte(rand.Int() % vm.GP_REG_MAX)
	reg_v := rand.Float64()
	reg_sub := rand.Float64()
	asm := fmt.Sprintf(`
	mov r%v, %v
	sub FLOAT r%v, %v
	`, r, reg_v, r, reg_sub)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	expectGpRegisters(t, asm, &mach, ExpMap{
		r: math.Float64bits(reg_v-reg_sub),
	})
}
func TestAddRR1(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	rB := byte(rA + 1%vm.GP_REG_MAX)
	rAV, rBV := rand.Uint64()/1000, rand.Uint64()/1000
	asm := fmt.Sprintf(`
	mov r%v, %v
	mov r%v, %v
	add UNSIGNED r%v, r%v
	`, rA, rAV, rB, rBV, rA, rB)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	expectGpRegisters(t, asm, &mach, ExpMap{
		rA: rAV+rBV,
		rB: rBV,
	})
}
func TestAddRR2(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	rB := byte(rA + 1%vm.GP_REG_MAX)
	rAV, rBV := int64(rand.Uint64()/1000), int64(rand.Uint64()/1000)
	asm := fmt.Sprintf(`
	mov r%v, %v
	mov r%v, %v
	add SIGNED r%v, r%v
	`, rA, rAV, rB, rBV, rA, rB)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	expectGpRegisters(t, asm, &mach, ExpMap{
		rA: uint64(rAV+rBV),
		rB: uint64(rBV),
	})
}
func TestAddRR3(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	rB := byte(rA + 1%vm.GP_REG_MAX)
	rAV, rBV := rand.Float64(), rand.Float64()
	asm := fmt.Sprintf(`
	mov r%v, %v
	mov r%v, %v
	add FLOAT r%v, r%v
	`, rA, rAV, rB, rBV, rA, rB)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	expectGpRegisters(t, asm, &mach, ExpMap{
		rA: math.Float64bits(rAV+rBV),
		rB: math.Float64bits(rBV),
	})
}
func TestSubRR1(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	rB := byte(rA + 1%vm.GP_REG_MAX)
	rAV, rBV := rand.Uint64()/1000, rand.Uint64()/1000
	asm := fmt.Sprintf(`
	mov r%v, %v
	mov r%v, %v
	sub UNSIGNED r%v, r%v
	`, rA, rAV, rB, rBV, rA, rB)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	expectGpRegisters(t, asm, &mach, ExpMap{
		rA: uint64(rAV-rBV),
		rB: uint64(rBV),
	})
}
func TestSubRR2(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	rB := byte(rA + 1%vm.GP_REG_MAX)
	rAV, rBV := int64(rand.Uint64()/1000), int64(rand.Uint64()/1000)
	asm := fmt.Sprintf(`
	mov r%v, %v
	mov r%v, %v
	sub SIGNED r%v, r%v
	`, rA, rAV, rB, rBV, rA, rB)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	expectGpRegisters(t, asm, &mach, ExpMap{
		rA: uint64(rAV-rBV),
		rB: uint64(rBV),
	})
}
func TestSubRR3(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	rB := byte(rA + 1%vm.GP_REG_MAX)
	rAV, rBV := rand.Float64()/1000, rand.Float64()/1000
	asm := fmt.Sprintf(`
	mov r%v, %v
	mov r%v, %v
	sub FLOAT r%v, r%v
	`, rA, rAV, rB, rBV, rA, rB)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	expectGpRegisters(t, asm, &mach, ExpMap{
		rA: math.Float64bits(rAV-rBV),
		rB: math.Float64bits(rBV),
	})
}
func TestDivRR1(t *testing.T) {
	rAV, rBV := rand.Uint64()/1000, rand.Uint64()/1000
	asm := fmt.Sprintf(`
	mov r0, %v
	mov r1, %v
	div UNSIGNED
	`, rAV, rBV)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	expectGpRegisters(t, asm, &mach, ExpMap{
		vm.R0_IDX: rAV,
		vm.R1_IDX: rBV,
		vm.R2_IDX: rAV/rBV,
		vm.R3_IDX: rAV%rBV,
	})
}
func TestDivRR2(t *testing.T) {
	rAV, rBV := int64(rand.Uint64()/1000), int64(rand.Uint64()/1000)
	asm := fmt.Sprintf(`
	mov r0, %v
	mov r1, %v
	div SIGNED
	`, rAV, rBV)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	expectGpRegisters(t, asm, &mach, ExpMap{
		vm.R0_IDX: uint64(rAV),
		vm.R1_IDX: uint64(rBV),
		vm.R2_IDX: uint64(rAV/rBV),
		vm.R3_IDX: uint64(rAV%rBV),
	})
}
func TestDivRR3(t *testing.T) {
	rAV, rBV := rand.Float64()/1000, rand.Float64()/1000
	asm := fmt.Sprintf(`
	mov r0, %v
	mov r1, %v
	div FLOAT
	`, rAV, rBV)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	expectGpRegisters(t, asm, &mach, ExpMap{
		vm.R0_IDX: math.Float64bits(rAV),
		vm.R1_IDX: math.Float64bits(rBV),
		vm.R2_IDX: math.Float64bits(rAV/rBV),
		vm.R3_IDX: 0,
	})
}
func TestMulRR1(t *testing.T) {
	rAV, rBV := rand.Uint64()/1000, rand.Uint64()/1000
	asm := fmt.Sprintf(`
	mov r0, %v
	mov r1, %v
	mul UNSIGNED
	`, rAV, rBV)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	expectGpRegisters(t, asm, &mach, ExpMap{
		vm.R0_IDX: rAV,
		vm.R1_IDX: rBV,
		vm.R2_IDX: rAV*rBV,
	})
}
func TestMulRR2(t *testing.T) {
	rAV, rBV := rand.Int()/1000, rand.Int()/1000
	asm := fmt.Sprintf(`
	mov r0, %v
	mov r1, %v
	mul SIGNED
	`, rAV, rBV)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	expectGpRegisters(t, asm, &mach, ExpMap{
		vm.R0_IDX: uint64(rAV),
		vm.R1_IDX: uint64(rBV),
		vm.R2_IDX: uint64(rAV*rBV),
	})
}
func TestMulRR3(t *testing.T) {
	rAV, rBV := rand.Float64()/1000, rand.Float64()/1000
	asm := fmt.Sprintf(`
	mov r0, %v
	mov r1, %v
	mul FLOAT
	`, rAV, rBV)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	expectGpRegisters(t, asm, &mach, ExpMap{
		vm.R0_IDX: math.Float64bits(rAV),
		vm.R1_IDX: math.Float64bits(rBV),
		vm.R2_IDX: math.Float64bits(rAV*rBV),
	})
}
func TestIncAndDec1(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	rAV := uint64(rand.Float64() * 1000)
	decrT, incrT := uint64(rand.Float64()*10), uint64(rand.Float64()*10)
	asm := fmt.Sprintf(`
	mov r%v, %v
	`, rA, rAV)
	for range incrT {
		asm = fmt.Sprintf("%s\ninc r%v", asm, rA)
	}
	for range decrT {
		asm = fmt.Sprintf("%s\ndec r%v", asm, rA)
	}
	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	expectGpRegisters(t, asm, &mach, ExpMap{
		rA: rAV + incrT - decrT,
	})

}
