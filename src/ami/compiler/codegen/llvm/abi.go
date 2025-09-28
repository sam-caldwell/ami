package llvm

import "strings"

// abiType maps AMI surface types to LLVM ABI-safe types for public function signatures.
// It removes raw pointer exposure by mapping any pointer-like/classified type to an i64 handle.
func abiType(t string) string {
    mt := mapType(t)
    if mt == "ptr" { return "i64" }
    return mt
}

// isUnsafePointerType detects explicit raw pointer types that must be rejected for public ABI.
// It conservatively treats "ptr", leading "*" (legacy), or textual "pointer<...>" as unsafe.
func isUnsafePointerType(t string) bool {
    s := strings.TrimSpace(t)
    if s == "ptr" { return true }
    if strings.HasPrefix(s, "*") { return true }
    ls := strings.ToLower(s)
    if strings.HasPrefix(ls, "pointer<") { return true }
    return false
}

