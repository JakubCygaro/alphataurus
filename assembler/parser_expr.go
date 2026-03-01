package assembler

import "fmt"
func (p *Parser) parseExpression(arthTy int) (Expr, error) {
	if err := p.lexer.ReadNextToken(); err != nil {
		return Expr{}, err
	}
	start := p.lexer.CurrentToken()
	switch start.Ty {
	case TOKEN_TINTEGER_LIT
	default:
		return Expr{}, fmt.Errorf("Not a valid expression %s", p.lexer.CurrentPosition())
	}
}
