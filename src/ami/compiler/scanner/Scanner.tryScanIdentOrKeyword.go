package scanner

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// tryScanIdentOrKeyword scans identifiers or keywords from the current position.
func (s *Scanner) tryScanIdentOrKeyword(src string, n int, start int) (token.Token, bool) {
	ch, size := utf8.DecodeRuneInString(src[s.offset:])
	if !(ch == '_' || unicode.IsLetter(ch)) {
		return token.Token{}, false
	}
	s.offset += size
	for s.offset < n {
		r, sz := utf8.DecodeRuneInString(src[s.offset:])
		if r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			s.offset += sz
			continue
		}
		break
	}
	lex := src[start:s.offset]
	if k, ok := token.LookupKeyword(strings.ToLower(lex)); ok {
		return token.Token{Kind: k, Lexeme: lex, Pos: s.file.Pos(start)}, true
	}
	return token.Token{Kind: token.Ident, Lexeme: lex, Pos: s.file.Pos(start)}, true
}
