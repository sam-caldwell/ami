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
    devs := CudaDevices()
    if len(devs) != 1 || devs[0].Backend != "cuda" || devs[0].ID != 0 || devs[0].Name == "" {
        t.Fatalf("CudaDevices env dummy mismatch: %+v", devs)
    }
}
