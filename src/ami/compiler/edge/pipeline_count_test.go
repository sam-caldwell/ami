package edge

import "testing"

func TestPipeline_Count_EmptyAndNil(t *testing.T) {
	var nilP *Pipeline
	if c := nilP.Count(); c != 0 {
		t.Fatalf("nil receiver count want 0 got %d", c)
	}
	p := &Pipeline{UpstreamName: "ing"}
	if c := p.Count(); c != 0 {
		t.Fatalf("empty count want 0 got %d", c)
	}
}

func TestPipeline_Count_IncrementsAndDecrements(t *testing.T) {
	p := &Pipeline{UpstreamName: "ing", Backpressure: BackpressureBlock}
	for i := 0; i < 3; i++ {
		_ = p.Push(Event{Payload: i})
	}
	if c := p.Count(); c != 3 {
		t.Fatalf("after 3 pushes want 3 got %d", c)
	}
	_, _ = p.Pop()
	if c := p.Count(); c != 2 {
		t.Fatalf("after 1 pop want 2 got %d", c)
	}
}
