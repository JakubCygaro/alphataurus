package alphavm

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strings"
	"testing"

	"github.com/JakubCygaro/alphataurus/pkg/assembler"
	"github.com/JakubCygaro/alphataurus/pkg/linker"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

const (
	DEFAULT_STACK_SIZE = 16
)

func execute(elf vm.AlphaELFFile) (vm.VmState, error) {
	return executeStackSize(elf, DEFAULT_STACK_SIZE)
}
func executeStackSize(elf vm.AlphaELFFile, stacksz uint64) (vm.VmState, error) {
	mach := vm.CreateVmState(stacksz)
	if err := mach.Execute(elf); err != nil {
		return mach, err
	}
	return mach, nil
}
func assemble(source string) ([]byte, error) {
	asmblr := assembler.NewAssembler(*bufio.NewReader(strings.NewReader(source)))
	code, err := asmblr.Assemble()
	return code, err
}
func assembleAndLink(source string) (vm.AlphaELFFile, error) {
	code, err := assemble(source)
	if err != nil {
		return vm.AlphaELFFile{}, err
	}
	ld := linker.NewLinker()
	linked, err := ld.Link([]linker.LinkerInput{linker.Bytes(code)})
	return linked, err
}
func assembleAndExecute(source string) (vm.VmState, error) {
	elf, err := assembleAndLink(source)
	mach := vm.CreateVmState(16)
	if err != nil {
		return mach, err
	}
	err = mach.Execute(elf)
	if err != nil {
		return mach, err
	}
	return mach, nil
}

type ExpMap map[byte]vm.Register

func expectGpRegisters(asm string, mach *vm.VmState, regStates ExpMap) error {
	var msg string
	err := false
	for k, tr := range regStates {
		var tro any
		tr.GetValAs(vm.TY_UINT, vm.SZ_64, &tro)
		trv := tro.(uint64)
		if r, _ := mach.GetGpRXAsUint64(k); r != trv {
			msg = strings.Join([]string{
				fmt.Sprintf("State of general purpose register r%v was different from expected", k),
				fmt.Sprintf("\t[uint64]  expected (%v) \t got (%v)", trv, r),
				fmt.Sprintf("\t[int64]   expected (%v) \t got (%v)", int64(trv), int64(r)),
				fmt.Sprintf("\t[float64] expected (%v) \t got (%v)", math.Float64frombits(trv), math.Float64frombits(r)),
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

type testingStack vm.VmStack

func makeTestingStack(size int) testingStack {
	return make(testingStack, size)
}

func (s *testingStack) push(value any) {
	if u8, ok := value.(uint8); ok {
		*s = append(*s, u8)
	} else if u16, ok := value.(uint16); ok {
		binary.BigEndian.AppendUint16(*s, u16)
	} else if u32, ok := value.(uint32); ok {
		binary.BigEndian.AppendUint32(*s, u32)
	} else if u64, ok := value.(uint64); ok {
		binary.BigEndian.AppendUint64(*s, u64)
	}
}

func expectStack(mach *vm.VmState, stack vm.VmStack) error {
	lines := make([]string, 0)
	vmStack := mach.GetStack()
	if len(vmStack) != len(stack) {
		lines = append(lines,
			fmt.Sprintf("Input stack and vm stack sizes do not match\nlen(vm) = %v\tlen(input) = %v",
				len(vmStack), len(stack)))
		return fmt.Errorf("%s", strings.Join(lines, "\n"))
	}
	for i, v := range stack {
		if vmStack[i] != v {
			lines = append(lines, fmt.Sprintf("Stack value at [%v] was different from expected", i))
			// lines = append(lines,
			// 	fmt.Sprintf("\t[uint64]  expected (%v) \t got (%v)",
			// 		v, vmStack[i]),
			// 	fmt.Sprintf("\t[int64]   expected (%v) \t got (%v)",
			// 		int64(v), int64(vmStack[i])),
			// 	fmt.Sprintf("\t[float64] expected (%v) \t got (%v)",
			// 		math.Float64frombits(v), math.Float64frombits(vmStack[i])),
			// )
			lines = append(lines,
				fmt.Sprintf("got (%v) expected (%v)", vmStack[i], v),
			)
			lines = append(lines,
				fmt.Sprintf("got (0x%x) expected (0x%x)", vmStack[i], v),
			)
		}
	}
	if len(lines) > 0 {
		return fmt.Errorf("%s", strings.Join(lines, "\n"))
	}
	return nil
}

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
		t.Errorf(err.Error())
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
		t.Errorf(err.Error())
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
		t.Errorf(err.Error())
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
		t.Errorf(err.Error())
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
		t.Errorf(err.Error())
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
		t.Errorf(err.Error())
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
		t.Errorf(err.Error())
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
		t.Errorf(err.Error())
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
		t.Errorf(err.Error())
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
		t.Errorf(err.Error())
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
		t.Errorf(err.Error())
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
		t.Errorf(err.Error())
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
		t.Errorf(err.Error())
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
		t.Errorf(err.Error())
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
		t.Errorf(err.Error())
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
		t.Errorf(err.Error())
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
		t.Errorf(err.Error())
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
		t.Errorf(err.Error())
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
		t.Errorf(err.Error())
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
		t.Errorf(err.Error())
	}

}
func TestCmp1(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	rAV := uint64(rand.Float64() * 1000)
	rBV := rAV + 1
	asm := fmt.Sprintf(`
	section '.code'
	@entry
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
	rAV := rand.Float64() * 1000
	rBV := rAV + 1
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov r%v, %v
		cmp FLOAT r%v, %v
	`, rA, rAV, rA, rBV)
	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	flags := mach.GetFlags()
	if !flags.Sf {
		t.Errorf("Compilation of:\n%s\n", asm)
		t.Errorf("Sign flag was not set")
		t.Errorf("%+v", flags)
	}
}
func TestJmpE1(t *testing.T) {
	asm := `
	section '.code'
	@entry
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
		vm.R0_IDX: vm.RegisterWithValue(0),
		vm.R1_IDX: vm.RegisterWithValue(10),
		vm.R4_IDX: vm.RegisterWithValue(420),
		vm.R5_IDX: vm.RegisterWithValue(1337),
	}); err != nil {
		t.Errorf(err.Error())
	}
}
func TestJmpG2(t *testing.T) {
	asm := `
	section '.code'
	@entry
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
		vm.R1_IDX: vm.RegisterWithValue(10),
		vm.R0_IDX: vm.RegisterWithValue(0),
	}); err != nil {
		t.Errorf(err.Error())
	}
}
func TestJmpG1(t *testing.T) {
	asm := `
	section '.code'
	@entry
		mov r0, 10
		mov r1, 0
		inc r1
		dec r0
		cmp r0, 0
		jg 0x1001
	`
	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R1_IDX: vm.RegisterWithValue(10),
		vm.R0_IDX: vm.RegisterWithValue(0),
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
		section '.code'
		@entry
			mov r%v, %v
		`, rA, expr)
		mach, err := assembleAndExecute(asm)
		if err != nil {
			t.Error(err)
			t.Errorf("Compilation of:\n %s", asm)
			t.FailNow()
		}
		if err := expectGpRegisters(asm, &mach, ExpMap{
			rA: vm.RegisterWithValue(uint64(endVal)),
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
		section '.code'
		@entry
			mov r%v, %v
		`, rA, expr)
		mach, err := assembleAndExecute(asm)
		if err != nil {
			t.Error(err)
			t.Errorf("Compilation of:\n %s", asm)
			t.FailNow()
		}
		if err := expectGpRegisters(asm, &mach, ExpMap{
			rA: vm.RegisterWithValue(uint64(math.Float64bits(endVal))),
		}); err != nil {
			t.Errorf(err.Error())
			break
		}
	}
}

