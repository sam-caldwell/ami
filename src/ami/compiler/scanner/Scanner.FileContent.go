package scanner

// FileContent exposes the underlying file content for tools in the same module
// that need whole-line scanning (e.g., pragma collection). This method has no
// side effects and does not advance the scanner.
func (s *Scanner) FileContent() string {
	if s == nil || s.file == nil {
		return ""
	}
	return s.file.Content
}
