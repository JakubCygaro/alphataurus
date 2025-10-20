package vm;

const (
	OPCODE_SIZE = 4 // opcode size in bytes (32-bits)
	ARGUMENT_SIZE = 8 // argument size in bytes (64-bits)
	INSTRUCTION_SIZE = OPCODE_SIZE + ARGUMENT_SIZE // size of a single instruction in bytes


	OP_MOVIR0 = 0x00_01_00_00 // move imediate value to register 0
	OP_MOVIR1 = 0x00_01_00_01 // move imediate value to register 0

	OP_ADDUR0R1 = 0x00_02_00_00 // add register to register and store into second register UNSINGED
	OP_ADDSR0R1 = 0x00_02_00_01 // add register to register and store into second register SINGED


	OP_JMP = 0x00_10_00_00 // jump to instruction
	OP_JMPE = 0x00_10_00_01 // jump if equal
	OP_TESTR0R1 = 0x00_11_00_00 // test registers
)
