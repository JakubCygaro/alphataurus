package assembler

import (
	"fmt"

	pe "github.com/JakubCygaro/alphataurus/pkg/assembler/parser/errors"
)

// this function takes in a string representation of an expression
// or an error that happend during the production of said string
// and uses it to create a ParserError, via the user provided
// errMake parameter
type MakeErrWithExprFn func(string) pe.ParserError

func MakeParserErrorWithExpr(
	expr *Expr,
	errMake MakeErrWithExprFn,
) pe.ParserError {

	if em, err := expr.Emit(); err != nil {
		return errMake(
			fmt.Sprintf("(EXPRESSION EMIT ERROR: %s)", err.Error()),
		)
	} else {
		return errMake(em)
	}
}
