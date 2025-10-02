package gpu

import "testing"

func TestCuda_EnvGatedAvailability(t *testing.T) {
    // Default: should be false
    if CudaAvailable() {
        t.Fatalf("CudaAvailable default should be false")
    }
    t.Setenv("AMI_GPU_FORCE_CUDA", "1")
    if !CudaAvailable() {
        t.Fatalf("CudaAvailable should be true when AMI_GPU_FORCE_CUDA=1")
    }
}

