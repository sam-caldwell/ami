package parser

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// advance reads next token, collecting comments into pending and stopping at the
// first non-comment token (or EOF).
func (p *Parser) advance() {
    for {
        p.cur = p.s.Next()
        if p.cur.Kind == token.LineComment || p.cur.Kind == token.BlockComment {
            p.pending = append(p.pending, ast.Comment{Pos: p.cur.Pos, Text: p.cur.Lexeme})
            continue
        }
        return
    }
}

