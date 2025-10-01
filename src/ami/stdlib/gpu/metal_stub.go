//go:build !darwin

package gpu

// Non-Darwin stubs.
func MetalAvailable() bool { return false }
func MetalDevices() []Device { return nil }

