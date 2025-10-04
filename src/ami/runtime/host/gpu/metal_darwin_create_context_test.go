//go:build darwin

package gpu

import "testing"

func TestMetalDarwinCreateContext_FilePair(t *testing.T) { _, _ = MetalCreateContext(Device{ID:0}) }

