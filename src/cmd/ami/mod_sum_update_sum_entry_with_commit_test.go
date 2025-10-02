package main

import "testing"

func Test_updateSumEntryWithCommit_arrayForm(t *testing.T) {
    m := map[string]any{
        "packages": []any{
            map[string]any{"name": "foo", "version": "1.0.0"},
        },
    }
    ok := updateSumEntryWithCommit(m, "foo", "1.0.0", "deadbeef", "cafebabe", "git+ssh://example")
    if !ok { t.Fatal("expected update to succeed") }
    arr := m["packages"].([]any)
    mm := arr[0].(map[string]any)
    if mm["sha256"] != "deadbeef" { t.Fatalf("sha mismatch: %v", mm["sha256"]) }
    if mm["commit"] != "cafebabe" { t.Fatalf("commit mismatch: %v", mm["commit"]) }
    if mm["source"] != "git+ssh://example" { t.Fatalf("source mismatch: %v", mm["source"]) }
}

func Test_updateSumEntryWithCommit_objectForm(t *testing.T) {
    m := map[string]any{
        "packages": map[string]any{
            "bar": map[string]any{},
        },
    }
    ok := updateSumEntryWithCommit(m, "bar", "2.3.4", "00ff", "aa11", "file+git:///abs/path")
    if !ok { t.Fatal("expected update to succeed") }
    pkgs := m["packages"].(map[string]any)
    mm := pkgs["bar"].(map[string]any)
    if mm["version"] != "2.3.4" { t.Fatalf("version: %v", mm["version"]) }
    if mm["sha256"] != "00ff" { t.Fatalf("sha: %v", mm["sha256"]) }
    if mm["commit"] != "aa11" { t.Fatalf("commit: %v", mm["commit"]) }
    if mm["source"] != "file+git:///abs/path" { t.Fatalf("source: %v", mm["source"]) }
}

