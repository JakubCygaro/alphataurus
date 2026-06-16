package errors

import (
	"fmt"
)

type constructMessage func() string

type LexerError struct {
	Line, Col int
	construct constructMessage
}

func (e LexerError) Error() string {
	return fmt.Sprintf("Lexer error at (%v:%v): %s", e.Line, e.Col, e.construct())
}
func UnrecognizedChar(char rune, line, col int) LexerError {
	err := LexerError{
		Line: line,
		Col:  col,
		construct: func() string {
			return fmt.Sprintf("Unrecognized character `%c`", char)
		},
	}
	return err
}

func MalformedIntegerLit(line, col int) LexerError {
	err := LexerError{
		Line: line,
		Col:  col,
		construct: func() string {
			return "Malformed integer number literal"
		},
	}
	return err
}
func MalformedFloatLit(line, col int) LexerError {
	err := LexerError{
		Line: line,
		Col:  col,
		construct: func() string {
			return "Malformed floating point number literal"
		},
	}
	return err
}
