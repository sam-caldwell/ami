package scanner

// Comment captures a source comment with its starting position.
type Comment struct {
	Text   string
	Line   int
	Column int
	Offset int
}
