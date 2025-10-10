package driver

import "github.com/sam-caldwell/ami/src/ami/compiler/ir"

// Backward-compatible helper for existing time.* method lowering that expects
// an int64 receiver handle.
func synthesizeMethodRecvArg(st *lowerState, fullName string) (ir.Value, bool) {
    return synthesizeMethodRecvArgWithFallback(st, fullName, "int64")
}

