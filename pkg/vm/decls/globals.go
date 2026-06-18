package decls
const (
	OPCODE_SIZE      = 4                           // opcode size in bytes (32-bits)
	ARGUMENT_SIZE    = 8                           // argument size in bytes (64-bits)
	INSTRUCTION_SIZE = OPCODE_SIZE + ARGUMENT_SIZE // size of a single instruction in bytes
)
