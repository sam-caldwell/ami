package scanner

import (
	"unicode"
	"unicode/utf8"

	"github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// tryScanNumberOrDuration scans a duration literal or numeric literal.
func (s *Scanner) tryScanNumberOrDuration(src string, n int, start int) (token.Token, bool) {
	ch, size := utf8.DecodeRuneInString(src[s.offset:])
	if !unicode.IsDigit(ch) {
		return token.Token{}, false
	}
	// Attempt duration literal
	if lex, ok := s.scanDurationLiteral(); ok {
		return token.Token{Kind: token.DurationLit, Lexeme: lex, Pos: s.file.Pos(start)}, true
	}
	// radix prefixes: 0x, 0b, 0o
	if src[s.offset] == '0' && s.offset+1 < n {
		switch src[s.offset+1] {
		case 'x', 'X':
			s.offset += 2
			for s.offset < n {
				r, sz := utf8.DecodeRuneInString(src[s.offset:])
				if ('0' <= r && r <= '9') || ('a' <= r && r <= 'f') || ('A' <= r && r <= 'F') {
					s.offset += sz
					continue
				}
				break
			}
			return token.Token{Kind: token.Number, Lexeme: src[start:s.offset], Pos: s.file.Pos(start)}, true
		case 'b', 'B':
			s.offset += 2
			for s.offset < n {
				r, sz := utf8.DecodeRuneInString(src[s.offset:])
				if r == '0' || r == '1' {
					s.offset += sz
					continue
				}
				break
			}
			return token.Token{Kind: token.Number, Lexeme: src[start:s.offset], Pos: s.file.Pos(start)}, true
		case 'o', 'O':
			s.offset += 2
			for s.offset < n {
				r, sz := utf8.DecodeRuneInString(src[s.offset:])
				if '0' <= r && r <= '7' {
					s.offset += sz
					continue
				}
				break
			}
			return token.Token{Kind: token.Number, Lexeme: src[start:s.offset], Pos: s.file.Pos(start)}, true
		}
	}
	// decimal int/float
	s.offset += size
	for s.offset < n {
		r, sz := utf8.DecodeRuneInString(src[s.offset:])
		if unicode.IsDigit(r) {
			s.offset += sz
			continue
		}
		break
	}
	// fractional
	if s.offset < n && src[s.offset] == '.' {
		s.offset++
		for s.offset < n {
			r, sz := utf8.DecodeRuneInString(src[s.offset:])
			if unicode.IsDigit(r) {
				s.offset += sz
				continue
			}
			break
		}
	}
	// exponent
	if s.offset < n && (src[s.offset] == 'e' || src[s.offset] == 'E') {
		s.offset++
		if s.offset < n && (src[s.offset] == '+' || src[s.offset] == '-') {
			s.offset++
		}
		for s.offset < n {
			r, sz := utf8.DecodeRuneInString(src[s.offset:])
			if unicode.IsDigit(r) {
				s.offset += sz
				continue
			}
			break
		}
	}
	return token.Token{Kind: token.Number, Lexeme: src[start:s.offset], Pos: s.file.Pos(start)}, true
}
