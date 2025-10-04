package gpu

// CudaDevices lists CUDA devices in stub mode.
func CudaDevices() []Device {
    if CudaAvailable() {
        return []Device{{Backend: "cuda", ID: 0, Name: "cuda-ci-0"}}
    }
    return nil
}

