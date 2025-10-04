package io

import (
    "os"
    "time"
)

// FileInfo is a simplified stat structure.
type FileInfo struct {
    Name    string
    Size    int64
    Mode    os.FileMode
    ModTime time.Time
    IsDir   bool
}

