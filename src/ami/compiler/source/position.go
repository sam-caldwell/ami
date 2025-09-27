package source

// Position represents a position in a source file.
type Position struct {
    Offset int // byte offset from start of file
    Line   int // 1-based line number
    Column int // 1-based column number (in bytes)
}

