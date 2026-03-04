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

func RedeclaredLabel(label, first, second string, line, col uint64) LexerError {
	err := LexerError {
		Line: line,
		Col: col,
		construct: func() string {
			return fmt.Sprintf("label '%s' redeclared at (%s), first declared at (%s)", label, first, second)
		},
	}
	return err
}

func UnresolvedLabel(label string) LexerError {
	err := LexerError {
		Line: 0,
		Col: 0,
		construct: func() string {
			return fmt.Sprintf("label '%s' unresolved", label)
		},
	}
	return err
}
