package llvm

import "testing"

func TestLinkObjects_ErrorsOnMissingArgs(t *testing.T) {
    if err := LinkObjects("", []string{"a.o"}, "a.out", ""); err == nil { t.Fatalf("expected error for empty clang") }
    if err := LinkObjects("clang", nil, "a.out", ""); err == nil { t.Fatalf("expected error for no objects") }
}

