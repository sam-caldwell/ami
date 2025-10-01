package llvm

// abiType maps AMI surface types to LLVM ABI-safe types for public function signatures.
// It removes raw pointer exposure by mapping any pointer-like/classified type to an i64 handle.
func abiType(t string) string {
	mt := mapType(t)
	if mt == "ptr" {
		return "i64"
	}
	return mt
}
