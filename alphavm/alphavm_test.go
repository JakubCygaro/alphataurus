package alphavm

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strings"
	"testing"

	"github.com/JakubCygaro/alphataurus/assembler"
	"github.com/JakubCygaro/alphataurus/internal/vm"
)

const (
	DEFAULT_STACK_SIZE = 16
)

func execute(code []byte) (vm.VmState, error) {
	return executeStackSize(code, DEFAULT_STACK_SIZE)
}
func executeStackSize(code []byte, stacksz uint64) (vm.VmState, error) {
	mach := vm.CreateVmState(stacksz)
	if err := mach.Execute(code); err != nil {
		return mach, err
	}
	return mach, nil
}
func assemble(source string) ([]byte, error) {
	asmblr := assembler.NewAssembler(*bufio.NewReader(strings.NewReader(source)))
	code, _, err := asmblr.EmitBytecode()
	return code, err
}
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

func expectGpRegisters(asm string, mach *vm.VmState, regStates ExpMap) error {
	var msg string
	err := false
	for k, v := range regStates {
		if r, _ := mach.GetGpRXAsUint64(k); r != v {
			msg = strings.Join([]string{
				fmt.Sprintf("State of general purpose register r%v was different from expected", k),
				fmt.Sprintf("\t[uint64]  expected (%v) \t got (%v)", v, r),
				fmt.Sprintf("\t[int64]   expected (%v) \t got (%v)", int64(v), int64(r)),
				fmt.Sprintf("\t[float64] expected (%v) \t got (%v)", math.Float64frombits(v), math.Float64frombits(r)),
			}, "\n")
			err = true
		}
	}
	if err {
		msg = strings.Join([]string{
			fmt.Sprintf("Compilation of %s", asm),
			msg,
		}, "\n")
		return fmt.Errorf(msg)
	}
	return nil
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
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R0_IDX: uint64(r0_v),
		vm.R1_IDX: uint64(r1_v),
		vm.R2_IDX: math.Float64bits(r2_v),
		vm.R3_IDX: math.Float64bits(r3_v),
	}); err != nil {
		t.Errorf(err.Error())
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
	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	if err = expectGpRegisters(asm, &mach, ExpMap{
		vm.R2_IDX: uint64(r2_v),
		vm.R5_IDX: math.Float64bits(r5_v),
	}); err != nil {
		t.Errorf(err.Error())
	}
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
	if err = expectGpRegisters(asm, &mach, ExpMap{
		r: uint64(reg_v + reg_add),
	}); err != nil {
		t.Errorf(err.Error())
	}
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
	if err := expectGpRegisters(asm, &mach, ExpMap{
		r: uint64(reg_v + reg_add),
	}); err != nil {
		t.Errorf(err.Error())
	}
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
	if err := expectGpRegisters(asm, &mach, ExpMap{
		r: math.Float64bits(reg_v + reg_add),
	}); err != nil {
		t.Errorf(err.Error())
	}
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
	if err := expectGpRegisters(asm, &mach, ExpMap{
		r: uint64(reg_v - reg_sub),
	}); err != nil {
		t.Errorf(err.Error())
	}
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
	if err := expectGpRegisters(asm, &mach, ExpMap{
		r: uint64(reg_v - reg_sub),
	}); err != nil {
		t.Errorf(err.Error())
	}
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
	if err := expectGpRegisters(asm, &mach, ExpMap{
		r: math.Float64bits(reg_v - reg_sub),
	}); err != nil {
		t.Errorf(err.Error())
	}
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
	if err := expectGpRegisters(asm, &mach, ExpMap{
		rA: rAV + rBV,
		rB: rBV,
	}); err != nil {
		t.Errorf(err.Error())
	}
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
	if err := expectGpRegisters(asm, &mach, ExpMap{
		rA: uint64(rAV + rBV),
		rB: uint64(rBV),
	}); err != nil {
		t.Errorf(err.Error())
	}
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
	if err := expectGpRegisters(asm, &mach, ExpMap{
		rA: math.Float64bits(rAV + rBV),
		rB: math.Float64bits(rBV),
	}); err != nil {
		t.Errorf(err.Error())
	}
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
	if err := expectGpRegisters(asm, &mach, ExpMap{
		rA: uint64(rAV - rBV),
		rB: uint64(rBV),
	}); err != nil {
		t.Errorf(err.Error())
	}
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
	if err := expectGpRegisters(asm, &mach, ExpMap{
		rA: uint64(rAV - rBV),
		rB: uint64(rBV),
	}); err != nil {
		t.Errorf(err.Error())
	}
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
	if err := expectGpRegisters(asm, &mach, ExpMap{
		rA: math.Float64bits(rAV - rBV),
		rB: math.Float64bits(rBV),
	}); err != nil {
		t.Errorf(err.Error())
	}
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
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R0_IDX: rAV,
		vm.R1_IDX: rBV,
		vm.R2_IDX: rAV / rBV,
		vm.R3_IDX: rAV % rBV,
	}); err != nil {
		t.Errorf(err.Error())
	}
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
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R0_IDX: uint64(rAV),
		vm.R1_IDX: uint64(rBV),
		vm.R2_IDX: uint64(rAV / rBV),
		vm.R3_IDX: uint64(rAV % rBV),
	}); err != nil {
		t.Errorf(err.Error())
	}
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
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R0_IDX: math.Float64bits(rAV),
		vm.R1_IDX: math.Float64bits(rBV),
		vm.R2_IDX: math.Float64bits(rAV / rBV),
		vm.R3_IDX: 0,
	}); err != nil {
		t.Errorf(err.Error())
	}
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
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R0_IDX: rAV,
		vm.R1_IDX: rBV,
		vm.R2_IDX: rAV * rBV,
	}); err != nil {
		t.Errorf(err.Error())
	}
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
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R0_IDX: uint64(rAV),
		vm.R1_IDX: uint64(rBV),
		vm.R2_IDX: uint64(rAV * rBV),
	}); err != nil {
		t.Errorf(err.Error())
	}
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
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R0_IDX: math.Float64bits(rAV),
		vm.R1_IDX: math.Float64bits(rBV),
		vm.R2_IDX: math.Float64bits(rAV * rBV),
	}); err != nil {
		t.Errorf(err.Error())
	}
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
	if err := expectGpRegisters(asm, &mach, ExpMap{
		rA: rAV + incrT - decrT,
	}); err != nil {
		t.Errorf(err.Error())
	}

}
func TestCmp1(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	rAV := uint64(rand.Float64() * 1000)
	rBV := rAV + 1
	asm := fmt.Sprintf(`
		mov r%v, %v
		cmp r%v, %v
		mov r0, 1
		cmp r0, 3
	`, rA, rAV, rA, rBV)
	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	flags := mach.GetFlags()
	if !flags.Sf {
		t.Errorf("Sign flag was not set")
		t.Errorf("%+v", flags)
	}
}
func TestCmp2(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	rAV := uint64(rand.Float64() * 1000)
	rBV := rAV + 1
	asm := fmt.Sprintf(`
		mov r%v, %v
		cmp FLOAT r%v, %v
		mov r0, 1
		cmp FLOAT r0, 3
	`, rA, rAV, rA, rBV)
	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	flags := mach.GetFlags()
	if !flags.Sf {
		t.Errorf("Sign flag was not set")
		t.Errorf("%+v", flags)
	}
}
func TestJmpE1(t *testing.T) {
	asm := `
	ENTRY:
		mov r0, 10
		mov r1, 0
		jmp START
	ZERO:
		mov r4, 420
		jmp END
	START:
		inc r1
		dec r0
		cmp r0, 0
		je ZERO
		jg START
	END:
		mov r5, 1337
	`
	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R0_IDX: 0,
		vm.R1_IDX: 10,
		vm.R4_IDX: 420,
		vm.R5_IDX: 1337,
	}); err != nil {
		t.Errorf(err.Error())
	}
}
func TestJmpG2(t *testing.T) {
	asm := `
		mov r0, 10
		mov r1, 0
	L0:
		inc r1
		dec r0
		cmp r0, 0
		jg L0
	`
	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R1_IDX: 10,
		vm.R0_IDX: 0,
	}); err != nil {
		t.Errorf(err.Error())
	}
}
func TestJmpG1(t *testing.T) {
	asm := `
		mov r0, 10
		mov r1, 0
		inc r1
		dec r0
		cmp r0, 0
		jg 0x100
	`
	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R1_IDX: 10,
		vm.R0_IDX: 0,
	}); err != nil {
		t.Errorf(err.Error())
	}
}
func TestExpressions1(t *testing.T) {
	for range 100 {
		startingVal := rand.Intn(100)
		expr := fmt.Sprintf("%v", startingVal)
		endVal := startingVal
		for range rand.Intn(10) {
			op := rand.Intn(vm.OP_TDIV + 1)
			arg := rand.Intn(100)
			res := endVal
			var opCh rune
			switch op {
			case vm.OP_TADD:
				opCh = '+'
				res += arg
			case vm.OP_TSUB:
				opCh = '-'
				res -= arg
			case vm.OP_TDIV:
				if arg == 0 {
					arg = 1
				}
				opCh = '/'
				res /= arg
			case vm.OP_TMUL:
				opCh = '*'
				res *= arg
			}
			expr = fmt.Sprintf("(%v %c %v)", endVal, opCh, arg)
			endVal = res
		}
		rA := byte(rand.Int() % vm.GP_REG_MAX)
		asm := fmt.Sprintf(`
			mov r%v, %v
		`, rA, expr)
		mach, err := assembleAndExecute(asm)
		if err != nil {
			t.Error(err)
			t.Errorf("Compilation of:\n %s", asm)
			t.FailNow()
		}
		if err := expectGpRegisters(asm, &mach, ExpMap{
			rA: uint64(endVal),
		}); err != nil {
			t.Errorf(err.Error())
			break
		}
	}
}

