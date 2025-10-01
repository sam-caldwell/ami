package trigger

import (
    amitime "github.com/sam-caldwell/ami/src/ami/stdlib/time"
    amiio "github.com/sam-caldwell/ami/src/ami/stdlib/io"
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
    return nil, nil, ErrNotImplemented
}

