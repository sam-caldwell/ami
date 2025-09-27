package logger

import (
    "os"
    "path/filepath"
)

// FileSink appends lines to a file path, creating parent directories as needed.
type FileSink struct {
    Path string
    Perm os.FileMode
    f    *os.File
}

func NewFileSink(path string, perm os.FileMode) *FileSink {
    if perm == 0 { perm = 0o644 }
    return &FileSink{Path: path, Perm: perm}
}

func (s *FileSink) Start() error {
    if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil { return err }
    f, err := os.OpenFile(s.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, s.Perm)
    if err != nil { return err }
    s.f = f
    return nil
}

func (s *FileSink) Write(line []byte) error {
    if s.f == nil { if err := s.Start(); err != nil { return err } }
    _, err := s.f.Write(line)
    return err
}

func (s *FileSink) Close() error {
    if s.f != nil { return s.f.Close() }
    return nil
}

