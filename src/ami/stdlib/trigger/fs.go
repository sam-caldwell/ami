package trigger

import (
    stdtime "time"
    amitime "github.com/sam-caldwell/ami/src/ami/stdlib/time"
    amiio "github.com/sam-caldwell/ami/src/ami/stdlib/io"
    amios "github.com/sam-caldwell/ami/src/ami/stdlib/os"
)

// FsEvent represents a generic filesystem operation.
type FsEvent int

const (
    FsCreate FsEvent = iota + 1
    FsModify
    FsRemove
    FsRename
)

// FileEvent is the payload for filesystem notifications.
// Handle may be nil for events where a handle is unavailable (e.g., removed).
type FileEvent struct {
    Handle *amiio.FHO
    Op     FsEvent
    Time   amitime.Time
}

// FsNotify watches path for the given FsEvent and emits FileEvent occurrences.
// NOTE: Requires additional os support; currently returns ErrNotImplemented.
func FsNotify(path string, _ FsEvent) (<-chan Event[FileEvent], func(), error) {
    // Use stdlib os polling watcher; small interval for responsiveness.
    osc, stop := amios.Watch(path, 25*stdtime.Millisecond)
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
