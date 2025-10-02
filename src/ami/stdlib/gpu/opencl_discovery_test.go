package gpu

import "testing"

// These tests focus on S-8 F-8-4 discovery helpers. In stub builds they
// should remain deterministically unavailable with empty discovery results.

func TestOpenCL_Discovery_DefaultUnavailable(t *testing.T) {
    if OpenCLAvailable() {
        t.Fatalf("OpenCLAvailable() stub should be false by default")
    }
    if ps := OpenCLPlatforms(); ps != nil && len(ps) != 0 {
        t.Fatalf("OpenCLPlatforms() expected nil or empty, got %v", ps)
    }
}

