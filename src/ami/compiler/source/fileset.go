package source

import "github.com/sam-caldwell/ami/src/ami/compiler/token"

// Position represents a 1-based line/column with a 0-based byte offset.
type Position struct {
	Offset int
	Line   int
	Column int
}

// File models a single source file and its line offsets.
type File struct {
	Name        string
	Size        int
	Base        int
	lineOffsets []int // 0-based byte offsets where each line starts
}

// FileSet tracks a collection of files and provides position mapping.
type FileSet struct {
	files []*File
	base  int
}

// NewFileSet creates an empty fileset.
func NewFileSet() *FileSet { return &FileSet{} }

// AddFileFromSource registers a file in the set and computes line offsets.
func (fs *FileSet) AddFileFromSource(name, src string) *File {
	f := &File{Name: name, Size: len(src), Base: fs.base}
	f.lineOffsets = make([]int, 0, 16)
	// first line always starts at 0
	f.lineOffsets = append(f.lineOffsets, 0)
	for i := 0; i < len(src); i++ {
		if src[i] == token.LexLf && i+1 < len(src) {
			f.lineOffsets = append(f.lineOffsets, i+1)
		}
	}
	fs.files = append(fs.files, f)
	// advance base; +1 to avoid overlaps similar to go/token behavior
	fs.base += f.Size + 1
	return f
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
