//go:build darwin

package gpu

import "testing"

func TestMetalDarwinCompileLibrary_FilePair(t *testing.T) { _, _ = MetalCompileLibrary("kernel void k(){}"); }

