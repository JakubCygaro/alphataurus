package alphavm

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/JakubCygaro/alphataurus/pkg/assembler"
	"github.com/JakubCygaro/alphataurus/pkg/linker"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
	"math"
	"math/rand"
	"strings"
)

const (
	DEFAULT_STACK_SIZE = 16
)

func expectedExitCode(expected, got uint64) error {
	return fmt.Errorf("Expected exit code %v, but got %v", expected, got)
}
func compilationOfErr(asm string) error {
	return fmt.Errorf("Compilation of:\n%s", asm)
}
func assemblingErrorExpected() error {
	return fmt.Errorf("An assembling error was expected")
}
func randomGpRegisterWord() byte {
	return byte(rand.Int() % vm.GP_REG_MAX+1)
}
func randomGpRegisterWithSize() (byte, byte) {
	return randomGpRegisterWord(), byte(rand.Int()%vm.SZ_64 + 1)
}
func regStr(reg, sz byte) string {
	switch reg {
	case vm.BP_IDX:
		return "bp"
	case vm.IP_IDX:
		return "ip"
	case vm.SP_IDX:
		return "sp"
	default:
		var suf string = ""
		switch sz {
		case vm.SZ_8:
			suf = "b"
		case vm.SZ_16:
			suf = "q"
		case vm.SZ_32:
			suf = "h"
		}
		return fmt.Sprintf("r%v%s", reg, suf)
	}
}
func nextRandomGpRegister(reg byte) byte {
	return byte((reg + 1) % vm.GP_REG_MAX + 1)
}
func execute(elf vm.AlphaELFFile) (vm.VmState, error) {
	return executeStackSize(elf, DEFAULT_STACK_SIZE)
}
func executeStackSize(elf vm.AlphaELFFile, stacksz uint64) (vm.VmState, error) {
	mach := vm.CreateVmState(stacksz)
	if err := mach.Execute(elf); err != nil {
		return mach, err
	}
	return mach, nil
}
func assemble(source string) ([]byte, error) {
	asmblr := assembler.NewAssembler(*bufio.NewReader(strings.NewReader(source)))
	code, err := asmblr.Assemble()
	return code, err
}
func assembleAndLink(source string) (vm.AlphaELFFile, error) {
	code, err := assemble(source)
	if err != nil {
		return vm.AlphaELFFile{}, fmt.Errorf("Assembling error: %v", err)
	}
	ld := linker.NewLinker()
	linked, err := ld.Link([]linker.LinkerInput{linker.Bytes(code)})
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
		return fmt.Errorf(msg)
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
