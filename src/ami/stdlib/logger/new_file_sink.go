package logger

import "os"

func NewFileSink(path string, perm os.FileMode) *FileSink {
    if perm == 0 { perm = 0o644 }
    return &FileSink{Path: path, Perm: perm}
}

