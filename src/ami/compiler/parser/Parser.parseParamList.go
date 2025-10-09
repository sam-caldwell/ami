package parser

import (
    "fmt"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func (p *Parser) parseParamList() ([]ast.Param, source.Position, source.Position, error) {
    if p.cur.Kind != token.LParenSym {
        return nil, source.Position{}, source.Position{}, fmt.Errorf("expected '(', got %q", p.cur.Lexeme)
    }
    lp := p.cur.Pos
    p.next()
    var params []ast.Param
    var pend []ast.Comment
    for p.cur.Kind != token.RParenSym && p.cur.Kind != token.EOF {
        if p.cur.Kind == token.LineComment || p.cur.Kind == token.BlockComment {
            pend = append(pend, ast.Comment{Pos: p.cur.Pos, Text: p.cur.Lexeme})
            p.next()
            continue
        }
        if p.cur.Kind != token.Ident {
            return nil, lp, source.Position{}, fmt.Errorf("expected param name, got %q", p.cur.Lexeme)
        }
        nameTok := p.cur
        p.next()
        var typ string
        var typePos source.Position
        if p.isTypeName(p.cur.Kind) {
            typePos = p.cur.Pos
            typ = p.cur.Lexeme
            p.next()
            // Extend with cross-package selectors like pkg.Type
            typ = p.captureQualifiedType(typ)
            // Capture Struct{...} types fully
            if typ == "Struct" && p.cur.Kind == token.LBraceSym {
                startOff := typePos.Offset
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
                            goto doneParamStruct
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
            doneParamStruct:
                src := p.s.FileContent()
                if lastBraceOff > startOff && lastBraceOff+1 <= len(src) {
                    typ = src[startOff : lastBraceOff+1]
                }
            }
            // If generic arguments follow, capture full "Base<...>" from source
            if p.cur.Kind == token.Lt || p.cur.Kind == token.Shl {
                startOff := typePos.Offset
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
                            goto doneParamGeneric
                        }
                        continue
                    case token.Shr:
                        // token position is at first '>' of '>>'; set last to second '>'
                        depth -= 2
                        lastGtOff = p.cur.Pos.Offset + 1
                        p.next()
                        if depth == 0 {
                            goto doneParamGeneric
                        }
                        continue
                    }
                    p.next()
                }
            doneParamGeneric:
                src := p.s.FileContent()
                if lastGtOff > startOff && lastGtOff+1 <= len(src) {
                    typ = src[startOff : lastGtOff+1]
                }
            }
        }
        params = append(params, ast.Param{Name: nameTok.Lexeme, Pos: nameTok.Pos, Type: typ, TypePos: typePos, Leading: pend})
        pend = nil
        if p.cur.Kind == token.CommaSym {
            p.next()
            continue
        }
        if p.cur.Kind != token.RParenSym {
            return nil, lp, source.Position{}, fmt.Errorf("expected ',' or ')', got %q", p.cur.Lexeme)
        }
    }
    rp := p.cur.Pos
    // consume ')'
    if p.cur.Kind == token.RParenSym {
        p.next()
    }
    return params, lp, rp, nil
}
