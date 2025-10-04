package gpu

// CudaAvailable reports whether the CUDA backend is available.
func CudaAvailable() bool { return envBoolTrue("AMI_GPU_FORCE_CUDA") }

