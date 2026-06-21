package assembler

import "fmt"

func (a *Assembler) issueWarning(line, col int, format string, arg ...any) {
	if a.WarningSink != nil {
		a.WarningSink(
			AssemblerWarningData{
				Line: line,
				Col: col,
				Message: fmt.Sprintf(format, arg...),
			},
		)
	}
}

func (a* Assembler) parserWarningHandler(pwd ParserWarningData){
	a.issueWarning(pwd.Line, pwd.Col, "%s", pwd.Message)
}
