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
	bytecode = binary.BigEndian.AppendUint32(bytecode, uint32(mov))
	bytecode[0] = 0b00000000 //r0
	bytecode = binary.BigEndian.AppendUint64(bytecode, 2)

	bytecode = binary.BigEndian.AppendUint32(bytecode, uint32(mov))
	bytecode[len(bytecode)-4] = 0b00000001 //r1
	bytecode = binary.BigEndian.AppendUint64(bytecode, 4)

	mulrr := opCodes[vm.OP_MULRR]
	bytecode = binary.BigEndian.AppendUint32(bytecode, uint32(mulrr))
	bytecode = binary.BigEndian.AppendUint64(bytecode, 0x0)

	jmp := opCodes[vm.OP_JMP]
	bytecode = binary.BigEndian.AppendUint32(bytecode, uint32(jmp))
	bytecode = binary.BigEndian.AppendUint64(bytecode, 10)

	fmt.Println(len(bytecode))
	fmt.Println(bytecode)
	vm := vm.CreateVmState(1024)
	err := vm.Execute(bytecode)
	if err != nil {
		fmt.Println(err)
	}

}
