package tester

import (
    kv "github.com/sam-caldwell/ami/src/ami/runtime/kvstore"
    "testing"
)

// ResetDefaultKVForTest registers a cleanup that resets the default kvstore
// registry after the test completes, improving isolation across tests that
// touch the global kv registry via tester hooks.
func ResetDefaultKVForTest(t *testing.T) {
    t.Helper()
    t.Cleanup(func(){ kv.ResetDefault() })
}

