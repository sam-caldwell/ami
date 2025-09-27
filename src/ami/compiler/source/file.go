package source

// File holds a source file's name and content and can translate
// byte offsets to human-readable positions.
type File struct {
    Name    string
    Content string
}

// Pos converts a byte offset into a Position with 1-based line/column.
// If offset is out of range, returns a zero Position.
func (f *File) Pos(offset int) Position {
    if f == nil || offset < 0 || offset > len(f.Content) {
        return Position{}
    }
    line := 1
    col := 1
    for i := 0; i < offset; i++ {
        if f.Content[i] == '\n' {
            line++
            col = 1
        } else {
            col++
        }
    }
    if offset > 0 { col++ }
    return Position{Line: line, Column: col, Offset: offset}
}
