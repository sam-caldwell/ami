package edge

import "testing"

func TestFIFO_Validate_Happy(t *testing.T) {
	f := &FIFO{MinCapacity: 1, MaxCapacity: 8, Backpressure: BackpressureBlock, TypeName: "[]byte"}
	if err := f.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFIFO_Validate_Sad(t *testing.T) {
	cases := []*FIFO{
		{MinCapacity: -1, MaxCapacity: 1},
		{MinCapacity: 2, MaxCapacity: 1},
		{MinCapacity: 0, MaxCapacity: 1, Backpressure: "weird"},
	}
	for i, c := range cases {
		if err := c.Validate(); err == nil {
			t.Fatalf("case %d expected error", i)
		}
	}
}

func TestLIFO_Validate_Happy(t *testing.T) {
	l := &LIFO{MinCapacity: 0, MaxCapacity: 4, Backpressure: BackpressureDrop, TypeName: "some.T"}
	if err := l.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLIFO_Validate_Sad(t *testing.T) {
	cases := []*LIFO{
		{MinCapacity: -2, MaxCapacity: 0},
		{MinCapacity: 5, MaxCapacity: 4},
		{MinCapacity: 0, MaxCapacity: 1, Backpressure: "unknown"},
	}
	for i, c := range cases {
		if err := c.Validate(); err == nil {
			t.Fatalf("case %d expected error", i)
		}
	}
}

func TestPipeline_Validate_Happy(t *testing.T) {
	p := &Pipeline{UpstreamName: "csvReaderPipeline", MinCapacity: 10, MaxCapacity: 20, Backpressure: BackpressureBlock, TypeName: "someProject.CsvRecord"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPipeline_Validate_Sad(t *testing.T) {
	cases := []*Pipeline{
		{UpstreamName: "", MinCapacity: 0, MaxCapacity: 1},
		{UpstreamName: "X", MinCapacity: -1, MaxCapacity: 0},
		{UpstreamName: "X", MinCapacity: 2, MaxCapacity: 1},
		{UpstreamName: "X", MinCapacity: 0, MaxCapacity: 1, Backpressure: "reject"},
	}
	for i, c := range cases {
		if err := c.Validate(); err == nil {
			t.Fatalf("case %d expected error", i)
		}
	}
}
