package errors

import (
	"fmt"
)

type ParserError struct {
	ParsingWhat string
	Line, Col   int
	construct   constructMessage
}

func (e ParserError) Error() string {
	if e.ParsingWhat != "" {
		return fmt.Sprintf("Failed to parse %s at (%v:%v): %s", e.ParsingWhat, e.Line, e.Col, e.construct())
	} else {
		return fmt.Sprintf("Parsing error (%v:%v): %s", e.Line, e.Col, e.construct())
	}
}

func ExtraTokensOnLine(line, col int) ParserError {
	err := ParserError{
		Line: line,
		Col:  col,
		construct: func() string {
			return "Extra tokens on line"
		},
	}
	return err
}

func UnknownIdentifier(ident string, line, col int) ParserError {
	err := ParserError{
		Line: line,
		Col:  col,
		construct: func() string {
			return fmt.Sprintf("Unknown identifier '%s'", ident)
		},
	}
	return err
}

func FailedToParse(instruction, reason string, line, col int) ParserError {
	err := ParserError{
		ParsingWhat: instruction,
		Line:        line,
		Col:         col,
		construct: func() string {
			return fmt.Sprintf("%s", reason)
		},
	}
	return err
}

func PrematureEndOfInput(line, col int) ParserError {
	err := ParserError{
		Line: line,
		Col:  col,
		construct: func() string {
			return "Premature end of input"
		},
	}
	return err
}
func UnclosedSingleQuote(line, col int) ParserError {
	err := ParserError{
		Line: line,
		Col:  col,
		construct: func() string {
			return "Unclosed single quote"
		},
	}
	return err
}
func UnclosedParen(line, col int) ParserError {
	err := ParserError{
		Line: line,
		Col:  col,
		construct: func() string {
			return "Unclosed parentheses"
		},
	}
	return err
}
func BadSizeArgument(given, needed string, line, col int) ParserError {
	err := ParserError{
		Line: line,
		Col:  col,
		construct: func() string {
			return fmt.Sprintf("Bad data size argument %s - data size of %s is needed",
				given, needed)
		},
	}
	return err
}
func MismatchedRegisterSizes(line, col int) ParserError {
	err := ParserError{
		Line: line,
		Col:  col,
		construct: func() string {
			return "Mismatched register sizes"
		},
	}
	return err
}
func BadRegisterSize(line, col int) ParserError {
	err := ParserError{
		Line: line,
		Col:  col,
		construct: func() string {
			return "Mismatched register sizes"
		},
	}
	return err
}
func MissingDataSize(line, col int) ParserError {
	err := ParserError{
		Line: line,
		Col:  col,
		construct: func() string {
			return "Missing data size"
		},
	}
	return err
}
func UnnecessarySizeParameter(param string, line, col int) ParserError {
	err := ParserError{
		Line: line,
		Col:  col,
		construct: func() string {
			return fmt.Sprintf("Unnecessary size parameter %s",
				param)
		},
	}
	return err
}
