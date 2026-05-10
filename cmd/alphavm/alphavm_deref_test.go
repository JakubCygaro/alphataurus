package alphavm

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"testing"

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
		"push BYTE 0",
		"mov bp, sp",
	)
	for i := 0; i < stackSize; i += 8 {
		v := binary.BigEndian.Uint64(stack[i : i+8])
		lines = append(lines, fmt.Sprintf("mov WORD [bp+%v], %v", (i/8)*8, int64(v)))
	}
	asm := strings.Join(lines, "\n")
	b, err := assembleAndLink(asm)
	if err != nil {
		t.Error(err)
		t.Errorf("Compilation of:\n%s", asm)
		return
	}
	if mach, err := executeStackSize(b, uint64(stackSize)); err != nil {
		t.Error(err)
		t.Errorf("Compilation of:\n%s", asm)
	} else if err := expectStack(&mach, vm.VmStack(stack)); err != nil {
		t.Error(err.Error())
		t.Errorf("Compilation of:\n%s", asm)
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
		"push BYTE 0",
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
	stackSize := (rand.Intn(32-5) + 5) * 8
	stack := makeTestingStack(0)
	for range stackSize / 8 {
		stack.push(uint64(rand.Intn(101) - 50))
	}
	lines := make([]string, 0)
	lines = append(lines,
		"section '.code'",
		"@entry",
		"push BYTE 0",
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
		t.Errorf("Compilation of:\n%s", asm)
		t.Error(err)
		return
	}
	if mach, err := executeStackSize(b, uint64(len(stack))); err != nil {
		t.Errorf("Compilation of:\n%s", asm)
		t.Error(err)
	} else if err := expectStack(&mach, vm.VmStack(stack)); err != nil {
		t.Errorf("Compilation of:\n%s", asm)
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
func TestDeref4(t *testing.T) {
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