func TestExpressions1F(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	asm := fmt.Sprintf(`
		section '.code'
		@entry
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
		section '.code'
		@entry
			mov r%v, [[0]]
		`, rA)
	_, err := assemble(asm)
	if err == nil {
		t.Errorf("Expected assembling failure")
		t.Errorf("Compilation of:\n %s", asm)
	}
}
func TestDeref1(t *testing.T) {
	stackSize := rand.Intn(32-5) + 5
	stack := makeTestingStack(stackSize)
	for range stackSize {
		stack.push(uint64(rand.Intn(101) - 50))
	}
	lines := make([]string, 0)
	lines = append(lines, `
	section '.code'
	@entry
	`)
	for i, v := range stack {
		lines = append(lines, fmt.Sprintf("mov [bp+%v], %v", i+1, int64(v)))
	}
	asm := strings.Join(lines, "\n")
	b, err := assembleAndLink(asm)
	if err != nil {
		t.Error(err)
		t.Errorf("Compilation of:\n%s", asm)
	}
	if mach, err := executeStackSize(b, uint64(len(stack))); err != nil {
		t.Error(err)
		t.Errorf("Compilation of:\n%s", asm)
	} else if err := expectStack(&mach, vm.VmStack(stack)); err != nil {
		t.Error(err.Error())
		t.Errorf("Compilation of:\n%s", asm)
	}
}
func TestDeref2(t *testing.T) {
	stackSize := rand.Intn(32-5) + 5
	stack := makeTestingStack(stackSize)
	for range stackSize {
		stack.push(uint64(rand.Intn(101) - 50))
	}
	lines := make([]string, 0)
	lines = append(lines, `
	section '.code'
	@entry
	`)
	for i, v := range stack {
		rA := byte(rand.Int() % vm.GP_REG_MAX)
		lines = append(lines, fmt.Sprintf("mov r%v, %v", rA, int64(v)))
		lines = append(lines, fmt.Sprintf("mov [bp+%v], %v", i+1, int64(v)))
	}
	asm := strings.Join(lines, "\n")
	b, err := assembleAndLink(asm)
	if err != nil {
		t.Error(err)
		t.Errorf("Compilation of:\n%s", asm)
	}
	if mach, err := executeStackSize(b, uint64(len(stack))); err != nil {
		t.Error(err)
		t.Errorf("Compilation of:\n%s", asm)
	} else if err := expectStack(&mach, vm.VmStack(stack)); err != nil {
		t.Error(err.Error())
		t.Errorf("Compilation of:\n%s", asm)
	}
}
func TestDeref3(t *testing.T) {
	stackSize := rand.Intn(32-5) + 5
	stack := makeTestingStack(stackSize)
	for range stackSize {
		stack.push(uint64(rand.Intn(101) - 50))
	}
	lines := make([]string, 0)
	lines = append(lines, `
	section '.code'
	@entry
	`)
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	lines = append(lines, fmt.Sprintf("mov r%v, 1", rA))
	for _, v := range stack {
		lines = append(lines, fmt.Sprintf("mov [bp+r%v], %v", rA, int64(v)))
		lines = append(lines, fmt.Sprintf("inc r%v", rA))
	}
	asm := strings.Join(lines, "\n")
	b, err := assembleAndLink(asm)
	if err != nil {
		t.Error(err)
		t.Errorf("Compilation of:\n%s", asm)
	}
	if mach, err := executeStackSize(b, uint64(len(stack))); err != nil {
		t.Error(err)
		t.Errorf("Compilation of:\n%s", asm)
	} else if err := expectStack(&mach, vm.VmStack(stack)); err != nil {
		t.Error(err.Error())
		t.Errorf("Compilation of:\n%s", asm)
	}
}
func TestDeref1F(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	asm := fmt.Sprintf(`
		section '.code'
		@entry
			mov r%v, [0x0]
		`, rA)
	b, err := assembleAndLink(asm)
	if err != nil {
		t.Error(err)
		t.Errorf("Compilation of:\n %s", asm)
	}
	if _, err := execute(b); err == nil {
		t.Errorf("Expected execution failure")
		t.Errorf("Compilation of:\n%s", asm)
	} else if ok, err := regexp.MatchString("Segmentation fault", err.Error()); !ok || err != nil {
		t.Errorf("Expected segmentation fault")
		t.Error(err)
	}
}
func TestDeref2F(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	asm := fmt.Sprintf(`
		section '.code'
		@entry
			mov r%v, [0xffffffff]
		`, rA)
	b, err := assembleAndLink(asm)
	if err != nil {
		t.Error(err)
		t.Errorf("Compilation of:\n%s", asm)
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
	lines = append(lines, `
		section '.code'
		@entry
	`)
	for range DEFAULT_STACK_SIZE {
		val := rand.Intn(10000) - 5000
		if rand.Intn(100) < 50 {
			lines = append(lines, fmt.Sprintf("push %v", val))
		} else {
			rA := byte(rand.Int() % vm.GP_REG_MAX)
			lines = append(lines, fmt.Sprintf("mov r%v, %v", rA, val))
			lines = append(lines, fmt.Sprintf("push r%v", rA))
		}
	}
	for range DEFAULT_STACK_SIZE {
		rA := byte(rand.Int() % vm.GP_REG_MAX)
		lines = append(lines, fmt.Sprintf("pop r%v", rA))
	}
	asm = strings.Join(lines, "\n")
	if _, err := assembleAndExecute(asm); err != nil {
		t.Error(err)
		t.Errorf("Compilation of:\n%s", asm)
	}
}
func TestStack2(t *testing.T) {
	const stackSize = DEFAULT_STACK_SIZE
	stack := makeTestingStack(stackSize)
	a := uint64(rand.Intn(101) - 50)
	b := uint64(rand.Intn(101) - 50)
	c := a + b
	stack.push(a)
	stack.push(b)
	stack.push(c)
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	rB := (rA + 1) % vm.GP_REG_MAX
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov bp, sp
		mov [bp+1], %v
		mov [bp+2], %v
		mov r%v, [bp+1]
		mov r%v, [bp+2]
		add SIGNED r%v, r%v
		mov [bp+3], r%v
	`, stack[0], stack[1], rA, rB, rA, rB, rA)
	if mach, err := assembleAndExecute(asm); err != nil {
		t.Error(err)
		t.Errorf("Compilation of:\n%s", asm)
	} else if err := expectStack(&mach, vm.VmStack(stack)); err != nil {
		t.Error(err)
		t.Errorf("Compilation of:\n%s", asm)
	} else if err := expectGpRegisters(asm, &mach, ExpMap{
		rA: vm.Register(stack[2*8 : 2*8+9]),
		rB: vm.Register(stack[1*8 : 1*8+9]),
	}); err != nil {
		t.Error(err)
		t.Errorf("Compilation of:\n%s", asm)
	}
}
func TestStack1F(t *testing.T) {
	var asm string
	lines := make([]string, 0, DEFAULT_STACK_SIZE)
	lines = append(lines, `
		section '.code'
		@entry
	`)
	for range DEFAULT_STACK_SIZE + 1 {
		lines = append(lines, "push 1")
	}
	asm = strings.Join(lines, "\n")
	b, err := assembleAndLink(asm)
	if err != nil {
		t.Error(err)
		t.Errorf("Compilation of:\n%s", asm)
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
	section '.code'
	@entry
		pop
	`
	b, err := assembleAndLink(asm)
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
