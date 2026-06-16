package alphavm

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"testing"

	tc "github.com/JakubCygaro/alphataurus/internal/pkg/tests_commons"
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
	rA := tc.RandomGpRegisterWord()
	rB := tc.NextRandomGpRegister(rA)
	rC := tc.NextRandomGpRegister(rB)
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
		r := tc.RandomGpRegisterWord()
		v := binary.BigEndian.Uint64(stack[i : i+8])
		code = append(code,
			fmt.Sprintf(
				"mov %s, [0x%x]",
				tc.RegStr(r, vm.SZ_64),
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
func TestMovDRI1(t *testing.T) {
	into := makeIntoRegistersList()
	lines := make([]string, 0)
	lines = append(lines,
		"section '.code'",
		"@entry",
		"_start:",
	)
	for _, ir := range into {
		if !vm.IsArthRAllowed(byte(ir.Reg)) ||
			ir.Reg == vm.SP_IDX {
			continue
		}
		var val uint64
		switch ir.Size {
		case vm.SZ_8:
			val = uint64(byte(rand.Int63()))
		case vm.SZ_16:
			val = uint64(uint16(rand.Int63()))
		case vm.SZ_32:
			val = uint64(uint32(rand.Int63()))
		case vm.SZ_64:
			val = uint64(rand.Int63())
		}
		sz, _ := assembler.GetSizeKeyword(ir.Size)
		// bytes := vm.DataSizeToByteCount(ir.Size)
		lines = append(lines,
			fmt.Sprintf(
				"push %s %v",
				sz, val,
			),
			fmt.Sprintf(
				"mov %s, [sp]",
				tc.RegStr(byte(ir.Reg), ir.Size),
			),
			fmt.Sprintf(
				"pop %s",
				sz,
			),
			macroAssertEqRI(
				assembler.RegisterData(ir),
				val,
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
func TestMovRDO1_1(t *testing.T) {
	stackSize := (rand.Intn(32-5) + 5) * 8
	stack := makeTestingStack(stackSize)
	for range stackSize / 8 {
		stack.push(uint64(rand.Intn(101) - 50))
	}
	code := make([]string, 0)
	codeSize := (len(code) + (stackSize/8)*2) * vm.INSTRUCTION_SIZE
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
		randomOffsetReg := tc.RandomGpRegisterWord()
		code = append(code,
			fmt.Sprintf(
				"mov %s, 0x%x",
				tc.RegStr(randomOffsetReg, vm.SZ_64),
				stackBase+i+7,
			),
			fmt.Sprintf(
				"mov WORD [0x0+%s], %v",
				tc.RegStr(randomOffsetReg, vm.SZ_64),
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
func TestMovRDO1_2(t *testing.T) {
	val := rand.Int63n(10_000) - 5_000
	r := tc.RandomGpRegisterWord()
	// off := tc.NextRandomGpRegister(r)
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		push WORD %v+1
		push WORD %v+1
		mov bp, sp
		mov WORD [bp-8], %v
		mov %s, [bp-8]
		%s
		mov %s, [bp-8]
		%s
		mov %s, [bp-8]
		%s
		mov %s, [bp-8]
		%s
		exit 0
	`,
		val,
		val,
		val,
		tc.RegStr(r, vm.SZ_64),
		macroAssertEqRI(toReg(int(r), vm.SZ_64), val, SIGNED),
		tc.RegStr(r, vm.SZ_64),
		macroAssertEqRI(toReg(int(r), vm.SZ_32), val, SIGNED),
		tc.RegStr(r, vm.SZ_64),
		macroAssertEqRI(toReg(int(r), vm.SZ_16), val, SIGNED),
		tc.RegStr(r, vm.SZ_64),
		macroAssertEqRI(toReg(int(r), vm.SZ_8), val, SIGNED),
	)
	b, err := assembleAndLink(asm)
	if err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
		return
	}
	if mach, err := executeStackSize(b, 8*2); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if exitV := mach.GetExitCode(); exitV != 0 {
		t.Error(compilationOfErr(asm))
		t.Error(expectedExitCode(0, exitV))
	}
}
func TestMovRDO1_3(t *testing.T) {
	val := rand.Int63n(10_000) - 5_000
	r := tc.RandomGpRegisterWord()
	// off := tc.NextRandomGpRegister(r)
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		push WORD %v+1
		push WORD %v+1
		mov bp, sp
		mov r0, bp
		sub UNSIGNED r0, 8
		add r0, r0
		mov WORD [r0/2], %v
		mov %s, [r0/2]
		%s
		mov %s, [r0/2]
		%s
		mov %s, [r0/2]
		%s
		mov %s, [r0/2]
		%s
		exit 0
	`,
		val,
		val,
		val,
		tc.RegStr(r, vm.SZ_64),
		macroAssertEqRI(toReg(int(r), vm.SZ_64), val, SIGNED),
		tc.RegStr(r, vm.SZ_64),
		macroAssertEqRI(toReg(int(r), vm.SZ_32), val, SIGNED),
		tc.RegStr(r, vm.SZ_64),
		macroAssertEqRI(toReg(int(r), vm.SZ_16), val, SIGNED),
		tc.RegStr(r, vm.SZ_64),
		macroAssertEqRI(toReg(int(r), vm.SZ_8), val, SIGNED),
	)
	b, err := assembleAndLink(asm)
	if err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
		return
	}
	if mach, err := executeStackSize(b, 8*2); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if exitV := mach.GetExitCode(); exitV != 0 {
		t.Error(compilationOfErr(asm))
		t.Error(expectedExitCode(0, exitV))
	}
}
func TestMovRDO1_4(t *testing.T) {
	val := rand.Int63n(10_000) - 5_000
	r := tc.RandomGpRegisterWord()
	// off := tc.NextRandomGpRegister(r)
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		push WORD %v
		;; align the stack on a multiple of 2
		rsh sp, 1
		lsh sp, 1
		push WORD %v+1
		push WORD %v+1
		mov bp, sp
		mov r0, bp
		sub UNSIGNED r0, 8
		mov r1, 2
		div UNSIGNED WORD
		mov r0, r2
		mov WORD [r0*2], %v
		mov %s, [r0*2]
		%s
		mov %s, [r0*2]
		%s
		mov %s, [r0*2]
		%s
		mov %s, [r0*2]
		%s
		exit 0
	`,
		val,
		val,
		val,
		val,
		tc.RegStr(r, vm.SZ_64),
		macroAssertEqRI(toReg(int(r), vm.SZ_64), val, SIGNED),
		tc.RegStr(r, vm.SZ_64),
		macroAssertEqRI(toReg(int(r), vm.SZ_32), val, SIGNED),
		tc.RegStr(r, vm.SZ_64),
		macroAssertEqRI(toReg(int(r), vm.SZ_16), val, SIGNED),
		tc.RegStr(r, vm.SZ_64),
		macroAssertEqRI(toReg(int(r), vm.SZ_8), val, SIGNED),
	)
	b, err := assembleAndLink(asm)
	if err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
		return
	}
	if mach, err := executeStackSize(b, 8*3); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if exitV := mach.GetExitCode(); exitV != 0 {
		t.Error(compilationOfErr(asm))
		t.Error(expectedExitCode(0, exitV))
	}
}
func TestMovRDO2_1(t *testing.T) {
	stackSize := (rand.Intn(32-5) + 5) * 8
	stack := makeTestingStack(stackSize)
	for range stackSize / 8 {
		stack.push(uint64(rand.Intn(101) - 50))
	}
	code := make([]string, 0)
	codeSize := (len(code) + (stackSize/8)*3) * vm.INSTRUCTION_SIZE
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
		randomOffsetRegA := tc.RandomGpRegisterWord()
		randomOffsetRegB := tc.NextRandomGpRegister(randomOffsetRegA)
		off := stackBase + i + 7
		offHalf1 := off / 2
		offHalf2 := off - offHalf1
		code = append(code,
			fmt.Sprintf(
				"mov %s, 0x%x",
				tc.RegStr(randomOffsetRegA, vm.SZ_64),
				offHalf1,
			),
			fmt.Sprintf(
				"mov %s, 0x%x",
				tc.RegStr(randomOffsetRegB, vm.SZ_64),
				offHalf2,
			),
			fmt.Sprintf(
				"mov WORD [0x0+%s+%s], %v",
				tc.RegStr(randomOffsetRegA, vm.SZ_64),
				tc.RegStr(randomOffsetRegB, vm.SZ_64),
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
func TestMovRDO2_2(t *testing.T) {
	// fStack := makeTestingStack(0)
	val := rand.Int63n(10_000) - 5_000
	r := tc.RandomGpRegisterWord()
	off := tc.NextRandomGpRegister(r)
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		push WORD %v
		mov %s, 8
		mov %s, [bp+%s]
		%s
		mov %s, [bp+%s]
		%s
		mov %s, [bp+%s]
		%s
		mov %s, [bp+%s]
		%s
		exit 0
	`,
		val,
		tc.RegStr(off, vm.SZ_64),
		tc.RegStr(r, vm.SZ_64), tc.RegStr(off, vm.SZ_64),
		macroAssertEqRI(toReg(int(r), vm.SZ_64), val, SIGNED),
		tc.RegStr(r, vm.SZ_64), tc.RegStr(off, vm.SZ_32),
		macroAssertEqRI(toReg(int(r), vm.SZ_64), val, SIGNED),
		tc.RegStr(r, vm.SZ_64), tc.RegStr(off, vm.SZ_16),
		macroAssertEqRI(toReg(int(r), vm.SZ_64), val, SIGNED),
		tc.RegStr(r, vm.SZ_64), tc.RegStr(off, vm.SZ_8),
		macroAssertEqRI(toReg(int(r), vm.SZ_64), val, SIGNED),
	)
	b, err := assembleAndLink(asm)
	if err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
		return
	}
	if mach, err := executeStackSize(b, 8); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if exitV := mach.GetExitCode(); exitV != 0 {
		t.Error(compilationOfErr(asm))
		t.Error(expectedExitCode(0, exitV))
	}
}
func TestMovRDO2_3(t *testing.T) {
	val := rand.Int63n(10_000) - 5_000
	r := tc.RandomGpRegisterWord()
	off := tc.NextRandomGpRegister(r)
	asm := fmt.Sprintf(`
	section '.code'
	@entry
		push WORD %v
		push WORD %v+1
		mov bp, sp
		mov %s, -8
		mov %s, [bp+%s]
		%s
		mov %s, [bp+%s]
		%s
		mov %s, [bp+%s]
		%s
		mov %s, [bp+%s]
		%s
		exit 0
	`,
		val,
		val,
		tc.RegStr(off, vm.SZ_64),
		tc.RegStr(r, vm.SZ_64), tc.RegStr(off, vm.SZ_64),
		macroAssertEqRI(toReg(int(r), vm.SZ_64), val, SIGNED),
		tc.RegStr(r, vm.SZ_64), tc.RegStr(off, vm.SZ_32),
		macroAssertEqRI(toReg(int(r), vm.SZ_64), val, SIGNED),
		tc.RegStr(r, vm.SZ_64), tc.RegStr(off, vm.SZ_16),
		macroAssertEqRI(toReg(int(r), vm.SZ_64), val, SIGNED),
		tc.RegStr(r, vm.SZ_64), tc.RegStr(off, vm.SZ_8),
		macroAssertEqRI(toReg(int(r), vm.SZ_64), val, SIGNED),
	)
	b, err := assembleAndLink(asm)
	if err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
		return
	}
	if mach, err := executeStackSize(b, 8*2); err != nil {
		t.Error(compilationOfErr(asm))
		t.Error(err)
	} else if exitV := mach.GetExitCode(); exitV != 0 {
		t.Error(compilationOfErr(asm))
		t.Error(expectedExitCode(0, exitV))
	}
}
