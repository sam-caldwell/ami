package llvm

import "testing"

func TestCompileLLToObject_EmptyClangPath(t *testing.T) {
    if err := CompileLLToObject("", "in.ll", "out.o", DefaultTriple); err == nil {
        t.Fatalf("expected error for empty clang path")
    }
}

