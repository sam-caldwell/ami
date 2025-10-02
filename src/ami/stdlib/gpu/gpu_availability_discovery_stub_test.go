package gpu

import "testing"

func TestGPUAvailabilityAndDiscovery(t *testing.T) {
    if CudaAvailable() { t.Fatalf("CudaAvailable() stub should be false in stub build") }
    if OpenCLAvailable() { t.Fatalf("OpenCLAvailable() stub should be false in stub build") }
    if d := CudaDevices(); d != nil && len(d) != 0 { t.Fatalf("CudaDevices() expected empty slice; got %v", d) }
    if p := OpenCLPlatforms(); p != nil && len(p) != 0 { t.Fatalf("OpenCLPlatforms() expected empty slice; got %v", p) }
}

