package source

// File models a single source file and its line offsets.
type File struct {
    Name        string
    Size        int
    Base        int
    lineOffsets []int // 0-based byte offsets where each line starts
}

// PositionFor converts an absolute offset within the file to Position.
// Offsets outside [0, Size] are clamped.
func (f *File) PositionFor(offset int) Position {
    if offset < 0 {
        offset = 0
    }
    if offset > f.Size {
        offset = f.Size
    }
    // binary search for the line start index whose start <= offset
    lo, hi := 0, len(f.lineOffsets)
    for lo < hi {
        mid := (lo + hi) / 2
        if f.lineOffsets[mid] <= offset {
            lo = mid + 1
        } else {
            hi = mid
        }
    }
    lineIdx := lo - 1
    if lineIdx < 0 {
        lineIdx = 0
    }
    lineStart := f.lineOffsets[lineIdx]
    col := offset - lineStart + 1
    return Position{Offset: offset, Line: lineIdx + 1, Column: col}
}

// LineStartOffset returns the byte offset of the given 1-based line.
func (f *File) LineStartOffset(line int) int {
    if line <= 0 || line > len(f.lineOffsets) {
        return -1
    }
    return f.lineOffsets[line-1]
}

