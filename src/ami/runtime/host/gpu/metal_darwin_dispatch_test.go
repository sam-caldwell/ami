//go:build darwin

package gpu

import "testing"

func TestMetalDarwinDispatch_FilePair(t *testing.T) {
    if err := MetalDispatch(Context{}, Pipeline{}, [3]uint32{1,1,1}, [3]uint32{1,1,1}); err != ErrInvalidHandle {
        t.Fatalf("expected ErrInvalidHandle, got %v", err)
    }
}

