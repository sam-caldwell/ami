package gpu

// CudaExplain formats a deterministic message for CUDA.
func CudaExplain(op string, err error) string { return Explain("cuda", op, err) }

