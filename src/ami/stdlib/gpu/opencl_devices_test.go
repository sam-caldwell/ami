package gpu

import "testing"

func TestOpenCL_Devices_EnvForcedDummy(t *testing.T) {
    // default: not forced -> nil
    if ds := OpenCLDevices(Platform{Name: "X"}); ds != nil {
        t.Fatalf("OpenCLDevices default should be nil, got %+v", ds)
    }
    t.Setenv("AMI_GPU_FORCE_OPENCL", "1")
    plats := OpenCLPlatforms()
    if len(plats) == 0 { t.Skip("platforms empty under env force; unexpected") }
    ds := OpenCLDevices(plats[0])
    if len(ds) != 1 || ds[0].Backend != "opencl" || ds[0].ID != 0 || ds[0].Name == "" {
        t.Fatalf("OpenCLDevices env dummy mismatch: %+v", ds)
    }
}

