package llvm

import "strings"

// mapType returns a conservative LLVM type for a textual AMI IR type.
// Unknown and generic/container types map to opaque pointer-like handles (ptr).
func mapType(t string) string {
    tt := strings.TrimSpace(t)
    switch tt {
    case "", "void":
        return "void"
    case "bool":
        return "i1"
    case "int8":
        return "i8"
    case "int16":
        return "i16"
    case "int32":
        return "i32"
    case "int64", "int":
        return "i64"
    case "uint8":
        return "i8"
    case "uint16":
        return "i16"
    case "uint32":
        return "i32"
    case "uint64", "uint":
        return "i64"
    case "real", "float64":
        return "double"
    case "string":
        return "ptr"
    case "Owned":
        // Explicit mapping for Owned handle type
        return "ptr"
    case "Duration", "Time", "SignalType":
        return "i64"
    default:
        // Generic/container/event/error/owned types → opaque handle
        if strings.Contains(tt, "<") || strings.Contains(tt, ">") { return "ptr" }
        return "ptr"
    }
}
