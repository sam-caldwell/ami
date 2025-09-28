package kvstore

import (
    "testing"
    "time"
)

func TestStore_PutGet_DeleteOnRead(t *testing.T) {
    s := New()
    s.Put("k", 123, WithMaxReads(2))
    if v, ok := s.Get("k"); !ok || v.(int) != 123 { t.Fatalf("get1: %v %v", v, ok) }
    if _, ok := s.Get("k"); !ok { t.Fatalf("get2 missing") }
    if _, ok := s.Get("k"); ok { t.Fatalf("expected deletion after 2 reads") }
}

func TestStore_TTL_Expiry(t *testing.T) {
    s := New()
    s.Put("k", "v", WithTTL(10*time.Millisecond))
    if _, ok := s.Get("k"); !ok { t.Fatalf("unexpected miss before expiry") }
    time.Sleep(15 * time.Millisecond)
    if _, ok := s.Get("k"); ok { t.Fatalf("expected expired") }
    m := s.Metrics()
    if m.Expirations == 0 { t.Fatalf("expected expiration metric > 0") }
}

func TestDefault_Reset(t *testing.T) {
    ResetDefault()
    Put("a", 1)
    if _, ok := Get("a"); !ok { t.Fatalf("missing") }
    ResetDefault()
    if _, ok := Get("a"); ok { t.Fatalf("expected cleared") }
}

