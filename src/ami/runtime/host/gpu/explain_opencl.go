package gpu

// OpenCLExplain formats a deterministic message for OpenCL.
func OpenCLExplain(op string, err error) string { return Explain("opencl", op, err) }

