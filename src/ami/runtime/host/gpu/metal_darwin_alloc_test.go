//go:build darwin

package gpu

import "testing"

func TestMetalDarwinAlloc_FilePair(t *testing.T) { _, _ = MetalAlloc(1) }

