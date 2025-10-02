//go:build !darwin

package gpu

// MetalDevices returns nil devices on non-Darwin.
func MetalDevices() []Device { return nil }

