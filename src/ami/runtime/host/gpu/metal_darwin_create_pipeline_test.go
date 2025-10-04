//go:build darwin

package gpu

import "testing"

func TestMetalDarwinCreatePipeline_FilePair(t *testing.T) { _, _ = MetalCreatePipeline(Library{valid:true, libId:1}, "k") }

