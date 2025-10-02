package gpu

// MetalExplain formats a deterministic message for Metal.
func MetalExplain(op string, err error) string { return Explain("metal", op, err) }

