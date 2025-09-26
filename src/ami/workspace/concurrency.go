package workspace

import (
    "runtime"
    "strings"
)

// ResolveConcurrency returns effective concurrency as an integer.
func (w *Workspace) ResolveConcurrency() int {
    switch v := w.Toolchain.Compiler.Concurrency.(type) {
    case int:
        if v >= 1 {
            return v
        }
    case int64:
        if v >= 1 {
            return int(v)
        }
    case int32:
        if v >= 1 {
            return int(v)
        }
    case string:
        if strings.ToUpper(v) == "NUM_CPU" {
            return runtime.NumCPU()
        }
    }
    return 1
}

