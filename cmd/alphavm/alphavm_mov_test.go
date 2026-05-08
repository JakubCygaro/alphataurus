package alphavm
import (
	"fmt"
	"math"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
	"testing"
)
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
