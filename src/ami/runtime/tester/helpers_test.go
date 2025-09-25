package tester

import "testing"

func TestBuildKVInput_ComposesJSON(t *testing.T) {
    ResetDefaultKVForTest(t)
    s, err := BuildKVInput(map[string]any{"x":1}, WithKV("P","N"), KVPut("k", 42), KVGet("k"), KVEmit())
    if err != nil { t.Fatalf("compose: %v", err) }
    if s == "" || !containsAll(s, []string{"\"kv_pipeline\":\"P\"","\"kv_node\":\"N\"","\"kv_put_key\":\"k\"","\"kv_put_val\":42","\"kv_get_key\":\"k\"","\"kv_emit\":true","\"x\":1"}) {
        t.Fatalf("unexpected JSON: %s", s)
    }
}

func containsAll(s string, parts []string) bool {
    for _, p := range parts {
        if !contains(s, p) { return false }
    }
    return true
}

func contains(s, sub string) bool {
    return len(s) >= len(sub) && (func() bool { return indexOf(s, sub) >= 0 })()
}

func indexOf(s, sub string) int {
    // simple substring search
    for i := 0; i+len(sub) <= len(s); i++ {
        if s[i:i+len(sub)] == sub { return i }
    }
    return -1
}
