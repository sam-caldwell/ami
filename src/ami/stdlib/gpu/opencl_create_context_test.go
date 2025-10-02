package gpu

import "testing"

func TestOpenCLCreateContext_FilePair(t *testing.T) { _, _ = OpenCLCreateContext(Platform{Name:"x"}) }

