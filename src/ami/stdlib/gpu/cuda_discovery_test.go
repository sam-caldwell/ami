package gpu

import "testing"

// These tests focus on S-8 F-8-2 discovery helpers. In stub builds they
// should remain deterministically unavailable with empty discovery results.

func TestCuda_Discovery_DefaultUnavailable(t *testing.T) {
    if CudaAvailable() {
        t.Fatalf("CudaAvailable() stub should be false by default")
    }
    if ds := CudaDevices(); ds != nil && len(ds) != 0 {
        t.Fatalf("CudaDevices() expected nil or empty, got %v", ds)
    }
}

