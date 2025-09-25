package kvstore

import "testing"

func TestStore_Metrics_Counters(t *testing.T) {
	s := New(Options{})
	defer s.Close()
	s.Put("a", 1)
	s.Put("b", 2)
	s.Get("a")
	s.Get("a")
	s.Get("c") // miss
	s.Del("b")
	m := s.Metrics()
	if m.Hits != 2 || m.Misses != 1 {
		t.Fatalf("unexpected metrics: %+v", m)
	}
	if m.Entries != 1 {
		t.Fatalf("expected 1 entry; got %+v", m)
	}
}
