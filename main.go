package main

import (
	"encoding/binary"
	"fmt"

	"github.com/JakubCygaro/alphataurus/internal/vm"
)

func main() {
	bytecode := make([]byte, 0)
	opCodes := vm.GenerateOpcodeMap()

	mov := opCodes[vm.OP_MOVIR]
	//load in the loop register
	bytecode = binary.BigEndian.AppendUint32(bytecode, uint32(mov))
	bytecode[len(bytecode)-4] = 0b00000101 //r5
	bytecode = binary.BigEndian.AppendUint64(bytecode, 10)

	bytecode = binary.BigEndian.AppendUint32(bytecode, uint32(mov))
	bytecode[len(bytecode)-4] = 0b00000000 //r0
	bytecode = binary.BigEndian.AppendUint64(bytecode, 2)

	bytecode = binary.BigEndian.AppendUint32(bytecode, uint32(mov))
	bytecode[len(bytecode)-4] = 0b00000001 //r1
	bytecode = binary.BigEndian.AppendUint64(bytecode, 4)

	divrr := opCodes[vm.OP_DIVRR]
	bytecode = binary.BigEndian.AppendUint32(bytecode, uint32(divrr))
	bytecode = binary.BigEndian.AppendUint64(bytecode, 0x00000200000000)

	bytecode = binary.BigEndian.AppendUint32(bytecode, uint32(mov))
	bytecode[len(bytecode)-4] = 0b00100000 //r2 -> r0
	bytecode = binary.BigEndian.AppendUint64(bytecode, 4)

	cmp := opCodes[vm.OP_CMP]
	bytecode = binary.BigEndian.AppendUint32(bytecode, uint32(cmp))
	bytecode[len(bytecode)-4] = 0b11110101 //r5
	bytecode = binary.BigEndian.AppendUint64(bytecode, 0)

	dec := opCodes[vm.OP_DECR]
	bytecode = binary.BigEndian.AppendUint32(bytecode, uint32(dec))
	bytecode = binary.BigEndian.AppendUint64(bytecode, 5)

	jmpg := opCodes[vm.OP_JMPG]
	bytecode = binary.BigEndian.AppendUint32(bytecode, uint32(jmpg))
	bytecode = binary.BigEndian.AppendUint64(bytecode, 1)

	// fmt.Println(len(bytecode))
	fmt.Println(bytecode)
	mach := vm.CreateVmState(1024)
	err := mach.Execute(bytecode)
	if err != nil {
		fmt.Println(err)
	}
	var res any
	mach.GetGpRXAs(vm.R2_IDX, vm.TY_FLOAT64, &res)
	fmt.Printf("result: %f\n", res)

}
