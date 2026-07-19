package errors

import (
	"fmt"
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
)

type constructMessage func() string

type ParserError struct {
	Line, Col int
	construct constructMessage
}

func (e ParserError) Error() string {
	return fmt.
		Sprintf("(%v:%v): %s", e.Line, e.Col, e.construct())
}

func MakeParserError(line, col int, format string, a ...any) ParserError {
	err := ParserError{
		Line: line,
		Col:  col,
		construct: func() string {
			return fmt.
				Sprintf(format, a...)
		},
	}
	return err
}
func DisallowedDest(line, col int, dest string) ParserError {
	return MakeParserError(
		line,
		col,
		"Disallowed destination `%s`.",
		dest,
	)
}
func DisallowedDestReg(line, col int, reg lx.RegisterData) ParserError {
	return MakeParserError(
		line,
		col,
		"Disallowed destination register `%s`.",
		reg.String(),
	)
}
func DisallowedSrcReg(line, col int, reg lx.RegisterData) ParserError {
	return MakeParserError(
		line,
		col,
		"Disallowed source register `%s`.",
		reg.String(),
	)
}
func DisallowedSrc(line, col int, src string) ParserError {
	return MakeParserError(
		line,
		col,
		"Disallowed source `%s`.",
		src,
	)
}
func UnnecessarySizeParameter(line, col int, size byte) ParserError {
	sz, _ := lx.GetSizeKeyword(size)
	return MakeParserError(
		line,
		col,
		"Unnecessary size parameter `%s`.",
		sz,
	)
}
func MissingSizeParameter(line, col int) ParserError {
	return MakeParserError(
		line,
		col,
		"Unnecessary size parameter.",
	)
}
// line1 and col1 will be used as the location of the error
func MismatchedRegisterSizes(
	line1, col1, line2, col2 int,
	reg1, reg2 lx.RegisterData,
) ParserError {
	return MakeParserError(
		line1,
		col2,
		"Mismatched register sizes `%s` and `%s`.",
		reg1, reg2,
	)
}
func ExtraTokensOnLine(
	line, col int,
	token lx.Token,
) ParserError {
	return MakeParserError(
		line,
		col,
		"Extra tokens on the line `%s`.",
		token.ForceValAsString(),
	)
}
func UnknownIdentifier(
	line, col int,
	iden string,
) ParserError {
	return MakeParserError(
		line,
		col,
		"Unknown identifier `%s`.",
		iden,
	)
}
func MissingComma(
	line, col int,
	got lx.Token,
) ParserError {
	return MakeParserError(
		line,
		col,
		"Instruction missing a comma, got `%s` instead.",
		got.ForceValAsString(),
	)
}
func PrematureEndOfInput(
	line, col int,
) ParserError {
	return MakeParserError(
		line,
		col,
		"Premature end of input.",
	)
}
func UnclosedParen(
	line, col int,
) ParserError {
	return MakeParserError(
		line,
		col,
		"Unclosed parentheses.",
	)
}
