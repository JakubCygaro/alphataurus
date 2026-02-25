package assembler

import (
	"bufio"
)

type Parser struct {
	lexer Lexer
}
func NewParser(reader bufio.Reader) Parser {
	return Parser {
		lexer: NewLexer(reader),
	}
}
