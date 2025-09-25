package scanner

// ConsumeComments returns and clears any comments encountered immediately
// before the next non-space token.
func (s *Scanner) ConsumeComments() []Comment {
	if len(s.pending) == 0 {
		return nil
	}
	out := make([]Comment, len(s.pending))
	copy(out, s.pending)
	s.pending = nil
	return out
}
