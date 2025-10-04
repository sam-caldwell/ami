package gpu

import "testing"

func TestOpenCLBuildProgram_FilePair(t *testing.T) { _, _ = OpenCLBuildProgram("__kernel void k(){}"); }

