package errors

import (
	"fmt"
)

type constructMessage func() string

type AssemblerError struct {
	Line      int
	Col       int
	construct constructMessage
}

func (e AssemblerError) Error() string {
	return fmt.
		Sprintf("(%v:%v): %s", e.Line, e.Col, e.construct())
}
func MakeAssemblerError(line, col int, format string, a... any) AssemblerError {
	err := AssemblerError{
		Line: line,
		Col:  col,
		construct: func() string {
			return fmt.Sprintf(format, a...)
		},
	}
	return err
}
