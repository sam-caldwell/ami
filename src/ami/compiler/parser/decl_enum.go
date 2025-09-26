package parser

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// parseEnumDecl parses: enum IDENT '{' members '}'
// member: IDENT [ '=' (NUMBER | STRING) ]
func (p *Parser) parseEnumDecl() astpkg.EnumDecl {
    // consume 'enum'
    p.next()
    name := ""
    if p.cur.Kind == tok.IDENT {
        name = p.cur.Lexeme
        p.next()
    } else {
        return astpkg.EnumDecl{}
    }
    if p.cur.Kind != tok.LBRACE {
        return astpkg.EnumDecl{}
    }
    p.next() // consume '{'
    var members []astpkg.EnumMember
    for p.cur.Kind != tok.EOF {
        if p.cur.Kind == tok.RBRACE {
            p.next()
            break
        }
        // skip stray commas
        if p.cur.Kind == tok.COMMA {
            p.next()
            continue
        }
        if p.cur.Kind != tok.IDENT {
            // skip until next comma or '}'
            p.next()
            continue
        }
        memName := p.cur.Lexeme
        p.next()
        memVal := ""
        if p.cur.Kind == tok.ASSIGN {
            p.next()
            switch p.cur.Kind {
            case tok.STRING:
                // preserve quotes so semantics can distinguish string vs number
                memVal = p.cur.Lexeme
                p.next()
            case tok.NUMBER:
                memVal = p.cur.Lexeme
                p.next()
            case tok.MINUS:
                // allow negative numbers
                p.next()
                if p.cur.Kind == tok.NUMBER {
                    memVal = "-" + p.cur.Lexeme
                    p.next()
                }
            default:
                // unknown value; leave empty
            }
        }
        members = append(members, astpkg.EnumMember{Name: memName, Value: memVal})
        if p.cur.Kind == tok.COMMA {
            p.next()
        }
    }
    return astpkg.EnumDecl{Name: name, Members: members}
}

