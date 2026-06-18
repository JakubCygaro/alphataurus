package alphavm

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/JakubCygaro/alphataurus/pkg/assembler"
	"github.com/JakubCygaro/alphataurus/pkg/linker"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
	aelf "github.com/JakubCygaro/alphataurus/pkg/vm/aelf"
	"math"
	tc "github.com/JakubCygaro/alphataurus/internal/pkg/tests_commons"
	"github.com/JakubCygaro/alphataurus/pkg/vm/decls"
	"strings"
)

const (
	DEFAULT_STACK_SIZE = 16
)

func expectedExitCode(expected, got uint64) error {
	return fmt.Errorf("Expected exit code %v, but got %v", expected, got)
}
func compilationOfErr(asm ...string) error {
	return fmt.Errorf("Compilation of:\n%s", strings.Join(asm, "---\n"))
}
func assemblingErrorExpected() error {
	return fmt.Errorf("An assembling error was expected")
}
func execute(elf aelf.AlphaELFFile) (vm.VmState, error) {
	return executeStackSize(elf, DEFAULT_STACK_SIZE)
}
func executeStackSize(elf aelf.AlphaELFFile, stacksz uint64) (vm.VmState, error) {
	mach := vm.CreateVmState(stacksz)
	if err := mach.Execute(elf); err != nil {
		return mach, err
	}
	return mach, nil
}
func assemble(source string) ([]byte, error) {
	asmblr := assembler.NewAssembler(bufio.NewReader(strings.NewReader(source)))
	code, err := asmblr.Assemble()
	return code, err
}
func assembleAndLink(source ...string) (aelf.AlphaELFFile, error) {
	assembled := make([]linker.LinkerInput, 0)
	for _, s := range source {
		if code, err := assemble(s); err != nil {
			return aelf.AlphaELFFile{}, fmt.Errorf("Assembling error: %v", err)
		} else {
			assembled = append(assembled, linker.Bytes(code))
		}
	}
	ld := linker.NewLinker()
	linked, err := ld.Link(assembled)
	if err != nil {
		err = fmt.Errorf("Linking error: %v", err)
	}
	return linked, err
}
func assembleAndExecute(source string) (vm.VmState, error) {
	elf, err := assembleAndLink(source)
	mach := vm.CreateVmState(1024)
	if err != nil {
		return mach, err
	}
	err = mach.Execute(elf)
	if err != nil {
		return mach, err
	}
	return mach, nil
}

type ExpMap map[byte]vm.Register

func expectGpRegisters(asm string, mach *vm.VmState, regStates ExpMap) error {
	var msg string
	err := false
	for k, tr := range regStates {
		var tro any
		tr.GetValAs(vm.TY_UINT, vm.SZ_64, &tro)
		trv := tro.(uint64)
		if r, _ := mach.GetGpRXAsU64(k); r != trv {
			msg = strings.Join([]string{
				fmt.Sprintf("State of general purpose register r%v was different from expected", k),
				fmt.Sprintf("\t[uint64]  expected (%v) \t got (%v)", trv, r),
				fmt.Sprintf("\t[int64]   expected (%v) \t got (%v)", int64(trv), int64(r)),
				fmt.Sprintf("\t[float64] expected (%v) \t got (%v)", math.Float64frombits(trv), math.Float64frombits(r)),
			}, "\n")
			err = true
		}
	}
	if err {
		msg = strings.Join([]string{
			fmt.Sprintf("Compilation of %s", asm),
			msg,
		}, "\n")
		return fmt.Errorf("%s", msg)
	}
	return nil
}

type testingStack vm.VmStack

func makeTestingStack(capacity int) testingStack {
	return make(testingStack, 0, capacity)
}

func (s *testingStack) push(value any) any {
	if u8, ok := value.(uint8); ok {
		*s = append(*s, u8)
	} else if u16, ok := value.(uint16); ok {
		*s = binary.BigEndian.AppendUint16(*s, u16)
	} else if u32, ok := value.(uint32); ok {
		*s = binary.BigEndian.AppendUint32(*s, u32)
	} else if u64, ok := value.(uint64); ok {
		*s = binary.BigEndian.AppendUint64(*s, u64)
	}
	return value
}

