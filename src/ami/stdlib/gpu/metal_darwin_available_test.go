//go:build darwin

package gpu

import "testing"

func TestMetalDarwinAvailable_FilePair(t *testing.T) { _ = MetalAvailable() }

