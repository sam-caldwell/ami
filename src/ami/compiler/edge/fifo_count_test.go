package edge

import "testing"

func TestFIFO_Count_EmptyAndNil(t *testing.T) {
    var nilF *FIFO
    if c := nilF.Count(); c != 0 { t.Fatalf("nil receiver count want 0 got %d", c) }
    f := &FIFO{}
    if c := f.Count(); c != 0 { t.Fatalf("empty count want 0 got %d", c) }
}

func TestFIFO_Count_IncrementsAndDecrements(t *testing.T) {
    f := &FIFO{Backpressure: BackpressureBlock}
    for i := 0; i < 3; i++ { _ = f.Push(Event{Payload: i}) }
    if c := f.Count(); c != 3 { t.Fatalf("after 3 pushes want 3 got %d", c) }
    _, _ = f.Pop()
    if c := f.Count(); c != 2 { t.Fatalf("after 1 pop want 2 got %d", c) }
}

