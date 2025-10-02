package workspace

import "testing"

func TestDefaultCacheRoot(t *testing.T) {
    if p, err := DefaultCacheRoot(); err != nil || p == "" {
        t.Fatalf("unexpected result: %v %q", err, p)
    }
}

