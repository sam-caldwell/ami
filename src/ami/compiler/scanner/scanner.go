package scanner

import (
    "unicode"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// Scanner performs a minimal lexical scan over a source.File.
type Scanner struct {
    file   *source.File
    offset int
}

// New creates a new Scanner for the provided file.
func New(f *source.File) *Scanner { return &Scanner{file: f, offset: 0} }

// Next returns the next token. This is a minimal implementation sufficient
// for basic tests and scaffolding.
func (s *Scanner) Next() token.Token {
    if s == nil || s.file == nil {
        return token.Token{Kind: token.EOF, Pos: source.Position{}}
    }
    src := s.file.Content
    n := len(src)
    // skip whitespace
    for s.offset < n && unicode.IsSpace(rune(src[s.offset])) { s.offset++ }
    if s.offset >= n {
        return token.Token{Kind: token.EOF, Pos: s.file.Pos(s.offset)}
    }
    start := s.offset
    ch := rune(src[s.offset])
    // identifiers ([A-Za-z_][A-Za-z0-9_]*) simplified: letters only for now
    if unicode.IsLetter(ch) {
        s.offset++
        for s.offset < n {
            r := rune(src[s.offset])
            if unicode.IsLetter(r) || unicode.IsDigit(r) {
                s.offset++
                continue
            }
            break
        }
        return token.Token{Kind: token.Ident, Lexeme: src[start:s.offset], Pos: s.file.Pos(start)}
    }
    // single-character symbol
    s.offset++
    return token.Token{Kind: token.Symbol, Lexeme: src[start:s.offset], Pos: s.file.Pos(start)}
}

