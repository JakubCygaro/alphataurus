package alphavm

import (
	"fmt"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
	"math"
	"math/rand"
	"testing"
)

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
			t.Error(err.Error())
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
			t.Error(err.Error())
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
