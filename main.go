package main

import (
	"encoding/binary"
	"fmt"

	"github.com/JakubCygaro/alphataurus/internal/vm"
)

func main() {
	bytecode := make([]byte, 0)

	bytecode = binary.BigEndian.AppendUint32(bytecode, vm.OP_MOVIR0)
	bytecode = binary.BigEndian.AppendUint64(bytecode, 69)

	bytecode = binary.BigEndian.AppendUint32(bytecode, vm.OP_MOVIR1)
	bytecode = binary.BigEndian.AppendUint64(bytecode, 400)

	bytecode = binary.BigEndian.AppendUint32(bytecode, vm.OP_ADDSR0R1)
	bytecode = binary.BigEndian.AppendUint64(bytecode, 1)

	bytecode = binary.BigEndian.AppendUint32(bytecode, vm.OP_JMP)
	bytecode = binary.BigEndian.AppendUint64(bytecode, 0)

	fmt.Println(len(bytecode))
	fmt.Println(bytecode)
	vm := vm.CreateVmState(1024)
	err := vm.Execute(bytecode)
	if err != nil {
		fmt.Println(err)
	}

}
