package parser

import "github.com/sam-caldwell/ami/src/ami/compiler/token"

// syncTop synchronizes the token stream to the next likely top-level boundary.
func (p *Parser) syncTop() {
    p.syncUntil(token.SemiSym, token.KwFunc, token.KwImport, token.KwPipeline, token.KwError, token.EOF)
}

