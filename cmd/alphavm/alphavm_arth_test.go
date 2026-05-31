package alphavm

import (
	"fmt"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
	"math"
	"math/rand"
	"testing"
)

func TestAddIR1(t *testing.T) {
	r0_v, r1_v := 69, -420
	r2_v := r0_v + r1_v
	r3_v := 80.085
	r4_v := -133.7
	r5_v := r3_v + r4_v
	asm := fmt.Sprintf(`
	section '.code'
	@entry
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
	if err = expectGpRegisters(asm, &mach, ExpMap{
		vm.R2_IDX: vm.RegisterWithValue(uint64(r2_v)),
		vm.R5_IDX: vm.RegisterWithValue(math.Float64bits(r5_v)),
	}); err != nil {
		t.Error(err.Error())
	}
}
func TestAddIR2(t *testing.T) {
	r := byte(rand.Int() % vm.GP_REG_MAX)
	reg_v := rand.Uint64()
	reg_add := rand.Uint64()
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov r%v, %v
		add UNSIGNED r%v, %v
	`, r, reg_v, r, reg_add)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	if err = expectGpRegisters(asm, &mach, ExpMap{
		r: vm.RegisterWithValue(uint64(reg_v + reg_add)),
	}); err != nil {
		t.Error(err.Error())
	}
}
func TestAddIR3(t *testing.T) {
	r := byte(rand.Int() % vm.GP_REG_MAX)
	reg_v := int64(rand.Uint64())
	reg_add := int64(rand.Uint64())
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov r%v, %v
		add SIGNED r%v, %v
	`, r, reg_v, r, reg_add)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		r: vm.RegisterWithValue(uint64(reg_v + reg_add)),
	}); err != nil {
		t.Error(err.Error())
	}
}
func TestAddIR4(t *testing.T) {
	r := byte(rand.Int() % vm.GP_REG_MAX)
	reg_v := rand.Float64() * 1000
	reg_add := rand.Float64() * 1000
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov r%v, %v
		add FLOAT r%v, %v
	`, r, reg_v, r, reg_add)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		r: vm.RegisterWithValue(math.Float64bits(reg_v + reg_add)),
	}); err != nil {
		t.Error(err.Error())
	}
}
func TestSubIR1(t *testing.T) {
	r := byte(rand.Int() % vm.GP_REG_MAX)
	reg_v := (rand.Uint64())
	reg_sub := (rand.Uint64())
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov r%v, %v
		sub UNSIGNED r%v, %v
	`, r, reg_v, r, reg_sub)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		r: vm.RegisterWithValue(uint64(reg_v - reg_sub)),
	}); err != nil {
		t.Error(err.Error())
	}
}
func TestSubIR2(t *testing.T) {
	r := byte(rand.Int() % vm.GP_REG_MAX)
	reg_v := int64(rand.Uint64())
	reg_sub := int64(rand.Uint64())
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov r%v, %v
		sub SIGNED r%v, %v
	`, r, reg_v, r, reg_sub)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		r: vm.RegisterWithValue(uint64(reg_v - reg_sub)),
	}); err != nil {
		t.Error(err.Error())
	}
}
func TestSubIR3(t *testing.T) {
	r := byte(rand.Int() % vm.GP_REG_MAX)
	reg_v := rand.Float64()
	reg_sub := rand.Float64()
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov r%v, %v
		sub FLOAT r%v, %v
	`, r, reg_v, r, reg_sub)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		r: vm.RegisterWithValue(math.Float64bits(reg_v - reg_sub)),
	}); err != nil {
		t.Error(err.Error())
	}
}
func TestAddRR1(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	rB := byte(rA + 1%vm.GP_REG_MAX)
	rAV, rBV := rand.Uint64()/1000, rand.Uint64()/1000
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov r%v, %v
		mov r%v, %v
		add UNSIGNED r%v, r%v
	`, rA, rAV, rB, rBV, rA, rB)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		rA: vm.RegisterWithValue(rAV + rBV),
		rB: vm.RegisterWithValue(rBV),
	}); err != nil {
		t.Error(err.Error())
	}
}
func TestAddRR2(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	rB := byte(rA + 1%vm.GP_REG_MAX)
	rAV, rBV := int64(rand.Uint64()/1000), int64(rand.Uint64()/1000)
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov r%v, %v
		mov r%v, %v
		add SIGNED r%v, r%v
	`, rA, rAV, rB, rBV, rA, rB)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		rA: vm.RegisterWithValue(uint64(rAV + rBV)),
		rB: vm.RegisterWithValue(uint64(rBV)),
	}); err != nil {
		t.Error(err.Error())
	}
}
func TestAddRR3(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	rB := byte(rA + 1%vm.GP_REG_MAX)
	rAV, rBV := rand.Float64(), rand.Float64()
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov r%v, %v
		mov r%v, %v
		add FLOAT r%v, r%v
	`, rA, rAV, rB, rBV, rA, rB)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		rA: vm.RegisterWithValue(math.Float64bits(rAV + rBV)),
		rB: vm.RegisterWithValue(math.Float64bits(rBV)),
	}); err != nil {
		t.Error(err.Error())
	}
}
func TestSubRR1(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	rB := byte(rA + 1%vm.GP_REG_MAX)
	rAV, rBV := rand.Uint64()/1000, rand.Uint64()/1000
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov r%v, %v
		mov r%v, %v
		sub UNSIGNED r%v, r%v
	`, rA, rAV, rB, rBV, rA, rB)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		rA: vm.RegisterWithValue(uint64(rAV - rBV)),
		rB: vm.RegisterWithValue(uint64(rBV)),
	}); err != nil {
		t.Error(err.Error())
	}
}
func TestSubRR2(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	rB := byte(rA + 1%vm.GP_REG_MAX)
	rAV, rBV := int64(rand.Uint64()/1000), int64(rand.Uint64()/1000)
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov r%v, %v
		mov r%v, %v
		sub SIGNED r%v, r%v
	`, rA, rAV, rB, rBV, rA, rB)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		rA: vm.RegisterWithValue(uint64(rAV - rBV)),
		rB: vm.RegisterWithValue(uint64(rBV)),
	}); err != nil {
		t.Error(err.Error())
	}
}
func TestSubRR3(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	rB := byte(rA + 1%vm.GP_REG_MAX)
	rAV, rBV := rand.Float64()/1000, rand.Float64()/1000
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov r%v, %v
		mov r%v, %v
		sub FLOAT r%v, r%v
	`, rA, rAV, rB, rBV, rA, rB)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		rA: vm.RegisterWithValue(math.Float64bits(rAV - rBV)),
		rB: vm.RegisterWithValue(math.Float64bits(rBV)),
	}); err != nil {
		t.Error(err.Error())
	}
}
func TestDivRR1(t *testing.T) {
	rAV, rBV := rand.Uint64()/1000, rand.Uint64()/1000
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov r0, %v
		mov r1, %v
		div UNSIGNED WORD
	`, rAV, rBV)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R0_IDX: vm.RegisterWithValue(rAV),
		vm.R1_IDX: vm.RegisterWithValue(rBV),
		vm.R2_IDX: vm.RegisterWithValue(rAV / rBV),
		vm.R3_IDX: vm.RegisterWithValue(rAV % rBV),
	}); err != nil {
		t.Error(err)
	}
}
func TestDivRR2(t *testing.T) {
	rAV, rBV := int64(rand.Uint64()/1000), int64(rand.Uint64()/1000)
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov r0, %v
		mov r1, %v
		div SIGNED WORD
	`, rAV, rBV)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R0_IDX: vm.RegisterWithValue(uint64(rAV)),
		vm.R1_IDX: vm.RegisterWithValue(uint64(rBV)),
		vm.R2_IDX: vm.RegisterWithValue(uint64(rAV / rBV)),
		vm.R3_IDX: vm.RegisterWithValue(uint64(rAV % rBV)),
	}); err != nil {
		t.Error(err)
	}
}
func TestDivRR3(t *testing.T) {
	rAV, rBV := rand.Float64()/1000, rand.Float64()/1000
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov r0, %v
		mov r1, %v
		div FLOAT WORD
	`, rAV, rBV)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R0_IDX: vm.RegisterWithValue(math.Float64bits(rAV)),
		vm.R1_IDX: vm.RegisterWithValue(math.Float64bits(rBV)),
		vm.R2_IDX: vm.RegisterWithValue(math.Float64bits(rAV / rBV)),
		vm.R3_IDX: vm.RegisterWithValue(0),
	}); err != nil {
		t.Error(err)
	}
}
func TestMulRR1(t *testing.T) {
	rAV, rBV := rand.Uint64()/1000, rand.Uint64()/1000
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov r0, %v
		mov r1, %v
		mul UNSIGNED WORD
	`, rAV, rBV)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R0_IDX: vm.RegisterWithValue(rAV),
		vm.R1_IDX: vm.RegisterWithValue(rBV),
		vm.R2_IDX: vm.RegisterWithValue(rAV * rBV),
	}); err != nil {
		t.Error(err)
	}
}
func TestMulRR2(t *testing.T) {
	rAV, rBV := rand.Int()/1000, rand.Int()/1000
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov r0, %v
		mov r1, %v
		mul SIGNED WORD
	`, rAV, rBV)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R0_IDX: vm.RegisterWithValue(uint64(rAV)),
		vm.R1_IDX: vm.RegisterWithValue(uint64(rBV)),
		vm.R2_IDX: vm.RegisterWithValue(uint64(rAV * rBV)),
	}); err != nil {
		t.Error(err)
	}
}
func TestMulRR3(t *testing.T) {
	rAV, rBV := rand.Float64()/1000, rand.Float64()/1000
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov r0, %v
		mov r1, %v
		mul FLOAT WORD
	`, rAV, rBV)

	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R0_IDX: vm.RegisterWithValue(math.Float64bits(rAV)),
		vm.R1_IDX: vm.RegisterWithValue(math.Float64bits(rBV)),
		vm.R2_IDX: vm.RegisterWithValue(math.Float64bits(rAV * rBV)),
	}); err != nil {
		t.Error(err)
	}
}
func TestIncAndDec1(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	rAV := uint64(rand.Float64() * 1000)
	decrT, incrT := uint64(rand.Float64()*10), uint64(rand.Float64()*10)
	asm := fmt.Sprintf(`
	section '.code'
	@entry
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
	if err := expectGpRegisters(asm, &mach, ExpMap{
		rA: vm.RegisterWithValue(rAV + incrT - decrT),
	}); err != nil {
		t.Error(err.Error())
	}

}
