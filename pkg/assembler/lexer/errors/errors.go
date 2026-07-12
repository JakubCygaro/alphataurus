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
	return fmt.
		Sprintf("(%v:%v): %s", e.Line, e.Col, e.construct())
}
func MakeLexerError(line, col int, format string, a ...any) LexerError {
	err := LexerError{
		Line: line,
		Col:  col,
		construct: func() string {
			return fmt.Sprintf(format, a...)
		},
	}
	return err
}
func UnrecognizedCharacter(line, col int, char rune) LexerError {
	return MakeLexerError(
		line,
		col,
		"Unrecognized character `%c`.",
		char,
	)
}
func UnclosedSingleQuote(line, col int) LexerError {
	return MakeLexerError(
		line,
		col,
		"Unclosed single-quote string.",
	)
}
func PrematureEndOfInput(line, col int) LexerError {
	return MakeLexerError(
		line,
		col,
		"Premature end of input.",
	)
}
func UnsupportedEscapeSequence(line, col int, seq rune) LexerError {
	return MakeLexerError(
		line,
		col,
		"Unsupported escape sequence `\\%c`",
		seq,
	)
}
func SingleQuoteNewline(line, col int) LexerError {
	return MakeLexerError(
		line,
		col,
		"Single-quote string broken by newline character.",
	)
}
func DigitLiteralTooLong(line, col int, lit string) LexerError {
	return MakeLexerError(
		line,
		col,
		"Digit literal is too long `%s`",
		lit,
	)
}
func MalformedFloatLiteral(line, col int, lit string) LexerError {
	return MakeLexerError(
		line,
		col,
		"Malformed float literal `%s`",
		lit,
	)
}
func MalformedIntegerLiteral(line, col int, lit string) LexerError {
	return MakeLexerError(
		line,
		col,
		"Malformed integer literal `%s`",
		lit,
	)
}
