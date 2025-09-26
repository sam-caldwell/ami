package source

// Position represents a 1-based line/column with a 0-based byte offset.
type Position struct {
    Offset int
    Line   int
    Column int
}

