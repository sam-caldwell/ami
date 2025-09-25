package kvstore

import (
	"testing"
)

func TestStore_CapacityEvictsLRU(t *testing.T) {
	// capacity ~ size of two entries; push third to evict LRU
	s := New(Options{CapacityBytes: 20})
	defer s.Close()
	s.Put("k1", "1234567") // ~7 + key
	s.Put("k2", "7654321") // ~7 + key
	// Touch k1 to make it MRU; k2 becomes LRU
	if _, ok := s.Get("k1"); !ok {
		t.Fatalf("expected k1 present")
	}
	s.Put("k3", "abcdefg") // insert new -> should evict LRU (k2)

	if s.Has("k2") {
		t.Fatalf("expected k2 evicted")
	}
	if !s.Has("k1") || !s.Has("k3") {
		t.Fatalf("expected k1 and k3 remain")
	}
	m := s.Metrics()
	if m.Evictions == 0 {
		t.Fatalf("expected evictions > 0; got %+v", m)
	}
}
