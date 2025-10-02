package edge

import "testing"

func testPipeline_Validate_Happy(t *testing.T) {
	p := Pipeline{Name: "OtherPipe", Type: "Event<T>"}
	if err := p.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func testPipeline_Validate_Sad(t *testing.T) {
	p := Pipeline{}
	if err := p.Validate(); err == nil {
		t.Fatalf("expected error for empty name")
	}
}
