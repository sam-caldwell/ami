package parser

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func (p *Parser) parseErrorBlock() (*ast.ErrorBlock, error) {
    pos := p.cur.Pos
    p.next()
    body, err := p.parseStepBlock()
    if err != nil {
        return nil, err
    }
    eb := &ast.ErrorBlock{Pos: pos, Body: body, Leading: p.pending}
    p.pending = nil
    return eb, nil
}

