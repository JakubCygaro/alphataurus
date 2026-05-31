package errors

import (
	"fmt"
)

type AssemblerError struct {
	Line, Col uint64
	construct constructMessage
}

func (e AssemblerError) Error() string {
	return fmt.Sprintf("Assembling error (%v:%v): %s", e.Line, e.Col, e.construct())
}

func RedeclaredLabel(label, first string, line, col uint64) AssemblerError {
	err := AssemblerError{
		Line: line,
		Col:  col,
		construct: func() string {
			return fmt.Sprintf("label '%s' redeclared at (%s)", label, first)
		},
	}
	return err
}
func MultipleEntry(line, col uint64) AssemblerError {
	err := AssemblerError{
		Line: line,
		Col:  col,
		construct: func() string {
			return "multiple entry points defined"
		},
	}
	return err
}
func ImportedSymbolDeclared(name string, line, col uint64) AssemblerError {
	err := AssemblerError{
		Line: line,
		Col:  col,
		construct: func() string {
			return fmt.Sprintf("imported symbol '%s' declared at (%v:%v)", name, line, col)
		},
	}
	return err
}

func UnresolvedLabel(label string) AssemblerError {
	err := AssemblerError{
		Line: 0,
		Col:  0,
		construct: func() string {
			return fmt.Sprintf("label '%s' unresolved", label)
		},
	}
	return err
}
func UnresolvedSymbol(label string) AssemblerError {
	err := AssemblerError{
		Line: 0,
		Col:  0,
		construct: func() string {
			return fmt.Sprintf("symbol '%s' is unresolved", label)
		},
	}
	return err
}

func MultipleSymbolDefinitions(name string, line, col uint64) AssemblerError {
	err := AssemblerError{
		Line: line,
		Col:  col,
		construct: func() string {
			return fmt.Sprintf("symbol '%s' redefined", name)
		},
	}
	return err
}
func DisallowedTopLevelInstruction(line, col uint64) AssemblerError {
	err := AssemblerError{
		Line: line,
		Col:  col,
		construct: func() string {
			return "Disallowed top level instruction"
		},
	}
	return err
}
func DisallowedDestinationRegister(line, col uint64) AssemblerError {
	err := AssemblerError{
		Line: line,
		Col:  col,
		construct: func() string {
			return "Disallowed destination register"
		},
	}
	return err
}
func DisallowedSourceRegister(line, col uint64) AssemblerError {
	err := AssemblerError{
		Line: line,
		Col:  col,
		construct: func() string {
			return "Disallowed source register"
		},
	}
	return err
}
