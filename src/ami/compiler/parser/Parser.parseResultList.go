package parser

import (
    "fmt"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func (p *Parser) parseResultList() ([]ast.Result, source.Position, source.Position, error) {
    // Optional tuple of results in parentheses
    if p.cur.Kind != token.LParenSym {
        return nil, source.Position{}, source.Position{}, nil
    }
    lp := p.cur.Pos
    p.next()
    var results []ast.Result
    for p.cur.Kind != token.RParenSym && p.cur.Kind != token.EOF {
        if !p.isTypeName(p.cur.Kind) {
            return nil, lp, source.Position{}, fmt.Errorf("expected result ident, got %q", p.cur.Lexeme)
        }
        rpos := p.cur.Pos
        rtype := p.cur.Lexeme
        rtypePos := rpos
        p.next()
        // Extend with cross-package selectors like pkg.Type
        rtype = p.captureQualifiedType(rtype)
        // Capture Struct{...} results
        if rtype == "Struct" && p.cur.Kind == token.LBraceSym {
            startOff := rtypePos.Offset
            depthBrace := 0
            depthAngle := 0
            var lastBraceOff int
            for {
                if p.cur.Kind == token.EOF {
                    break
                }
                switch p.cur.Kind {
                case token.LBraceSym:
                    depthBrace++
                case token.RBraceSym:
                    if depthBrace > 0 {
                        depthBrace--
                    }
                    lastBraceOff = p.cur.Pos.Offset
                    p.next()
                    if depthBrace == 0 {
                        goto doneResultStruct
                    }
                    continue
                case token.Lt:
                    depthAngle++
                case token.Shl:
                    depthAngle += 2
                case token.Gt:
                    if depthAngle > 0 {
                        depthAngle--
                    }
                case token.Shr:
                    if depthAngle >= 2 {
                        depthAngle -= 2
                    }
                }
                p.next()
            }
        doneResultStruct:
            src := p.s.FileContent()
            if lastBraceOff > startOff && lastBraceOff+1 <= len(src) {
                rtype = src[startOff : lastBraceOff+1]
            }
        }
        if p.cur.Kind == token.Lt || p.cur.Kind == token.Shl {
            startOff := rpos.Offset
            depth := 0
            var lastGtOff int
            for {
                if p.cur.Kind == token.EOF {
                    break
                }
                switch p.cur.Kind {
                case token.Lt:
                    depth++
                case token.Shl:
                    depth += 2
                case token.Gt:
                    depth--
                    lastGtOff = p.cur.Pos.Offset
                    p.next()
                    if depth == 0 {
                        goto doneResultGeneric
                    }
                    continue
                case token.Shr:
                    depth -= 2
                    lastGtOff = p.cur.Pos.Offset + 1
                    p.next()
                    if depth == 0 {
                        goto doneResultGeneric
                    }
                    continue
                }
                p.next()
            }
        doneResultGeneric:
            src := p.s.FileContent()
            if lastGtOff > startOff && lastGtOff+1 <= len(src) {
                rtype = src[startOff : lastGtOff+1]
            }
        }
        results = append(results, ast.Result{Pos: rpos, Type: rtype, TypePos: rtypePos})
        if p.cur.Kind == token.CommaSym {
            p.next()
            continue
        }
        if p.cur.Kind != token.RParenSym {
            return nil, lp, source.Position{}, fmt.Errorf("expected ',' or ')', got %q", p.cur.Lexeme)
        }
    }
    rp := p.cur.Pos
    if p.cur.Kind == token.RParenSym {
        p.next()
    }
    return results, lp, rp, nil
}
