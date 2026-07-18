package errors

import (
	"fmt"
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
)

type constructMessage func() string

type AssemblerError struct {
	Line      int
	Col       int
	construct constructMessage
}

func (e AssemblerError) Error() string {
	if e.Line == 0 {
		return fmt.
			Sprintf("%s", e.construct())
	} else {
		return fmt.
			Sprintf("(%v:%v): %s", e.Line, e.Col, e.construct())
	}
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
func BadRegisterSize(line, col int, r lx.RegisterData) AssemblerError {
	return MakeAssemblerError(
		line,
		col,
		"Bad register size `%s`.",
		r.String(),
	)
}
func LabelRedeclared(
	line, col int,
	labelName string,
	declaredAtLine, declaredAtCol int,
) AssemblerError {
	return MakeAssemblerError(
		line,
		col,
		"Label `%s` redeclared, originally declared at (%v:%v).",
		labelName,
		declaredAtLine, declaredAtCol,
	)
}
