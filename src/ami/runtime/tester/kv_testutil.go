package tester

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/runtime/kvstore"
)

// ResetDefaultKVForTest resets the global kvstore default instance.
// Intended for test isolation in packages that rely on the default store.
func ResetDefaultKVForTest(t *testing.T) {
    t.Helper()
    kvstore.ResetDefault()
}

