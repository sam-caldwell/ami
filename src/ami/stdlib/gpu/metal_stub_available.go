//go:build !darwin

package gpu

// MetalAvailable reports availability on non-Darwin.
func MetalAvailable() bool { return false }

