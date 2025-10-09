package parser

import "github.com/sam-caldwell/ami/src/ami/compiler/token"

// captureQualifiedType extends a base type name by consuming dotted selectors
// to support cross-package qualified types like "pkg.Type". It assumes the
// parser has just advanced past the base identifier/keyword. It returns the
// updated type string and advances the token cursor over any ".ident" pairs.
func (p *Parser) captureQualifiedType(base string) string {
    // Accept zero or more sequences of "." Ident
    for p.cur.Kind == token.DotSym {
        // consume '.'
        p.next()
        if p.cur.Kind != token.Ident {
            // malformed selector; stop gracefully
            return base
        }
        base = base + "." + p.cur.Lexeme
        // consume ident
        p.next()
    }
    return base
}

