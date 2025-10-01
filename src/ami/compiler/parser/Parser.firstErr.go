package parser

func (p *Parser) firstErr() error {
    if len(p.errors) == 0 {
        return nil
    }
    return p.errors[0]
}

