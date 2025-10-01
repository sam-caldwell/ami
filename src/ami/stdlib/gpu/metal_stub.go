//go:build !darwin

package gpu

// Non-Darwin stubs.
func MetalAvailable() bool { return false }
func MetalDevices() []Device { return nil }

// helpers used by generic Release() paths
func metalReleaseLibrary(id int) {}
func metalReleasePipeline(id int) {}
func metalDestroyContextByID(id int) {}
func metalFreeBufferByID(id int) {}
