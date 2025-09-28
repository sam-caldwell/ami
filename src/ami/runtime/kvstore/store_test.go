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

func TestStore_SlidingTTL_ExtendsLifetime(t *testing.T) {
    s := New()
    s.Put("k", 1, WithTTL(15*time.Millisecond), WithSlidingTTL())
    for i := 0; i < 3; i++ {
        time.Sleep(8 * time.Millisecond)
        if _, ok := s.Get("k"); !ok { t.Fatalf("key expired early at i=%d", i) }
    }
    time.Sleep(20 * time.Millisecond)
    if _, ok := s.Get("k"); ok { t.Fatalf("expected expired after idle period") }
}

func TestStore_CapacityLRU_EvictsOldest(t *testing.T) {
    s := New()
    s.SetCapacity(2)
    s.Put("a", 1)
    s.Put("b", 2)
    s.Put("c", 3) // evict a
    if _, ok := s.Get("a"); ok { t.Fatalf("expected 'a' evicted") }
    // LRU touch: get b then insert d; evict c
    if _, ok := s.Get("b"); !ok { t.Fatalf("expected 'b' present") }
    s.Put("d", 4)
    if _, ok := s.Get("c"); ok { t.Fatalf("expected 'c' evicted after touching 'b'") }
    m := s.Metrics()
    if m.Evictions < 2 { t.Fatalf("expected evictions>=2, got %d", m.Evictions) }
}
