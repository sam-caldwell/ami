package edge

import "testing"

func TestLIFO_Count_EmptyAndNil(t *testing.T) {
	var nilL *LIFO
	if c := nilL.Count(); c != 0 {
		t.Fatalf("nil receiver count want 0 got %d", c)
	}
	l := &LIFO{}
	if c := l.Count(); c != 0 {
		t.Fatalf("empty count want 0 got %d", c)
	}
}

func TestLIFO_Count_IncrementsAndDecrements(t *testing.T) {
	l := &LIFO{Backpressure: BackpressureBlock}
	for i := 0; i < 3; i++ {
		_ = l.Push(Event{Payload: i})
	}
	if c := l.Count(); c != 3 {
		t.Fatalf("after 3 pushes want 3 got %d", c)
	}
	_, _ = l.Pop()
	if c := l.Count(); c != 2 {
		t.Fatalf("after 1 pop want 2 got %d", c)
	}
}
