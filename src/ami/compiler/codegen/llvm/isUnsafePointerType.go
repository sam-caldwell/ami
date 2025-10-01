package llvm

import "strings"

// isUnsafePointerType detects explicit raw pointer types that must be rejected for public ABI.
// It conservatively treats "ptr", leading "*" (legacy), or textual "pointer<...>" as unsafe.
func isUnsafePointerType(t string) bool {
	s := strings.TrimSpace(t)
	if s == "ptr" {
		return true
	}
	if strings.HasPrefix(s, "*") {
		return true
	}
	ls := strings.ToLower(s)
	if strings.HasPrefix(ls, "pointer<") {
		return true
	}
	return false
}
