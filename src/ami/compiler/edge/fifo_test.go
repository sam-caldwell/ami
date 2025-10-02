package edge

import "testing"

func testFIFO_Validate_Happy(t *testing.T) {
	f := FIFO{MinCapacity: 0, MaxCapacity: 128, Backpressure: "block"}
	if err := f.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if !f.Bounded() {
		t.Fatalf("expected bounded")
	}
	if f.Delivery() != "atLeastOnce" {
		t.Fatalf("delivery: %s", f.Delivery())
	}
}

func testFIFO_Validate_Sad(t *testing.T) {
	// invalid backpressure
	f := FIFO{MinCapacity: 0, MaxCapacity: 1, Backpressure: "invalid"}
	if err := f.Validate(); err == nil {
		t.Fatalf("expected error for invalid backpressure")
	}
}
