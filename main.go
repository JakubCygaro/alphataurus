package main

import (
	"bufio"
	"fmt"
	// "math/rand"
	"os"
	"strings"

	"github.com/JakubCygaro/alphataurus/assembler"
	"github.com/JakubCygaro/alphataurus/internal/vm"
)

const assembly = `
mov r0, 1.0
mov r1, 2.0
sub FLOAT r0, r1
cmp FLOAT r0, r1
`

func main() {
	// rA := 0
	// rB := 1
	// reg_v := rand.Float64()
	// // reg_sub := rand.Float64()
	// assembly := fmt.Sprintf(`
	// mov r%v, %v
	// mov r%v, -%v
	// sub FLOAT r%v, r%v
	// `, rA, reg_v, rB, reg_v, rA, rB)
	asm := assembler.NewAssembler(*bufio.NewReader(strings.NewReader(assembly)))
	bytecode, iCount, err := asm.EmitBytecode()
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(-1)
	} else {
		stride := int(len(bytecode) / iCount)
		fmt.Printf("stride %d \n", stride)
		for i := 0; i < len(bytecode)/stride; i++ {
			fmt.Println(bytecode[i*stride : (i+1)*stride])
		}
	}
	fmt.Printf("emited bytecode size: %d\n", len(bytecode))
	fmt.Printf("emited %d instructions\n", iCount)
	mach := vm.CreateVmState(1024)
	err = mach.Execute(bytecode)
	if err != nil {
		fmt.Println(err)
	}
	res, _ := mach.GetGpRXAsFloat64(vm.R0_IDX)
	// fmt.Println(, 1)
	fmt.Printf("r0 = %v\n", res)
	// mov := opCodes[vm.OP_MOVIR]
	// //load in the loop register
	// bytecode = binary.BigEndian.AppendUint32(bytecode, uint32(mov))
	// bytecode[len(bytecode)-4] = 0b00000101 //r5
	// bytecode = binary.BigEndian.AppendUint64(bytecode, 10)
	//
	// bytecode = binary.BigEndian.AppendUint32(bytecode, uint32(mov))
	// bytecode[len(bytecode)-4] = 0b00000000 //r0
	// bytecode = binary.BigEndian.AppendUint64(bytecode, 2)
	//
	// bytecode = binary.BigEndian.AppendUint32(bytecode, uint32(mov))
	// bytecode[len(bytecode)-4] = 0b00000001 //r1
	// bytecode = binary.BigEndian.AppendUint64(bytecode, 4)
	//
	// divrr := opCodes[vm.OP_DIVRR]
	// bytecode = binary.BigEndian.AppendUint32(bytecode, uint32(divrr))
	// bytecode = binary.BigEndian.AppendUint64(bytecode, 0x00000200000000)
	//
	// bytecode = binary.BigEndian.AppendUint32(bytecode, uint32(mov))
	// bytecode[len(bytecode)-4] = 0b00100000 //r2 -> r0
	// bytecode = binary.BigEndian.AppendUint64(bytecode, 4)
	//
	// cmp := opCodes[vm.OP_CMP]
	// bytecode = binary.BigEndian.AppendUint32(bytecode, uint32(cmp))
	// bytecode[len(bytecode)-4] = 0b11110101 //r5
	// bytecode = binary.BigEndian.AppendUint64(bytecode, 0)
	//
	// dec := opCodes[vm.OP_DECR]
	// bytecode = binary.BigEndian.AppendUint32(bytecode, uint32(dec))
	// bytecode = binary.BigEndian.AppendUint64(bytecode, 5)
	//
	// jmpg := opCodes[vm.OP_JMPG]
	// bytecode = binary.BigEndian.AppendUint32(bytecode, uint32(jmpg))
	// bytecode = binary.BigEndian.AppendUint64(bytecode, 1)
	//
	// // fmt.Println(len(bytecode))
	// fmt.Println(bytecode)
	// mach := vm.CreateVmState(1024)
	// err := mach.Execute(bytecode)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// var res any
	// mach.GetGpRXAs(vm.R2_IDX, vm.TY_FLOAT64, &res)
	// fmt.Printf("result: %f\n", res)

}
