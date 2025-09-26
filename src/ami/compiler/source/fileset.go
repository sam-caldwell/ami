package source

import "github.com/sam-caldwell/ami/src/ami/compiler/token"

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
