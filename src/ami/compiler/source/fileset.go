package source

// FileSet is a minimal container for a set of files.
type FileSet struct {
    Files []*File
}

// AddFile appends a new file and returns it.
func (s *FileSet) AddFile(name, content string) *File {
    f := &File{Name: name, Content: content}
    s.Files = append(s.Files, f)
    return f
}

// FileByName returns the file with the given name or nil when not found.
func (s *FileSet) FileByName(name string) *File {
    for _, f := range s.Files {
        if f != nil && f.Name == name { return f }
    }
    return nil
}

