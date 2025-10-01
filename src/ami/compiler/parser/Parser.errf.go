package parser

import "fmt"

func (p *Parser) errf(format string, args ...any) {
    // capture current token position for diagnostics
    se := SyntaxError{Msg: fmt.Sprintf(format, args...), Pos: p.cur.Pos}
    p.errors = append(p.errors, se)
}

