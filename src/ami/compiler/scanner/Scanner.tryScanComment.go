package scanner

import (
	"unicode/utf8"

	"github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// tryScanComment scans line or block comments if present at the current offset.
func (s *Scanner) tryScanComment(src string, n int, start int) (token.Token, bool) {
	if s.offset+1 >= n || src[s.offset] != '/' {
		return token.Token{}, false
	}
	if src[s.offset+1] == '/' {
		start := s.offset
		s.offset += 2
		cstart := s.offset
		for s.offset < n {
			r, size := utf8.DecodeRuneInString(src[s.offset:])
			if r == '\n' || size == 0 {
				break
			}
			s.offset += size
		}
		text := src[cstart:s.offset]
		return token.Token{Kind: token.LineComment, Lexeme: text, Pos: s.file.Pos(start)}, true
	}
	if src[s.offset+1] == '*' {
		start := s.offset
		s.offset += 2
		cstart := s.offset
		for s.offset+1 < n && !(src[s.offset] == '*' && src[s.offset+1] == '/') {
			s.offset++
		}
		text := src[cstart:s.offset]
		if s.offset+1 < n {
			s.offset += 2
		}
		return token.Token{Kind: token.BlockComment, Lexeme: text, Pos: s.file.Pos(start)}, true
	}
	return token.Token{}, false
}
