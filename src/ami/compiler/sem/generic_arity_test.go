package sem

import (
    "testing"
)

func TestGenericBaseAndArity(t *testing.T) {
    b, n, ok := genericBaseAndArity("Owned<int>")
    if !ok || b != "Owned" || n != 1 { t.Fatalf("unexpected: %v %d %v", b, n, ok) }
    b, n, ok = genericBaseAndArity("map<string, slice<int>>")
    if !ok || b != "map" || n != 2 { t.Fatalf("unexpected: %v %d %v", b, n, ok) }
    if m, base, w, g := isGenericArityMismatch("Owned<T>", "Owned<int,string>"); !m || base != "Owned" || w != 1 || g != 2 {
        t.Fatalf("expected mismatch for Owned arity; got m=%v base=%s w=%d g=%d", m, base, w, g)
    }
}

