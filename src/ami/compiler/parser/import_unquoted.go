package parser

import (
    "strings"

    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// parseImportUnquoted parses an unquoted import path optionally followed by a constraint.
// It accepts module paths like github.com/org/repo and ami/stdlib/io and a constraint
// in the form ">= vX.Y.Z" (operator and version may be separated by spaces).
func (p *Parser) parseImportUnquoted() (string, string, bool) {
    return p.parseImportUnquotedWithPrefix("")
}

func (p *Parser) parseImportUnquotedWithPrefix(prefix string) (string, string, bool) {
    var b strings.Builder
    if prefix != "" {
        b.WriteString(prefix)
    } else {
        if !(p.cur.Kind == tok.IDENT) {
            return "", "", false
        }
        b.WriteString(p.cur.Lexeme)
        p.next()
    }
    // accumulate path segments
    for {
        switch p.cur.Kind {
        case tok.DOT:
            b.WriteByte('.')
            p.next()
            if p.cur.Kind == tok.IDENT {
                b.WriteString(p.cur.Lexeme)
                p.next()
            }
            continue
        case tok.SLASH:
            b.WriteByte('/')
            p.next()
            if p.cur.Kind == tok.IDENT {
                b.WriteString(p.cur.Lexeme)
                p.next()
            }
            continue
        case tok.MINUS:
            // minus inside a segment
            b.WriteByte('-')
            p.next()
            if p.cur.Kind == tok.IDENT {
                b.WriteString(p.cur.Lexeme)
                p.next()
            }
            continue
        default:
        }
        break
    }
    path := b.String()
    // optional constraint: read rest of line and accept forms starting with ">="
    cons := ""
    {
        src := p.s.Source()
        start := p.cur.Offset
        end := start
        for end < len(src) {
            if src[end] == '\n' {
                break
            }
            end++
        }
        tail := strings.TrimSpace(src[start:end])
        if strings.HasPrefix(tail, ">=") {
            cons = ">=" + " " + strings.TrimSpace(strings.TrimPrefix(tail, ">="))
            // advance tokens to the end of line to keep parser in sync
            for p.cur.Kind != tok.EOF && p.cur.Offset < end {
                p.next()
            }
        }
    }
    if path == "" {
        return "", "", false
    }
    return path, cons, true
}

