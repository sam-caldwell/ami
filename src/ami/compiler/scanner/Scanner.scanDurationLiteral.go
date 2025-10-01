package scanner

import (
	"unicode"
	"unicode/utf8"
)

// scanDurationLiteral tries to scan a Go-style duration literal at the current offset.
// Supported units: ns, us, ms, s, m, h. Allows a decimal in the first segment only.
// Examples: 300ms, 5s, 2h45m, 1.5h
func (s *Scanner) scanDurationLiteral() (string, bool) {
	if s == nil || s.file == nil {
		return "", false
	}
	src := s.file.Content
	n := len(src)
	i := s.offset
	start := i
	// first number (digits, optional decimal)
	digits := 0
	for i < n {
		r, sz := utf8.DecodeRuneInString(src[i:])
		if unicode.IsDigit(r) {
			i += sz
			digits++
			continue
		}
		break
	}
	if digits == 0 {
		return "", false
	}
	// optional fractional part for the first segment
	if i < n && src[i] == '.' {
		i++
		frac := 0
		for i < n {
			r, sz := utf8.DecodeRuneInString(src[i:])
			if unicode.IsDigit(r) {
				i += sz
				frac++
				continue
			}
			break
		}
		// invalid if no fractional digits present
		if frac == 0 {
			return "", false
		}
	}
	// require a unit after the first numeric part
	unitLen := matchDurationUnit(src, i)
	if unitLen == 0 {
		return "", false
	}
	i += unitLen
	// zero or more additional integer+unit segments (no decimal allowed)
	for i < n {
		// require at least one digit to proceed; otherwise break
		j := i
		d := 0
		for j < n {
			r, sz := utf8.DecodeRuneInString(src[j:])
			if unicode.IsDigit(r) {
				j += sz
				d++
				continue
			}
			break
		}
		if d == 0 {
			break
		}
		u := matchDurationUnit(src, j)
		if u == 0 {
			return "", false
		}
		i = j + u
	}
	lex := src[start:i]
	s.offset = i
	return lex, true
}
