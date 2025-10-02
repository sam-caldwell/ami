//go:build darwin

package gpu

import "testing"

func TestMetalCreateContext_InvalidDevice_Error(t *testing.T) {
    if !MetalAvailable() { t.Skip("metal not available") }
    // Use an invalid device index to exercise error path
    if _, err := MetalCreateContext(Device{Backend: "metal", ID: -1}); err == nil {
        t.Fatalf("expected error for invalid device index")
    }
}

