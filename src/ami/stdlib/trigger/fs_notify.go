package trigger

import (
    stdtime "time"
    "strconv"
    amitime "github.com/sam-caldwell/ami/src/ami/stdlib/time"
    amiio "github.com/sam-caldwell/ami/src/ami/stdlib/io"
    amios "github.com/sam-caldwell/ami/src/ami/stdlib/os"
)

// FsNotify watches path for the given FsEvent and emits FileEvent occurrences.
// NOTE: Requires additional os support; currently returns ErrNotImplemented.
func FsNotify(path string, _ FsEvent) (<-chan Event[FileEvent], func(), error) {
    // Use stdlib os polling watcher; allow interval override via AMI_FS_WATCH_INTERVAL_MS
    interval := 25 * stdtime.Millisecond
    if v := amios.GetEnv("AMI_FS_WATCH_INTERVAL_MS"); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n > 0 { interval = stdtime.Duration(n) * stdtime.Millisecond }
    }
    osc, stop := amios.Watch(path, interval)
    out := make(chan Event[FileEvent], 16)
    done := make(chan struct{})
    go func(){
        defer close(out)
        for {
            select {
            case we, ok := <-osc:
                if !ok { return }
                var op FsEvent
                switch we.Op {
                case amios.OpCreate:
                    op = FsCreate
                case amios.OpModify:
                    op = FsModify
                case amios.OpRemove:
                    op = FsRemove
                case amios.OpRename:
                    op = FsRename
                default:
                    continue
                }
                var h *amiio.FHO
                if op != FsRemove {
                    if fh, err := amiio.Open(path); err == nil { h = fh }
                }
                out <- Event[FileEvent]{
                    Value: FileEvent{Handle: h, Op: op, Time: amitime.FromUnix(we.ModTime.Unix(), int64(we.ModTime.Nanosecond()))},
                    Timestamp: amitime.Now(),
                }
            case <-done:
                return
            }
        }
    }()
    stopFn := func(){ stop(); close(done) }
    return out, stopFn, nil
}

