package alphavm

import (
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strings"
	"testing"

	"github.com/JakubCygaro/alphataurus/pkg/assembler"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
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
		t.Error(err.Error())
	}
}
func TestMov2(t *testing.T) {
	r0_v, r1_v, r2_v, r3_v := 69, 420, 123, 123
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov r0b, %d
		mov r1q, %d
		mov r2h, %d
		mov r3, r2
	`, r0_v, r1_v, r2_v)
	mach, err := assembleAndExecute(asm)
	if err != nil {
		t.Error(err)
	}
	if err := expectGpRegisters(asm, &mach, ExpMap{
		vm.R0_IDX: vm.RegisterWithValueSized(uint64(r0_v), vm.SZ_8),
		vm.R1_IDX: vm.RegisterWithValueSized(uint64(r1_v), vm.SZ_16),
		vm.R2_IDX: vm.RegisterWithValueSized(uint64(r2_v), vm.SZ_32),
		vm.R3_IDX: vm.RegisterWithValue(uint64(r3_v)),
	}); err != nil {
		t.Error(err.Error())
	}
}
func TestMov3(t *testing.T) {
	rA, rAsz := randomGpRegisterWord(), byte(vm.SZ_64)
	rAv := rand.Int63n(int64(math.MaxUint32))
	rB, rBsz := randomGpRegisterWord(), byte(vm.SZ_32)
	rBv := rand.Int63n(int64(math.MaxUint16))
	rC, rCsz := randomGpRegisterWord(), byte(vm.SZ_16)
	rCv := rand.Int63n(int64(math.MaxUint8))
	rD, rDsz := randomGpRegisterWord(), byte(vm.SZ_8)
	rDv := rand.Int63n(int64(math.MaxUint8))
	rE := randomGpRegisterWord()
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		push WORD 0
		mov bp, sp
		mov WORD [bp],    0x00000000%x
		mov HALF [bp],    0x0000%x
		mov QUARTER [bp], 0x00%x
		mov BYTE [bp],    0x%x
		mov %s, [bp]
		mov %s, [bp]
		and %s,           0x00ff
		mov %s, [bp]
		and %s, 		  0x0000ffff
		mov %s, [bp]
		and %s,           0x00000000ffffffff
		mov %s, [bp]
		sub %s, %s
		exit %s
	`, rAv, rBv, rCv, rDv,
		regStr(rA, rAsz), regStr(rB, rBsz), regStr(rB, rBsz),
		regStr(rC, rCsz), regStr(rC, rCsz), regStr(rD, rDsz), regStr(rD, rDsz),
		regStr(rE, vm.SZ_8), regStr(rD, rDsz), regStr(rE, vm.SZ_8),
		regStr(rD, rDsz),
	)
	if mach, err := assembleAndExecute(asm); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if mach.GetExitCode() != 0 {
		t.Error(compilationOfErr(asm))
		t.Error(expectedExitCode(0, mach.GetExitCode()))
	}
}
func TestMov1F(t *testing.T) {
	rA, rAsz := randomGpRegisterWord(), byte(vm.SZ_8)
	rB, rBsz := nextRandomGpRegister(rA), byte(vm.SZ_64)
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov %s, %s
	`, regStr(rA, rAsz), regStr(rB, rBsz))
	if _, err := assemble(asm); err != nil {
		if ok, _ := regexp.MatchString("Mismatched register sizes", err.Error()); !ok {
			t.Error(compilationOfErr(asm))
			t.Errorf("Expected mismatched register sizes assembling error, " +
				"got a different error")
		}
	} else {
		t.Error(compilationOfErr(asm))
		t.Error(assemblingErrorExpected())
	}
}
func TestMov2F(t *testing.T) {
	rA, rAsz := byte(vm.IP_IDX), byte(vm.SZ_64)
	rB, rBsz := randomGpRegisterWord(), byte(vm.SZ_64)
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		mov %s, %s
	`, regStr(rA, rAsz), regStr(rB, rBsz))
	if _, err := assemble(asm); err != nil {
		if ok, _ := regexp.MatchString("Disallowed destination register", err.Error()); !ok {
			t.Error(compilationOfErr(asm))
			t.Errorf("Expected disallowed destination register error, " +
				"got a different error")
		}
	} else {
		t.Error(compilationOfErr(asm))
		t.Error(assemblingErrorExpected())
	}
}

type reg assembler.RegisterData

func makeFromRegistersList() []reg {
	from := make([]reg, 0)
	for r := range vm.GP_REG_MAX + 1 {
		if vm.IsMovFromRAllowed(byte(r)) {
			for sz := range vm.MAX_SZ + 1 {
				from = append(from, reg{r, byte(sz)})
			}
		}
	}
	for r := vm.GP_REG_MAX + 1; r <= vm.MAX_REG_IDX; r++ {
		if vm.IsMovFromRAllowed(byte(r)) {
			from = append(from, reg{r, byte(vm.SZ_64)})
		}
	}
	return from
}
func makeIntoRegistersList() []reg {
	into := make([]reg, 0)
	for r := range vm.GP_REG_MAX + 1 {
		if vm.IsMovIntoRAllowed(byte(r)) {
			for sz := range vm.MAX_SZ + 1 {
				into = append(into, reg{r, byte(sz)})
			}
		}
	}
	for r := vm.GP_REG_MAX + 1; r <= vm.MAX_REG_IDX; r++ {
		if vm.IsMovIntoRAllowed(byte(r)) {
			into = append(into, reg{r, byte(vm.SZ_64)})
		}
	}
	return into
}

// test moving a value between every and each register
func TestAllRRMoves1(t *testing.T) {
	from := makeFromRegistersList()
	into := makeIntoRegistersList()
	lines := make([]string, 0)
	lines = append(lines,
		"section '.code'",
		"@entry",
		"_start:",
	)
	for _, fr := range from {
		for _, ir := range into {
			if ir.Size < fr.Size {
				continue
			}
			lines = append(lines,
				fmt.Sprintf(
					"mov %s, %s",
					regStr(byte(ir.Reg), ir.Size),
					regStr(byte(fr.Reg), fr.Size),
				),
			)
		}
	}
	lines = append(lines,
		"exit 0")
	asm := strings.Join(lines, "\n")
	if _, err := assembleAndExecute(asm); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	}
}

// test moving a value between every and each register
func TestAllIRMoves1(t *testing.T) {
	into := makeIntoRegistersList()
	lines := make([]string, 0)
	lines = append(lines,
		"section '.code'",
		"@entry",
		"_start:",
	)
	for _, ir := range into {
		val := rand.Int63()
		if val == 0 {
			val += 1
		}
		lines = append(lines,
			fmt.Sprintf(
				"mov %s, %v",
				regStr(byte(ir.Reg), ir.Size),
				val,
			),
		)
		lines = append(lines,
			macroAssertEqRI(
				assembler.RegisterData(ir),
				uint64(val),
				UNSIGNED,
			),
		)
	}
	lines = append(lines,
		"exit 0")
	asm := strings.Join(lines, "\n")
	if mach, err := assembleAndExecute(asm); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if exitCode := mach.GetExitCode(); exitCode != 0 {
		t.Error(compilationOfErr(asm))
		t.Error(expectedExitCode(0, exitCode))
	}
}
