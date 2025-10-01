package parser

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// parseDecorator parses a decorator starting at '@'.
func (p *Parser) parseDecorator() (ast.Decorator, bool) {
    if p.cur.Kind != token.AtSym {
        return ast.Decorator{}, false
    }
    atPos := p.cur.Pos
    p.next()
    if p.cur.Kind != token.Ident {
        p.errf("expected decorator name after '@', got %q", p.cur.Lexeme)
        return ast.Decorator{Pos: atPos}, false
    }
    name := p.cur.Lexeme
    namePos := p.cur.Pos
    p.next()
    // Optional arg list
    var lparen, rparen source.Position
    var args []ast.Expr
    if p.cur.Kind == token.LParenSym {
        lparen = p.cur.Pos
        p.next()
        for p.cur.Kind != token.RParenSym && p.cur.Kind != token.EOF {
            e, ok := p.parseExprPrec(1)
            if ok {
                args = append(args, e)
            } else {
                p.errf("unexpected token in decorator args: %q", p.cur.Lexeme)
                p.syncUntil(token.CommaSym, token.RParenSym)
            }
            if p.cur.Kind == token.CommaSym {
                p.next()
                continue
            }
        }
        rparen = p.cur.Pos
        if p.cur.Kind == token.RParenSym {
            p.next()
        } else {
            p.errf("missing ')' to close decorator args")
        }
    }
    return ast.Decorator{Pos: atPos, NamePos: namePos, Name: name, LParen: lparen, Args: args, RParen: rparen}, true
}

