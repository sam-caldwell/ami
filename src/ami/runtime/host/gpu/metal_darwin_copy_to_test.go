//go:build darwin

package gpu

import "testing"

func TestMetalDarwinCopyTo_FilePair(t *testing.T) {
    if err := MetalCopyToDevice(Buffer{}, nil); err != ErrInvalidHandle {
        t.Fatalf("expected ErrInvalidHandle, got %v", err)
    }
}