func TestExpressions2(t *testing.T) {
	for range 100 {
		startingVal := rand.Float64()
		expr := fmt.Sprintf("%v", startingVal)
		endVal := startingVal
		for range rand.Intn(10) {
			op := rand.Intn(vm.OP_TDIV + 1)
			arg := rand.Float64()
			res := endVal
			var opCh rune
			switch op {
			case vm.OP_TADD:
				opCh = '+'
				res += arg
			case vm.OP_TSUB:
				opCh = '-'
				res -= arg
			case vm.OP_TDIV:
				if arg == 0 {
					arg = 1
				}
				opCh = '/'
				res /= arg
			case vm.OP_TMUL:
				opCh = '*'
				res *= arg
			}
			expr = fmt.Sprintf("(%v %c %v)", endVal, opCh, arg)
			endVal = res
		}
		rA := byte(rand.Int() % vm.GP_REG_MAX)
		asm := fmt.Sprintf(`
			mov r%v, %v
		`, rA, expr)
		mach, err := assembleAndExecute(asm)
		if err != nil {
			t.Error(err)
			t.Errorf("Compilation of:\n %s", asm)
			t.FailNow()
		}
		if err := expectGpRegisters(asm, &mach, ExpMap{
			rA: uint64(math.Float64bits(endVal)),
		}); err != nil {
			t.Errorf(err.Error())
			break
		}
	}
}

