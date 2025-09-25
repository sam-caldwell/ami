package kvstore

import "testing"

func TestRegistry_IsolatesPerPipelineAndNode(t *testing.T) {
    r := NewRegistry(Options{})
    defer r.Close()
    s1 := r.Get("P1", "N1")
    s2 := r.Get("P1", "N2")
    s3 := r.Get("P2", "N1")

    if s1 == s2 || s1 == s3 || s2 == s3 { t.Fatalf("stores must be distinct per (pipeline,node)") }

    s1.Put("k", "v1")
    s2.Put("k", "v2")
    s3.Put("k", "v3")
    if v, _ := s1.Get("k"); v.(string) != "v1" { t.Fatalf("s1 wrong val") }
    if v, _ := s2.Get("k"); v.(string) != "v2" { t.Fatalf("s2 wrong val") }
    if v, _ := s3.Get("k"); v.(string) != "v3" { t.Fatalf("s3 wrong val") }
}

func TestRegistry_Reset_ClearsStores(t *testing.T) {
    r := NewRegistry(Options{})
    s := r.Get("P","N")
    s.Put("k","v")
    if len(r.Snapshot()) == 0 { t.Fatalf("expected snapshot non-empty before reset") }
    r.Reset()
    if len(r.Snapshot()) != 0 { t.Fatalf("expected snapshot empty after reset") }
}

func TestResetDefault_ReplacesGlobalRegistry(t *testing.T) {
    // ensure default has at least one store
    d := Default()
    s := d.Get("PX","NY")
    s.Put("x",1)
    if len(d.Snapshot()) == 0 { t.Fatalf("expected default snapshot non-empty before reset") }
    // reset default and verify empty snapshot
    ResetDefault()
    d2 := Default()
    if len(d2.Snapshot()) != 0 { t.Fatalf("expected default snapshot empty after reset") }
}
