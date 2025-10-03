package kvstore

import (
    "testing"
)

func TestResetDefault_ResetsStore(t *testing.T) {
    ResetDefault()
    s := Default()
    s.Put("k", 1)
    if _, ok := s.Get("k"); !ok { t.Fatalf("expected present") }
    ResetDefault()
    if _, ok := Default().Get("k"); ok { t.Fatalf("expected reset to clear store") }
}

func TestNamespace_CreatesAndReusesStores(t *testing.T) {
    ResetRegistry()
    a := Namespace("a")
    b := Namespace("b")
    if a == nil || b == nil || a == b { t.Fatalf("namespace stores invalid: %v %v", a, b) }
    a.Put("k1", 1)
    if v, ok := Namespace("a").Get("k1"); !ok || v.(int) != 1 { t.Fatalf("ns a missing value: %v ok=%v", v, ok) }
}

func TestWithMaxReads_Basic(t *testing.T) {
    s := New()
    s.Put("k", 1, WithMaxReads(2))
    if v, ok := s.Get("k"); !ok || v.(int) != 1 { t.Fatalf("read1: %v ok=%v", v, ok) }
    if v, ok := s.Get("k"); !ok || v.(int) != 1 { t.Fatalf("read2: %v ok=%v", v, ok) }
    // third read should miss
    if _, ok := s.Get("k"); ok { t.Fatalf("expected delete-on-read after 2") }
    if s.Del("k") { t.Fatalf("already deleted") }
}

func TestWithMaxReads_IgnoresNonPositive(t *testing.T) {
    s := New()
    s.Put("k", 1, WithMaxReads(0))
    if _, ok := s.Get("k"); !ok { t.Fatalf("expected present after read when maxReads=0") }
    if !s.Has("k") { t.Fatalf("expected key to remain when maxReads=0") }
}

