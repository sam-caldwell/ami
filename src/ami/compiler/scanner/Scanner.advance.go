package scanner

func (s *Scanner) advance(w int) {
	s.off += w
	s.column += w
}
