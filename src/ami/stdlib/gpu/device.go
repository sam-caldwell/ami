package gpu

// Device represents a compute device for CUDA/Metal or an OpenCL device.
type Device struct {
    Backend string // "cuda" | "metal" | "opencl"
    ID      int
    Name    string
}

