package parser

import "github.com/sam-caldwell/ami/src/ami/compiler/token"

// syncUntil advances until one of the specified kinds is found.
func (p *Parser) syncUntil(kinds ...token.Kind) {
    set := make(map[token.Kind]struct{}, len(kinds))
    for _, k := range kinds {
        set[k] = struct{}{}
    }
    for {
        if _, ok := set[p.cur.Kind]; ok {
            return
        }
        if p.cur.Kind == token.EOF {
            return
        }
        p.next()
    }
}

