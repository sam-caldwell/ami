package gpu

// OpenCLAvailable reports whether the OpenCL backend is available.
func OpenCLAvailable() bool { return envBoolTrue("AMI_GPU_FORCE_OPENCL") }

