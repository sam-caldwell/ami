package parser

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// parseStepArgText parses a single step argument into Arg text, tolerating
// qualified calls and simple literals without requiring full expression parsing.
func (p *Parser) parseStepArgText() (ast.Arg, bool) {
    // Key=value form
    if p.cur.Kind == token.Ident || p.cur.Kind == token.KwType {
        key := p.cur.Lexeme
        pos := p.cur.Pos
        p.next()
        if p.cur.Kind == token.Assign {
            p.next()
            // Bare type/keywords
            if p.isTypeName(p.cur.Kind) || p.cur.Kind == token.KwErrorPipeline || p.cur.Kind == token.KwErrorEvent || p.cur.Kind == token.KwError || p.cur.Kind == token.KwTrue || p.cur.Kind == token.KwFalse {
                val := p.cur.Lexeme
                p.next()
                return ast.Arg{Pos: pos, Text: key + "=" + val}, true
            }
            // Qualified call: Ident(.Ident)*( ... )
            if p.cur.Kind == token.Ident {
                name := p.cur.Lexeme
                p.next()
                for p.cur.Kind == token.DotSym {
                    p.next()
                    if p.cur.Kind != token.Ident { break }
                    name += "." + p.cur.Lexeme
                    p.next()
                }
                if p.cur.Kind == token.LParenSym {
                    // consume balanced parens
                    p.next()
                    depth := 1
                    for depth > 0 && p.cur.Kind != token.EOF {
                        switch p.cur.Kind {
                        case token.LParenSym:
                            depth++
                        case token.RParenSym:
                            depth--
                        }
                        p.next()
                    }
                    return ast.Arg{Pos: pos, Text: key + "=" + name + "(…)"}, true
                }
                return ast.Arg{Pos: pos, Text: key + "=" + name}, true
            }
            // List literal: [ ... ]
            if p.cur.Kind == token.LBracketSym {
                start := p.cur.Pos.Offset
                depth := 0
                var last int
                for {
                    if p.cur.Kind == token.EOF { break }
                    if p.cur.Kind == token.LBracketSym { depth++ }
                    if p.cur.Kind == token.RBracketSym {
                        if depth > 0 { depth-- }
                        last = p.cur.Pos.Offset
                        p.next()
                        if depth == 0 { break }
                        continue
                    }
                    p.next()
                }
                src := p.s.FileContent()
                val := "[]"
                if last > start && last+1 <= len(src) {
                    val = src[start : last+1]
                }
                return ast.Arg{Pos: pos, Text: key + "=" + val}, true
            }
            // Fallback: consume until comma or ')' minimally
            for p.cur.Kind != token.CommaSym && p.cur.Kind != token.RParenSym && p.cur.Kind != token.EOF {
                p.next()
            }
            return ast.Arg{Pos: pos, Text: key + "="}, true
        }
        // Not key=value; treat name as simple arg expression name
        return ast.Arg{Pos: pos, Text: key}, true
    }
    // Fallback: try generic expression
    e, ok := p.parseExprPrec(1)
    if ok {
        return ast.Arg{Pos: ePos(e), Text: exprText(e)}, true
    }
    return ast.Arg{}, false
}

