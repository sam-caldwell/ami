//go:build darwin

package gpu

import "testing"

func TestMetalDarwinFree_FilePair(t *testing.T) { _ = MetalFree(Buffer{backend:"metal", valid:true, bufId:1}) }

