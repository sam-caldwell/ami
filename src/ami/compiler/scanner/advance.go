package scanner

// advance moves the scanner forward by w bytes and updates the column.
func (s *Scanner) advance(w int) {
    s.off += w
    s.column += w
}
