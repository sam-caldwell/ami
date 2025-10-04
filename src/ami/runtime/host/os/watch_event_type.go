package os

import "time"

// WatchEvent describes a filesystem change on a path.
type WatchEvent struct {
    Path    string
    Op      Op
    ModTime time.Time
    Size    int64
}

