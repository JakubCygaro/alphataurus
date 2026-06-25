package assembler

import "fmt"

func (p *Parser) issueWarning(line, col int, format string, a ...any) {
	if p.WarningSink != nil {
		p.WarningSink(
			ParserWarningData{
				Line: line,
				Col: col,
				Message: fmt.Sprintf(format, a...),
			},
		)
	}
}
