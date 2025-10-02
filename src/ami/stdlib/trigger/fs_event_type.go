package trigger

// FsEvent represents a generic filesystem operation.
type FsEvent int

const (
    FsCreate FsEvent = iota + 1
    FsModify
    FsRemove
    FsRename
)

