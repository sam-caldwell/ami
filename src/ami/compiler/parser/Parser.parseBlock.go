package parser

import (
    "fmt"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func (p *Parser) parseBlock() (*ast.BlockStmt, error) {
    if p.cur.Kind != token.LBraceSym {
        return nil, fmt.Errorf("expected '{', got %q", p.cur.Lexeme)
    }
    lb := p.cur.Pos
    // Tolerant block: consume until matching '}' with simple depth counter.
    depth := 0
    for {
        if p.cur.Kind == token.LBraceSym {
            depth++
        }
        p.next()
        if p.cur.Kind == token.RBraceSym {
            depth--
            if depth == 0 {
                rb := p.cur.Pos
                p.next()
                return &ast.BlockStmt{LBrace: lb, RBrace: rb}, nil
            }
        }
        // stop at EOF
        if p.cur.Kind == token.EOF {
            return &ast.BlockStmt{LBrace: lb}, nil
        }
    }
}

