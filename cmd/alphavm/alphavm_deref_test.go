package alphavm

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"testing"

	"github.com/JakubCygaro/alphataurus/pkg/assembler"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func TestDeref1(t *testing.T) {
	stackSize := (rand.Intn(32-5) + 5) * 8
	stack := makeTestingStack(stackSize)
	for range stackSize / 8 {
		stack.push(uint64(rand.Intn(101) - 50))
	}
	lines := make([]string, 0)
	lines = append(lines,
		"section '.code'",
		"@entry",
		"push WORD 0",
		"mov bp, sp",
	)
	for i := 0; i < stackSize; i += 8 {
		v := binary.BigEndian.Uint64(stack[i : i+8])
		lines = append(lines, fmt.Sprintf("mov WORD [bp+%v], %v", (i/8)*8, int64(v)))
	}
	asm := strings.Join(lines, "\n")
	b, err := assembleAndLink(asm)
	if err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
		return
	}
	if mach, err := executeStackSize(b, uint64(stackSize)); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if err := expectStack(&mach, vm.VmStack(stack)); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err.Error())
	}
}
func TestDeref2(t *testing.T) {
	stackSize := (rand.Intn(32-5) + 5) * 8
	stack := makeTestingStack(stackSize)
	for range stackSize / 8 {
		stack.push(uint64(rand.Intn(101) - 50))
	}
	lines := make([]string, 0)
	lines = append(lines,
		"section '.code'",
		"@entry",
		"push WORD 0",
		"mov bp, sp",
	)
	for i := 0; i < stackSize; i += 8 {
		v := binary.BigEndian.Uint64(stack[i : i+8])
		rA := byte(rand.Int() % vm.GP_REG_MAX)
		lines = append(lines, fmt.Sprintf("mov r%v, %v", rA, int64(v)))
		lines = append(lines, fmt.Sprintf("mov WORD [bp+%v], %v", (i/8)*8, int64(v)))
	}
	asm := strings.Join(lines, "\n")
	b, err := assembleAndLink(asm)
	if err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	}
	if mach, err := executeStackSize(b, uint64(len(stack))); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if err := expectStack(&mach, vm.VmStack(stack)); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err.Error())
	}
}
func TestDeref3(t *testing.T) {
	stackSize := (rand.Intn(32-5) + 5) * 8
	stack := makeTestingStack(0)
	for range stackSize / 8 {
		stack.push(uint64(rand.Intn(101) - 50))
	}
	lines := make([]string, 0)
	lines = append(lines,
		"section '.code'",
		"@entry",
		"push WORD 0",
		"mov bp, sp",
	)
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	lines = append(lines, fmt.Sprintf("mov r%v, 0", rA))
	for i := 0; i < stackSize; i += 8 {
		v := binary.BigEndian.Uint64(stack[i : i+8])
		lines = append(lines, fmt.Sprintf("mov WORD [bp+r%v], %v", rA, int64(v)))
		lines = append(lines, fmt.Sprintf("add UNSIGNED r%v, 8", rA))
	}
	asm := strings.Join(lines, "\n")
	b, err := assembleAndLink(asm)
	if err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
		return
	}
	if mach, err := executeStackSize(b, uint64(len(stack))); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if err := expectStack(&mach, vm.VmStack(stack)); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err.Error())
	}
}
func TestDeref4(t *testing.T) {
	rA := randomGpRegisterWord()
	rB := nextRandomGpRegister(rA)
	rC := nextRandomGpRegister(rB)
	asm := fmt.Sprintf(`
		section '.code'
		@entry
			mov r%v, bp
			mov BYTE [bp+1], 69
			mov r%vb, 1
			mov r%vb, [r%vh + r%vb]
			exit r%vb
		`, rA, rB, rC, rA, rB, rC)
	b, err := assembleAndLink(asm)
	if err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	}
	if s, err := execute(b); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if exit := s.GetExitCode(); exit != 69 {
		t.Error(compilationOfErr(asm))
		t.Errorf("Expected exit code %v got %v", 69, exit)
	}
}
func TestDeref5(t *testing.T) {
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	rB := (rA + 1) % vm.GP_REG_MAX
	rC := (rB + 1) % vm.GP_REG_MAX
	asm := fmt.Sprintf(`
		section '.code'
		@entry
			mov r%v, bp
			mov BYTE [bp+1], 69
			mov r%vb, 1
			mov r%vb, [r%vh + r%vb]
			exit r%vb
		`, rA, rB, rC, rA, rB, rC)
	b, err := assembleAndLink(asm)
	if err != nil {
		t.Errorf("Compilation of:\n%s", asm)
		t.Error(err)
	}
	if s, err := execute(b); err != nil {
		t.Errorf("Compilation of:\n %s", asm)
		t.Error(err)
	} else if exit := s.GetExitCode(); exit != 69 {
		t.Errorf("Compilation of:\n %s", asm)
		t.Errorf("Expected exit code %v got %v", 69, exit)
	}
}
func TestDeref6(t *testing.T) {
	stackSize := (rand.Intn(32-5) + 5) * 8
	stack := makeTestingStack(stackSize)
	for range stackSize / 8 {
		stack.push(uint64(rand.Intn(101) - 50))
	}
	code := make([]string, 0)
	codeSize := (len(code) + stackSize/8) * vm.INSTRUCTION_SIZE
	stackBase := codeSize + vm.ADDRESSDEADZONE_SIZE
	lines := make([]string, 0)
	lines = append(lines,
		"section '.code'",
		"@entry",
		fmt.Sprintf(";; code size should be %v (%v instructions)", codeSize, codeSize/12),
		fmt.Sprintf(";; stack base should thus be %v (0x%x)", stackBase, stackBase),
		fmt.Sprintf(";; fake stack size is %v", stackSize),
	)
	for i := 0; i < stackSize; i += 8 {
		v := binary.BigEndian.Uint64(stack[i : i+8])
		code = append(code,
			fmt.Sprintf(
				"mov WORD [0x%x], %v",
				stackBase+i+7,
				int64(v),
			),
		)
	}
	lines = append(lines, code...)
	asm := strings.Join(lines, "\n")
	b, err := assembleAndLink(asm)
	if err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
		return
	}
	if mach, err := executeStackSize(b, uint64(stackSize)); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if err := expectStack(&mach, vm.VmStack(stack)); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err.Error())
	}
}
func TestDeref7(t *testing.T) {
	stackSize := (rand.Intn(32-5) + 5) * 8
	stack := makeTestingStack(stackSize)
	for range stackSize / 8 {
		stack.push(uint64(rand.Intn(101) - 50))
	}
	code := make([]string, 0)
	pushCount := stackSize / 8
	movCount := pushCount
	assertMacroCount := movCount * 3
	codeSize := (len(code) + pushCount + movCount + assertMacroCount) *
		vm.INSTRUCTION_SIZE
	stackBase := codeSize + vm.ADDRESSDEADZONE_SIZE
	lines := make([]string, 0)
	lines = append(lines,
		"section '.code'",
		"@entry",
		fmt.Sprintf(";; code size should be %v (%v instructions)", codeSize, codeSize/12),
		fmt.Sprintf(";; stack base should thus be %v (0x%x)", stackBase, stackBase),
		fmt.Sprintf(";; fake stack size is %v", stackSize),
	)
	for i := 0; i < stackSize; i += 8 {
		v := binary.BigEndian.Uint64(stack[i : i+8])
		code = append(code,
			fmt.Sprintf(
				"push WORD %v",
				int64(v),
			),
		)
	}
	for i := 0; i < stackSize; i += 8 {
		r := randomGpRegisterWord()
		v := binary.BigEndian.Uint64(stack[i : i+8])
		code = append(code,
			fmt.Sprintf(
				"mov %s, [0x%x]",
				regStr(r, vm.SZ_64),
				stackBase+i+7,
			),
			macroAssertEqRI(
				assembler.RegisterData{Reg: int(r), Size: vm.SZ_64},
				v,
				UNSIGNED,
			),
		)
	}
	lines = append(lines, code...)
	asm := strings.Join(lines, "\n")
	b, err := assembleAndLink(asm)
	if err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
		return
	}
	if mach, err := executeStackSize(b, uint64(stackSize)); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if err := expectStack(&mach, vm.VmStack(stack)); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err.Error())
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
