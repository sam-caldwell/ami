package parser

// sFileContent returns the underlying source file content for pragma scanning.
func (p *Parser) sFileContent() string {
    if p == nil || p.s == nil {
        return ""
    }
    // Provide a tiny adapter via a method on scanner.Scanner.
    return p.s.FileContent()
}

