package parser

import "github.com/sam-caldwell/ami/src/ami/compiler/token"

// parseImportConstraint parses an optional version constraint that follows an import path.
// Supported form (scaffold): ">= vMAJOR.MINOR.PATCH[-PRERELEASE[.N]]" with or without spaces.
// Returns the canonical string (e.g., ">= v1.2.3-rc.1") or empty string when none.
func (p *Parser) parseImportConstraint() string {
    if p.cur.Kind != token.Ge {
        return ""
    }
    // capture operator
    op := p.cur.Lexeme
    if op == "" {
        op = ">="
    }
    p.next()
    // allow quoted version in a single string token
    if p.cur.Kind == token.String {
        lex := p.cur.Lexeme
        if len(lex) >= 2 {
            lex = lex[1 : len(lex)-1]
        }
        // basic validation: expect leading 'v'
        if len(lex) == 0 || lex[0] != 'v' {
            p.errf("version constraint must start with 'v', got %q", lex)
        }
        p.next()
        return op + " " + lex
    }
    // Otherwise, accumulate a tolerant SemVer string out of tokens: [Ident|Number|'.'|'-']+
    var out string
    for {
        switch p.cur.Kind {
        case token.Ident, token.Number:
            out += p.cur.Lexeme
            p.next()
        case token.DotSym:
            out += "."
            p.next()
        case token.Minus:
            out += "-"
            p.next()
        default:
            if out == "" {
                p.errf("expected version after operator, got %q", p.cur.Lexeme)
            }
            return op + " " + out
        }
    }
}

