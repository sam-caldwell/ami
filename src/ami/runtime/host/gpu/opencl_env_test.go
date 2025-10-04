package gpu

import "testing"

func TestOpenCL_EnvGatedAvailability(t *testing.T) {
    // Default: should be false
    if OpenCLAvailable() {
        t.Fatalf("OpenCLAvailable default should be false")
    }
    t.Setenv("AMI_GPU_FORCE_OPENCL", "true")
    if !OpenCLAvailable() {
        t.Fatalf("OpenCLAvailable should be true when AMI_GPU_FORCE_OPENCL=true")
    }
    plats := OpenCLPlatforms()
    if len(plats) != 1 || plats[0].Vendor == "" || plats[0].Name == "" || plats[0].Version == "" {
        t.Fatalf("OpenCLPlatforms env dummy mismatch: %+v", plats)
    }
}
