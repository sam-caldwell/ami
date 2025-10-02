package edge

import "testing"

func testLIFO_Validate_Happy(t *testing.T) {
	l := LIFO{MinCapacity: 2, MaxCapacity: 2, Backpressure: "dropNewest"}
	if err := l.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if !l.Bounded() {
		t.Fatalf("expected bounded")
	}
	if l.Delivery() != "bestEffort" {
		t.Fatalf("delivery: %s", l.Delivery())
	}
}

func testLIFO_Validate_Sad(t *testing.T) {
	// max < min
	l := LIFO{MinCapacity: 4, MaxCapacity: 2, Backpressure: "block"}
	if err := l.Validate(); err == nil {
		t.Fatalf("expected error for max<min")
	}
}
