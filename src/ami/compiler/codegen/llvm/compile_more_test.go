package llvm

import "testing"

func TestCompileLLToObject_ErrorsOnMissingClang(t *testing.T) {
    if err := CompileLLToObject("", "in.ll", "out.o", ""); err == nil { t.Fatalf("expected error for empty clang") }
}

