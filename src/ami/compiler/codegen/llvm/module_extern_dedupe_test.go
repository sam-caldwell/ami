package llvm

import "testing"

// TestRequireExtern_Deduplicates verifies that duplicate extern declarations are not appended.
func TestRequireExtern_Deduplicates(t *testing.T) {
    e := NewModuleEmitter("p", "u")
    e.RequireExtern("declare void @foo()")
    e.RequireExtern("declare void @foo()")
    s := e.Build()
    // Expect only one occurrence
    if first := len([]byte(s)); first == 0 { t.Fatalf("empty module") }
    // A crude check: second append would duplicate; ensure count of substring is 1
    count := 0
    for i := 0; i < len(s); i++ {
        if i+len("declare void @foo()") <= len(s) && s[i:i+len("declare void @foo()") ] == "declare void @foo()" { count++ }
    }
    if count != 1 { t.Fatalf("expected 1 extern decl, got %d\n%s", count, s) }
}

