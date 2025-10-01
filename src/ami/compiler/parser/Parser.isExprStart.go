package parser

import "github.com/sam-caldwell/ami/src/ami/compiler/token"

func (p *Parser) isExprStart(k token.Kind) bool {
    switch k {
    case token.Ident, token.Number, token.String, token.KwSlice, token.KwSet, token.KwMap,
        token.Bang, token.Minus, token.TildeSym, token.LParenSym:
        return true
    default:
        return false
    }
}

