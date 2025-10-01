package os

import (
    goos "os"
    "time"
)

// Op represents a filesystem operation kind.
type Op int

const (
    OpCreate Op = iota + 1
    OpModify
    OpRemove
    OpRename
)

// WatchEvent describes a filesystem change on a path.
type WatchEvent struct {
    Path    string
    Op      Op
    ModTime time.Time
    Size    int64
}

// Watch starts a simple polling watcher for a single file path.
// It emits an event when the file is created, modified (mtime or size changes),
// or removed. Rename is reported as Remove followed by Create if the path is
// recreated.
func Watch(path string, interval time.Duration) (<-chan WatchEvent, func()) {
    out := make(chan WatchEvent, 16)
    stop := make(chan struct{})
    // baseline stat (existence, mtime, size)
    var lastMT time.Time
    var lastSz int64
    existed := false
    if fi, err := goos.Stat(path); err == nil {
        existed = true
        lastMT = fi.ModTime()
        lastSz = fi.Size()
    }
    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()
        defer close(out)
        for {
            select {
            case <-ticker.C:
                fi, err := goos.Stat(path)
                if err != nil {
                    // treat not-exist as remove when it previously existed
                    if existed {
                        existed = false
                        out <- WatchEvent{Path: path, Op: OpRemove, ModTime: time.Now()}
                    }
                    continue
                }
                mt := fi.ModTime(); sz := fi.Size()
                if !existed {
                    existed = true; lastMT = mt; lastSz = sz
                    out <- WatchEvent{Path: path, Op: OpCreate, ModTime: mt, Size: sz}
                    continue
                }
                if mt != lastMT || sz != lastSz {
                    lastMT = mt; lastSz = sz
                    out <- WatchEvent{Path: path, Op: OpModify, ModTime: mt, Size: sz}
                }
            case <-stop:
                return
            }
        }
    }()
    return out, func(){ close(stop) }
}

