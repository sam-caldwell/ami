package sem

import "testing"

func TestFindGenericArityMismatchDeep_NestedOwned(t *testing.T) {
    exp := "slice<Owned<T>>"
    act := "slice<Owned<int,string>>"
    if m, base, w, g := findGenericArityMismatchDeep(exp, act); !m || base != "Owned" || w != 1 || g != 2 {
        t.Fatalf("expected mismatch at Owned 1 vs 2; got m=%v base=%s w=%d g=%d", m, base, w, g)
    }
}

