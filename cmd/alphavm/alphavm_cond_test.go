package alphavm
import (
	"fmt"
	"math/rand"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
	"testing"
)
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
		t.Error(compilationOfErr(asm))
		t.Error(err)
		t.FailNow()
	}
	flags := mach.GetFlags()
	if !flags.Sf {
		t.Error(compilationOfErr(asm))
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
		t.Error(compilationOfErr(asm))
		t.Error(err)
		t.FailNow()
	}
	flags := mach.GetFlags()
	if !flags.Sf {
		t.Error(compilationOfErr(asm))
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
		t.Error(compilationOfErr(asm))
		t.Error(err)
		return
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R0_IDX: vm.RegisterWithValue(0),
		vm.R1_IDX: vm.RegisterWithValue(10),
		vm.R4_IDX: vm.RegisterWithValue(420),
		vm.R5_IDX: vm.RegisterWithValue(1337),
	}); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err.Error())
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
		t.Error(compilationOfErr(asm))
		t.Error(err)
		return
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R1_IDX: vm.RegisterWithValue(10),
		vm.R0_IDX: vm.RegisterWithValue(0),
	}); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
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
		jg 0x1000 + 12
	`
	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
		return
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R1_IDX: vm.RegisterWithValue(10),
		vm.R0_IDX: vm.RegisterWithValue(0),
	}); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	}
}
