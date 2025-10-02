package buffer

import "testing"

func testFIFO_DropNewest(t *testing.T) {
	q := NewFIFO(0, 1, "dropNewest")
	if err := q.Push(1); err != nil {
		t.Fatal(err)
	}
	if err := q.Push(2); err != nil { /* ok */
	}
	v, ok := q.Pop()
	if !ok || v.(int) != 1 {
		t.Fatalf("got %v", v)
	}
}

func testLIFO_DropOldest(t *testing.T) {
	s := NewLIFO(0, 2, "dropOldest")
	_ = s.Push(1)
	_ = s.Push(2)
	_ = s.Push(3) // drop 1
	v, _ := s.Pop()
	if v.(int) != 3 {
		t.Fatalf("got %v", v)
	}
	v, _ = s.Pop()
	if v.(int) != 2 {
		t.Fatalf("got %v", v)
	}
}
