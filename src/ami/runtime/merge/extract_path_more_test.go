package merge

import "testing"

func Test_extractPath(t *testing.T) {
    obj := map[string]any{"a": map[string]any{"b": 3}}
    if v, ok := extractPath(obj, "a.b"); !ok || v.(int) != 3 { t.Fatalf("path result: %v ok=%v", v, ok) }
    if _, ok := extractPath(obj, "a.c"); ok { t.Fatalf("expected false for missing segment") }
    if v, ok := extractPath(obj, ""); !ok || v.(map[string]any)["a"] == nil { t.Fatalf("root path failed: %v ok=%v", v, ok) }
}

