package kvstore

import (
    "testing"
    "time"
)

func TestStore_TTL_AbsoluteExpiry(t *testing.T) {
    s := New(Options{SweepInterval: 25 * time.Millisecond})
    defer s.Close()
    s.Put("a", "alpha", TTL(40*time.Millisecond))
    if _, ok := s.Get("a"); !ok { t.Fatalf("expected a present initially") }
    // wait beyond TTL; allow at least one sweep
    time.Sleep(70 * time.Millisecond)
    if _, ok := s.Get("a"); ok { t.Fatalf("expected a expired") }
    m := s.Metrics()
    if m.Expirations == 0 { t.Fatalf("expected expirations > 0; got %+v", m) }
}

func TestStore_TTL_SlidingRefresh(t *testing.T) {
    s := New(Options{SweepInterval: 0}) // rely on access for refresh
    defer s.Close()
    s.Put("b", "bravo", SlidingTTL(50*time.Millisecond))

    deadline := time.Now().Add(160 * time.Millisecond)
    // Access repeatedly before TTL to refresh
    for time.Now().Before(deadline) {
        if _, ok := s.Get("b"); !ok { t.Fatalf("expected b present during sliding window") }
        time.Sleep(30 * time.Millisecond)
    }
    // After stopping refreshes, it should eventually expire
    time.Sleep(60 * time.Millisecond)
    if _, ok := s.Get("b"); ok { t.Fatalf("expected b expired after idle period") }
}

