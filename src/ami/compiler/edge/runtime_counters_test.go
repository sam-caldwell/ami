package edge

import "testing"

func testFIFOQueue_Counters(t *testing.T) {
	q, _ := NewFIFO(FIFO{MaxCapacity: 2, Backpressure: "dropNewest"})
	_ = q.Push(1)
	_ = q.Push(2)
	_ = q.Push(3) // dropNewest
	if push, pop, drop, full := q.Counters(); push != 3 || pop != 0 || drop != 1 || full != 0 {
		t.Fatalf("fifo counters after pushes: %d %d %d %d", push, pop, drop, full)
	}
	_, _ = q.Pop()
	_, _ = q.Pop()
	if push, pop, drop, full := q.Counters(); push != 3 || pop != 2 || drop != 1 || full != 0 {
		t.Fatalf("fifo counters after pops: %d %d %d %d", push, pop, drop, full)
	}
	// block case increments full
	qb, _ := NewFIFO(FIFO{MaxCapacity: 1, Backpressure: "block"})
	_ = qb.Push(9)
	_ = qb.Push(10)
	if _, _, _, full := qb.Counters(); full != 1 {
		t.Fatalf("fifo full count: %d", full)
	}
}

func testLIFOStack_Counters(t *testing.T) {
	s, _ := NewLIFO(LIFO{MaxCapacity: 2, Backpressure: "dropOldest"})
	_ = s.Push(1)
	_ = s.Push(2)
	_ = s.Push(3) // drop oldest
	if push, pop, drop, full := s.Counters(); push != 3 || pop != 0 || drop != 1 || full != 0 {
		t.Fatalf("lifo counters after pushes: %d %d %d %d", push, pop, drop, full)
	}
	_, _ = s.Pop()
	_, _ = s.Pop()
	if push, pop, drop, full := s.Counters(); push != 3 || pop != 2 || drop != 1 || full != 0 {
		t.Fatalf("lifo counters after pops: %d %d %d %d", push, pop, drop, full)
	}
	sb, _ := NewLIFO(LIFO{MaxCapacity: 1, Backpressure: "block"})
	_ = sb.Push(9)
	_ = sb.Push(10)
	if _, _, _, full := sb.Counters(); full != 1 {
		t.Fatalf("lifo full count: %d", full)
	}
}
