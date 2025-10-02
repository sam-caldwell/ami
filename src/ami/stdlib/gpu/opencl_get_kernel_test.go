package gpu

import "testing"

func TestOpenCLGetKernel_FilePair(t *testing.T) { _, _ = OpenCLGetKernel(Program{}, "k") }

