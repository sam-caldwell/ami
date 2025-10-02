//go:build darwin

package gpu

import "testing"

func TestMetalDarwinReleasePipeline_FilePair(t *testing.T) { metalReleasePipeline(0) }

