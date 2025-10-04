package os

// Op represents a filesystem operation kind.
type Op int

const (
    OpCreate Op = iota + 1
    OpModify
    OpRemove
    OpRename
)

