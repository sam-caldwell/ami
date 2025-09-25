package edge

import (
	"sync"
	"testing"
)

func TestLIFO_PushPop_LIFOOrder(t *testing.T) {
	st := &LIFO{Backpressure: BackpressureBlock}
	for i := 1; i <= 3; i++ {
		_ = st.Push(Event{Payload: i})
	}
	for i := 3; i >= 1; i-- {
		ev, ok := st.Pop()
		if !ok {
			t.Fatalf("pop empty")
		}
		if ev.Payload.(int) != i {
			t.Fatalf("want %d got %v", i, ev.Payload)
		}
	}
}

func TestLIFO_Backpressure_DropOldest(t *testing.T) {
	st := &LIFO{MaxCapacity: 2, Backpressure: BackpressureDrop}
	_ = st.Push(Event{Payload: 1})
	_ = st.Push(Event{Payload: 2})
	_ = st.Push(Event{Payload: 3}) // drop 1 keep [2,3]
	ev, ok := st.Pop()
	if !ok || ev.Payload.(int) != 3 {
		t.Fatalf("first pop want 3 got %v ok=%v", ev.Payload, ok)
	}
	ev, ok = st.Pop()
	if !ok || ev.Payload.(int) != 2 {
		t.Fatalf("second pop want 2 got %v ok=%v", ev.Payload, ok)
	}
}

func TestLIFO_Backpressure_BlockReturnsErr(t *testing.T) {
	st := &LIFO{MaxCapacity: 1, Backpressure: BackpressureBlock}
	_ = st.Push(Event{Payload: 1})
	if err := st.Push(Event{Payload: 2}); err == nil {
		t.Fatalf("expected ErrFull on block policy")
	}
}

func TestLIFO_ConcurrentPushes_CountMatches(t *testing.T) {
	const goroutines = 6
	const perG = 75
	st := &LIFO{Backpressure: BackpressureBlock}
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				_ = st.Push(Event{Payload: i})
			}
		}()
	}
	wg.Wait()
	if got := st.Count(); got != goroutines*perG {
		t.Fatalf("concurrent pushes: want %d got %d", goroutines*perG, got)
	}
}
