package alphavm

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"testing"

	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func TestStack1(t *testing.T) {
	var asm string
	lines := make([]string, 0, DEFAULT_STACK_SIZE*8)
	lines = append(lines, `
		section '.code'
		@entry
	`)
	for range DEFAULT_STACK_SIZE {
		val := rand.Intn(10000) - 5000
		if rand.Intn(100) < 50 {
			lines = append(lines, fmt.Sprintf("push WORD %v", val))
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
	stack := makeTestingStack(0)
	a := uint64(rand.Intn(101) - 50)
	b := uint64(rand.Intn(101) - 50)
	c := a + b
	stack.push(uint64(0))
	stack.push(a)
	stack.push(b)
	stack.push(c)
	rA := byte(rand.Int() % vm.GP_REG_MAX)
	rB := (rA + 1) % vm.GP_REG_MAX
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		push WORD 0
		mov bp, sp
		mov WORD [bp+0], %v
		mov WORD [bp+8], %v
		mov r%v, [bp+0]
		mov r%v, [bp+8]
		add SIGNED r%v, r%v
		mov [bp+16], r%v
	`, a, b, rA, rB, rA, rB, rA)
	elf, err := assembleAndLink(asm)
	if err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
		return
	}
	if mach, err := executeStackSize(elf, uint64(len(stack))); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if err := expectStack(&mach, vm.VmStack(stack)); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if err := expectGpRegisters(asm, &mach, ExpMap{
		rA: vm.Register(stack[16:24]),
		rB: vm.Register(stack[8:16]),
	}); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
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
		pop WORD
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
