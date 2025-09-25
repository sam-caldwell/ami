package scanner

import (
    "unicode"
    "unicode/utf8"
    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

type Scanner struct {
    src    string
    off    int
    line   int
    column int
}

func New(src string) *Scanner { return &Scanner{src: src, line: 1, column: 1} }

func (s *Scanner) Next() tok.Token {
    s.skipSpace()
    if s.off >= len(s.src) {
        return tok.Token{Kind: tok.EOF, Line: s.line, Column: s.column}
    }
    r, w := utf8.DecodeRuneInString(s.src[s.off:])
    startCol := s.column
    switch {
    case unicode.IsLetter(r) || r == '_':
        start := s.off
        s.off += w; s.column += w
        for s.off < len(s.src) {
            r, w = utf8.DecodeRuneInString(s.src[s.off:])
            if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_') { break }
            s.off += w; s.column += w
        }
        return tok.Token{Kind: tok.IDENT, Lexeme: s.src[start:s.off], Line: s.line, Column: startCol}
    case unicode.IsDigit(r):
        start := s.off
        s.off += w; s.column += w
        for s.off < len(s.src) {
            r, w = utf8.DecodeRuneInString(s.src[s.off:])
            if !unicode.IsDigit(r) { break }
            s.off += w; s.column += w
        }
        return tok.Token{Kind: tok.NUMBER, Lexeme: s.src[start:s.off], Line: s.line, Column: startCol}
    case r == '"':
        start := s.off
        s.off += w; s.column += w
        for s.off < len(s.src) {
            r, w = utf8.DecodeRuneInString(s.src[s.off:])
            if r == '"' { s.off += w; s.column += w; break }
            if r == '\n' { s.line++; s.column = 1 } else { s.column += w }
            s.off += w
        }
        return tok.Token{Kind: tok.STRING, Lexeme: s.src[start:s.off], Line: s.line, Column: startCol}
    default:
        s.off += w; s.column += w
        return tok.Token{Kind: tok.ILLEGAL, Lexeme: string(r), Line: s.line, Column: startCol}
    }
}

func (s *Scanner) skipSpace() {
    for s.off < len(s.src) {
        r, w := utf8.DecodeRuneInString(s.src[s.off:])
        if r == '\n' { s.line++; s.column = 1; s.off += w; continue }
        if unicode.IsSpace(r) { s.off += w; s.column += w; continue }
        break
    }
}
