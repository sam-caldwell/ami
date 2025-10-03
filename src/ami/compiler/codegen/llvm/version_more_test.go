package llvm

import "testing"

func TestVersion_EmptyPath_Error(t *testing.T) {
    if _, err := Version(""); err == nil {
        t.Fatalf("expected error for empty path")
    }
}

func TestVersion_BogusPath_ErrorWithToolError(t *testing.T) {
    // Attempt to execute a non-existent tool should result in ToolError
    if _, err := Version("/nonexistent/clang"); err == nil {
        t.Fatalf("expected error for bogus path")
    }
}