func expectStack(mach *vm.VmState, stack vm.VmStack) error {
	lines := make([]string, 0)
	vmStack := mach.GetStack()
	if len(vmStack) != len(stack) {
		lines = append(lines,
			fmt.Sprintf("Input stack and vm stack sizes do not match\nlen(vm) = %v\tlen(input) = %v",
				len(vmStack), len(stack)))
		return fmt.Errorf("%s", strings.Join(lines, "\n"))
	}
	for i, v := range stack {
		if vmStack[i] != v {
			lines = append(lines, fmt.Sprintf("Stack value at [%v] was different from expected", i))
			// lines = append(lines,
			// 	fmt.Sprintf("\t[uint64]  expected (%v) \t got (%v)",
			// 		v, vmStack[i]),
			// 	fmt.Sprintf("\t[int64]   expected (%v) \t got (%v)",
			// 		int64(v), int64(vmStack[i])),
			// 	fmt.Sprintf("\t[float64] expected (%v) \t got (%v)",
			// 		math.Float64frombits(v), math.Float64frombits(vmStack[i])),
			// )
			lines = append(lines,
				fmt.Sprintf("\tgot (%v) expected (%v)", vmStack[i], v),
			)
			lines = append(lines,
				fmt.Sprintf("\tgot (0x%x) expected (0x%x)", vmStack[i], v),
			)
			break
		}
	}
	if len(lines) > 0 {
		lines = append(lines,
			fmt.Sprintf("vm stack: %v", vmStack),
		)
		lines = append(lines,
			fmt.Sprintf("input stack: %v", stack),
		)
		return fmt.Errorf("%s", strings.Join(lines, "\n"))
	}
	return nil
}

type AssertTy int

const (
	ASSERT_EQ AssertTy = iota
	ASSERT_NEQ
	ASSERT_G
	ASSERT_GE
	ASSERT_L
	ASSERT_LE
)

func assertToOp(ty AssertTy) string {
	switch ty {
	case ASSERT_EQ:
		return "je"
	case ASSERT_NEQ:
		return "jne"
	case ASSERT_G:
		return "jg"
	case ASSERT_GE:
		return "jge"
	case ASSERT_L:
		return "jl"
	case ASSERT_LE:
		return "jle"
	default:
		panic("unknown assert type")
	}
}

type assertTypeKwd string

const (
	UNSIGNED assertTypeKwd = "UNSIGNED"
	SIGNED   assertTypeKwd = "SIGNED"
	FLOAT    assertTypeKwd = "FLOAT"
)

func toReg(r int, sz byte) assembler.RegisterData {
	return assembler.RegisterData{
		Reg:  r,
		Size: sz,
	}
}

// assert equality of r and expect, exit with code equal to expect on failure
func macroAssertEqRI(r assembler.RegisterData,
	expect any, cmpType assertTypeKwd) string {
	var exitV string
	if cmpType == FLOAT {
		if f, ok := expect.(float64); ok {
			exitV = fmt.Sprintf("0x%x", math.Float64bits(f))
		} else if x, ok := expect.(uint64); ok {
			exitV = fmt.Sprintf("0x%x", x)
		}
	} else {
		exitV = fmt.Sprintf("%v", expect)
	}
	return fmt.Sprintf(
		`cmp %s %s, %v
je [ip+%v]
exit %v`,
		cmpType,
		tc.RegStr(byte(r.Reg), r.Size),
		expect,
		decls.INSTRUCTION_SIZE,
		exitV,
	)
}

// assert comparison of r and expect, exit with code equal to expect on failure
//
// comparison is r (AssertTy) expect
func macroAssertUGenericRR(r, expect assembler.RegisterData, aT AssertTy) string {
	return fmt.Sprintf(`
		cmp UNSIGNED %s, %s
		%s [ip+%v]
		exit %v
	`,
		tc.RegStr(byte(r.Reg), r.Size),
		tc.RegStr(byte(expect.Reg), expect.Size),
		assertToOp(aT),
		decls.INSTRUCTION_SIZE,
		tc.RegStr(byte(expect.Reg), expect.Size),
	)
}
func macroAssertUGenericIR(r assembler.RegisterData, expect uint64, aT AssertTy) string {
	return fmt.Sprintf(`
		cmp UNSIGNED %s, %v
		%s [ip+%v]
		exit %v
	`,
		tc.RegStr(byte(r.Reg), r.Size),
		expect,
		assertToOp(aT),
		decls.INSTRUCTION_SIZE,
		expect,
	)
}

func getComparison(a, b uint64) AssertTy {
	switch {
	case a == b:
		return ASSERT_EQ
	default:
		return ASSERT_NEQ
	}
}
