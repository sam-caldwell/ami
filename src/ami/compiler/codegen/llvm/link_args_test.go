package llvm

import "testing"

func TestLinkObjects_EarlyErrors(t *testing.T) {
    // empty clang path
    if err := LinkObjects("", []string{"a.o"}, "bin", DefaultTriple); err == nil {
        t.Fatalf("expected error for empty clang path")
    }
    // no objects to link — should not attempt to invoke tool
    if err := LinkObjects("clang", nil, "bin", DefaultTriple); err == nil || err.Error() == "clang failed" {
        t.Fatalf("expected no objects error, got: %v", err)
    }
}

