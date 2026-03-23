package errors

import (
	"fmt"
)

type constructMessage func() string

type AlphaVMError struct {
	Type      int
	Pos       uint64
	construct constructMessage
}

func (e AlphaVMError) Error() string {
	return fmt.Sprintf("Execution error [%d] (0x%08x): %s", e.Type, e.Pos, e.construct())
}

func DisallowedSrcRegister(reg int, pos uint64) AlphaVMError {
	err := AlphaVMError{
		Type: ERR_DISALLOWED_SOURCE_REGISTER,
		Pos: pos,
		construct: func() string {
			return fmt.Sprintf("Disallowed source register %d", reg)
		},
	}
	return err
}
func DisallowedDestRegister(reg int, pos uint64) AlphaVMError {
	err := AlphaVMError{
		Type: ERR_DISALLOWED_DESTINATION_REGISTER,
		Pos: pos,
		construct: func() string {
			return fmt.Sprintf("Disallowed destination register %d", reg)
		},
	}
	return err
}
func BadCodeSectionSize() AlphaVMError {
	err := AlphaVMError{
		Type: ERR_BAD_CODE_SEC_SIZE,
		Pos: 0x0,
		construct: func() string {
			return "Bad code section size"
		},
	}
	return err
}
func BadArthmeticOperation(pos uint64) AlphaVMError {
	err := AlphaVMError{
		Type: ERR_ARTH_EXCEPTION,
		Pos: pos,
		construct: func() string {
			return "Arthmetic exception"
		},
	}
	return err
}
func BadOpcode(opcode uint32, pos uint64) AlphaVMError {
	err := AlphaVMError{
		Type: ERR_BAD_OPCODE,
		Pos: pos,
		construct: func() string {
			return fmt.Sprintf("Bad opcode (0x%08x)", opcode)
		},
	}
	return err
}
func DisallowedOp1Register(reg int, pos uint64) AlphaVMError {
	err := AlphaVMError{
		Type: ERR_DISALLOWED_OPERAND_1_REGISTER,
		Pos: pos,
		construct: func() string {
			return fmt.Sprintf("Disallowed first operand register (%d)", reg)
		},
	}
	return err
}
func DisallowedOp2Register(reg int, pos uint64) AlphaVMError {
	err := AlphaVMError{
		Type: ERR_DISALLOWED_OPERAND_2_REGISTER,
		Pos: pos,
		construct: func() string {
			return fmt.Sprintf("Disallowed second operand register (%d)", reg)
		},
	}
	return err
}
func StackUnderflow(pos uint64) AlphaVMError {
	err := AlphaVMError{
		Type: ERR_STACK_UNDERFLOW,
		Pos: pos,
		construct: func() string {
			return "Stack underflow"
		},
	}
	return err
}
func StackOverflow(pos uint64) AlphaVMError {
	err := AlphaVMError{
		Type: ERR_STACK_UNDERFLOW,
		Pos: pos,
		construct: func() string {
			return "Stack overflow"
		},
	}
	return err
}
func SegmentationFault(address uint64, pos uint64) AlphaVMError {
	err := AlphaVMError{
		Type: ERR_DISALLOWED_DESTINATION_REGISTER,
		Pos: pos,
		construct: func() string {
			return fmt.Sprintf("Segmentation fault, address (0x%08x)", address)
		},
	}
	return err
}