func TestExpressions1F(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	asm := fmt.Sprintf(`
			mov r%v, ( 0 / 0 )
		`, rA)
	_, err := assemble(asm)
	if err == nil {
		t.Errorf("Expected assembling failure")
		t.Errorf("Compilation of:\n %s", asm)
	}
}

func TestExpressions2F(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	asm := fmt.Sprintf(`
			mov r%v, [[0]]
		`, rA)
	_, err := assemble(asm)
	if err == nil {
		t.Errorf("Expected assembling failure")
		t.Errorf("Compilation of:\n %s", asm)
	}
}
func TestDeref1F(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	asm := fmt.Sprintf(`
			mov r%v, [0x0]
		`, rA)
	b, err := assemble(asm)
	if err != nil {
		t.Error(err)
		t.Errorf("Compilation of:\n %s", asm)
	}
	if _, err := execute(b); err == nil {
		t.Errorf("Expected execution failure")
		t.Errorf("Compilation of:\n %s", asm)
	} else if ok, err := regexp.MatchString("Segmentation fault", err.Error()); !ok || err != nil {
		t.Errorf("Expected segmentation fault")
		t.Error(err)
	}
}
func TestStack1(t *testing.T) {
	var asm string
	lines := make([]string, 0, DEFAULT_STACK_SIZE)
	for range DEFAULT_STACK_SIZE{
		val := rand.Intn(10000) - 5000
		if rand.Intn(100) < 50 {
			lines = append(lines, fmt.Sprintf("push %v", val))
		} else {
			rA := byte(rand.Int() % vm.GP_REG_MAX)
			lines = append(lines, fmt.Sprintf("mov r%v, %v", rA, val))
			lines = append(lines, fmt.Sprintf("push r%v", rA))
		}
	}
	for range DEFAULT_STACK_SIZE{
		rA := byte(rand.Int() % vm.GP_REG_MAX)
		lines = append(lines, fmt.Sprintf("pop r%v", rA))
	}
	asm = strings.Join(lines, "\n")
	if _, err := assembleAndExecute(asm); err != nil {
		t.Error(err)
		t.Errorf("Compilation of:\n %s", asm)
	}
}
func TestStack1F(t *testing.T) {
	var asm string
	lines := make([]string, 0, DEFAULT_STACK_SIZE)
	for range DEFAULT_STACK_SIZE + 1 {
		lines = append(lines, "push 1")
	}
	asm = strings.Join(lines, "\n")
	b, err := assemble(asm)
	if err != nil {
		t.Error(err)
		t.Errorf("Compilation of:\n %s", asm)
	}
	if _, err := execute(b); err == nil {
		t.Errorf("Expected execution failure")
		t.Errorf("Compilation of:\n%s", asm)
	} else if ok, err := regexp.MatchString("Stack overflow", err.Error()); !ok || err != nil {
		t.Errorf("Expected stack overflow")
		t.Error(err)
	}
}
func TestStack2F(t *testing.T) {
	asm := `
		pop
	`
	b, err := assemble(asm)
	if err != nil {
		t.Error(err)
		t.Errorf("Compilation of:\n %s", asm)
	}
	if _, err := execute(b); err == nil {
		t.Errorf("Expected execution failure")
		t.Errorf("Compilation of:\n%s", asm)
	} else if ok, err := regexp.MatchString("Stack underflow", err.Error()); !ok || err != nil {
		t.Errorf("Expected stack underflow")
		t.Error(err)
	}
}
