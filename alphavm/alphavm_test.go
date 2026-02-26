package alphavm

import (
	"bufio"
	"fmt"
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
