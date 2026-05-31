package alphavm

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"testing"

	"github.com/JakubCygaro/alphataurus/pkg/linker"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
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
func TestJumps1(t *testing.T) {
	type opToTest struct {
		opcode   string
		testFunc func(a, b int32) bool
	}
	conds := make([]opToTest, 0)
	conds = append(conds, opToTest{
		opcode: "je",
		testFunc: func(a, b int32) bool {
			return a == b
		},
	})
	conds = append(conds, opToTest{
		opcode: "jz",
		testFunc: func(a, b int32) bool {
			return a == b
		},
	})
	conds = append(conds, opToTest{
		opcode: "jne",
		testFunc: func(a, b int32) bool {
			return a != b
		},
	})
	conds = append(conds, opToTest{
		opcode: "jnz",
		testFunc: func(a, b int32) bool {
			return a != b
		},
	})
	conds = append(conds, opToTest{
		opcode: "jg",
		testFunc: func(a, b int32) bool {
			return a > b
		},
	})
	conds = append(conds, opToTest{
		opcode: "jge",
		testFunc: func(a, b int32) bool {
			return a >= b
		},
	})
	conds = append(conds, opToTest{
		opcode: "jl",
		testFunc: func(a, b int32) bool {
			return a < b
		},
	})
	conds = append(conds, opToTest{
		opcode: "jle",
		testFunc: func(a, b int32) bool {
			return a <= b
		},
	})
	// first test IP relative jumps
	for _, c := range conds {
		a := rand.Int31n(10_000) - 5000
		b := rand.Int31n(10_000) - 5000
		asm := fmt.Sprintf(`
		section '.code'
		@entry
			mov r0, %v
			cmp r0, %v
			%s passed
			exit 0
		passed:
			exit 1
		`, a, b, c.opcode)
		var expect uint64
		if c.testFunc(a, b) {
			expect = 1
		} else {
			expect = 0
		}
		if mach, err := assembleAndExecute(asm); err != nil {
			t.Error(compilationOfErr(asm))
			t.Error(err)
		} else if exit := mach.GetExitCode(); exit != expect {
			t.Error(compilationOfErr(asm))
			t.Error(expectedExitCode(expect, exit))
		}
	}
	// then test absolute jumps
	for _, c := range conds {
		a := rand.Int31n(10_000) - 5000
		b := rand.Int31n(10_000) - 5000
		asm := fmt.Sprintf(`
		section '.code'
		@entry
			mov r0, %v
			cmp r0, %v
			%s ABSOLUTE passed
			exit 0
		passed:
			exit 1
		`, a, b, c.opcode)
		var expect uint64
		if c.testFunc(a, b) {
			expect = 1
		} else {
			expect = 0
		}
		ld := linker.NewLinker()
		if obj, err := assemble(asm); err != nil {
			t.Error(compilationOfErr(asm))
			t.Error(err)
		} else if h, err := vm.LoadObjFileHeader(
			bufio.NewReader(bytes.NewReader(obj))); err != nil {
			t.Error(compilationOfErr(asm))
			t.Error(err)
		} else if h.RelocsSize == 0 {
			t.Error(compilationOfErr(asm))
			t.Errorf("Object file does not contain relocations. " +
				"Relocations were expected with ABSOLUTE jumps")
		} else if elf, err := ld.Link(
			[]linker.LinkerInput{ linker.Bytes(obj) }); err != nil {
			t.Error(compilationOfErr(asm))
			t.Error(err)
		} else if mach, err := execute(elf); err != nil {
			t.Error(compilationOfErr(asm))
			t.Error(err)
		} else if exit := mach.GetExitCode(); exit != expect {
			t.Error(compilationOfErr(asm))
			t.Error(expectedExitCode(expect, exit))
		}
	}
}
func TestClr1(t *testing.T) {
	asm := `
	section '.code'
	@entry
		mov r0, 10
		cmp r0, 11
		clr
		jl fail
		cmp r0, 10
		clr
		je fail
		exit 0
	fail:
		exit 1
	`
	if mach, err := assembleAndExecute(asm); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if exit := mach.GetExitCode(); exit != 0 {
		t.Error(compilationOfErr(asm))
		t.Error(expectedExitCode(0, exit))
	}
}
