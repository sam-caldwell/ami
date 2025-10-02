package merge

// SetSensitiveZeroizer installs a package-level zeroizer invoked on buffer drops.
// Passing nil disables zeroization.
func SetSensitiveZeroizer(fn func(any) int) { sensitiveZeroizer = fn }

