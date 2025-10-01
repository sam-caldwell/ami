package parser

import (
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// collectPragmas scans the raw file content for lines beginning with '#pragma '
// and returns them as AST pragmas with 1-based line positions.
func (p *Parser) collectPragmas() []ast.Pragma {
    if p == nil || p.s == nil {
        return nil
    }
    content := p.sFileContent()
    if content == "" {
        return nil
    }
    var out []ast.Pragma
    line := 1
    start := 0
    for i := 0; i <= len(content); i++ {
        if i == len(content) || content[i] == '\n' {
            ln := content[start:i]
            if len(ln) >= 8 && ln[:8] == "#pragma " {
                text := ln[8:]
                pr := ast.Pragma{Pos: source.Position{Line: line, Column: 1, Offset: start}, Text: text}
                // parse schema: domain:key [args...]
                // split first space to get head and rest
                head := text
                rest := ""
                if sp := indexSpace(text); sp >= 0 {
                    head = text[:sp]
                    rest = strings.TrimSpace(text[sp+1:])
                }
                if c := strings.Index(head, ":"); c >= 0 {
                    pr.Domain = head[:c]
                    pr.Key = head[c+1:]
                } else {
                    pr.Domain = head
                }
                pr.Value = rest
                if rest != "" {
                    // tokenize by spaces (no quoted parsing for now)
                    fields := strings.Fields(rest)
                    pr.Args = append(pr.Args, fields...)
                    pr.Params = map[string]string{}
                    for _, tok := range fields {
                        if eq := strings.Index(tok, "="); eq > 0 {
                            k := tok[:eq]
                            v := tok[eq+1:]
                            // strip surrounding quotes on value when present
                            if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
                                v = v[1 : len(v)-1]
                            }
                            pr.Params[k] = v
                        }
                    }
                    if len(pr.Params) == 0 {
                        pr.Params = nil
                    }
                }
                out = append(out, pr)
            }
            line++
            start = i + 1
        }
    }
    return out
}

