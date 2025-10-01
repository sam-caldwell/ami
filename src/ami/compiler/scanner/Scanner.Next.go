package scanner

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// Next returns the next token by delegating to focused scanners to reduce complexity.
func (s *Scanner) Next() token.Token {
    if s == nil || s.file == nil {
        return token.Token{Kind: token.EOF, Pos: source.Position{}}
    }
    src := s.file.Content
    n := len(src)

    // 1) whitespace
    s.skipWhitespace(src, n)
    if s.offset >= n {
        return token.Token{Kind: token.EOF, Pos: s.file.Pos(s.offset)}
    }
    start := s.offset

    // 2) comments
    if tok, ok := s.tryScanComment(src, n, start); ok { return tok }
    // 3) identifiers/keywords
    if tok, ok := s.tryScanIdentOrKeyword(src, n, start); ok { return tok }
    // 4) number or duration literal
    if tok, ok := s.tryScanNumberOrDuration(src, n, start); ok { return tok }
    // 5) string literal
    if tok, ok := s.tryScanStringLiteral(src, n, start); ok { return tok }
    // 6) operators (two-char first, then one-char)
    if tok, ok := s.tryScanTwoCharOperator(src, n, start); ok { return tok }
    if tok, ok := s.tryScanOneCharOperator(src, n, start); ok { return tok }
    // 7) punctuation symbols
    if tok, ok := s.tryScanPunctuationSymbol(src, n, start); ok { return tok }
    // 8) fallback symbol
    return s.scanFallbackSymbol(src, n, start)
}
