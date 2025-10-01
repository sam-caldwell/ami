package scanner

import (
	"unicode/utf8"

	"github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// tryScanStringLiteral scans a quoted string literal, with minimal escape support.
func (s *Scanner) tryScanStringLiteral(src string, n int, start int) (token.Token, bool) {
	ch, size := utf8.DecodeRuneInString(src[s.offset:])
	if ch != '"' {
		return token.Token{}, false
	}
	s.offset += size // skip opening quote
	for s.offset < n {
		r, sz := utf8.DecodeRuneInString(src[s.offset:])
		if r == '"' {
			s.offset += sz
			return token.Token{Kind: token.String, Lexeme: src[start:s.offset], Pos: s.file.Pos(start)}, true
		}
		if r == '\\' {
			s.offset += sz
			if s.offset < n {
				_, esz := utf8.DecodeRuneInString(src[s.offset:])
				s.offset += esz
				continue
			}
			break
		}
		s.offset += sz
	}
	return token.Token{Kind: token.Unknown, Lexeme: src[start:s.offset], Pos: s.file.Pos(start)}, true
}
