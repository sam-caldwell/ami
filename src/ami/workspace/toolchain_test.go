package workspace

import "testing"

func TestToolchain_BasenamePair(t *testing.T) {
    tc := Toolchain{}
    if tc.Compiler.Options != nil || tc.Linker.Options != nil || tc.Linter.Options != nil {
        t.Fatalf("unexpected non-zero: %+v", tc)
    }
}

