//go:build darwin

package gpu

import "testing"

func TestMetalDarwinCopyFrom_FilePair(t *testing.T) {
    if err := MetalCopyFromDevice(nil, Buffer{}); err != ErrInvalidHandle {
        t.Fatalf("expected ErrInvalidHandle, got %v", err)
    }
}

