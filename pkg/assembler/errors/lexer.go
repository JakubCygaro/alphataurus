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

func MalformedIntegerLit(line, col int, lit string) LexerError {
	err := LexerError{
		Line: line,
		Col:  col,
		construct: func() string {
			return fmt.Sprintf("Malformed integer number literal `%s`", lit)
		},
	}
	return err
}
func MalformedFloatLit(line, col int, lit string) LexerError {
	err := LexerError{
		Line: line,
		Col:  col,
		construct: func() string {
			return fmt.Sprintf("Malformed floating point number literal `%s`", lit)
		},
	}
	return err
}
func DigitLiteralTooLong(line, col int, lit string) LexerError {
	err := LexerError{
		Line: line,
		Col:  col,
		construct: func() string {
			return fmt.Sprintf("Digit literal is too big `%s`", lit)
		},
	}
	return err
}
func LUnclosedSingleQuote(line, col int) LexerError {
	err := LexerError{
		Line: line,
		Col:  col,
		construct: func() string {
			return "Unclosed single quote"
		},
	}
	return err
}
func LSingleQuoteNewline(line, col int) LexerError {
	err := LexerError{
		Line: line,
		Col:  col,
		construct: func() string {
			return "Single quote broken by a newline"
		},
	}
	return err
}
func UnsupportedEscape(line, col int, ch rune) LexerError {
	err := LexerError{
		Line: line,
		Col:  col,
		construct: func() string {
			return fmt.Sprintf("Unsupported escape sequence `%c`", ch)
		},
	}
	return err
}
func LPrematureEndOfInput(line, col int) LexerError {
	err := LexerError{
		Line: line,
		Col:  col,
		construct: func() string {
			return "Premature end of input"
		},
	}
	return err
}
