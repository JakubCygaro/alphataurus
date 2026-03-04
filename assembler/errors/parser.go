package errors
import (
	"fmt"
)
type ParserError struct {
	ParsingWhat string
	Line, Col uint64
	construct constructMessage
}

func (e ParserError) Error() string {
	if e.ParsingWhat != "" {
		return fmt.Sprintf("Failed to parse %s at (%v:%v): %s", e.ParsingWhat, e.Line, e.Col, e.construct())
	} else {
		return fmt.Sprintf("Parsing error (%v:%v): %s", e.Line, e.Col, e.construct())
	}
}

func ExtraTokensOnLine(line, col uint64) ParserError {
	err := ParserError {
		Line: line,
		Col: col,
		construct: func() string {
			return "Extra tokens on line"
		},
	}
	return err
}

func UnknownIdentifier(ident string, line, col uint64) ParserError {
	err := ParserError {
		Line: line,
		Col: col,
		construct: func() string {
			return fmt.Sprintf("Unknown identifier '%s'", ident)
		},
	}
	return err
}

func FailedToParse(instruction, reason string, line, col uint64) ParserError {
	err := ParserError {
		ParsingWhat: instruction,
		Line: line,
		Col: col,
		construct: func() string {
			return fmt.Sprintf("%s", reason)
		},
	}
	return err
}

func PrematureEndOfInput(line, col uint64) ParserError {
	err := ParserError {
		Line: line,
		Col: col,
		construct: func() string {
			return "Premature end of input"
		},
	}
	return err
}
func UnclosedParen(line, col uint64) ParserError {
	err := ParserError {
		Line: line,
		Col: col,
		construct: func() string {
			return "Unclosed parentheses"
		},
	}
	return err
}
