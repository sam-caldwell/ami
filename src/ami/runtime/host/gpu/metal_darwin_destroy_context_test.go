//go:build darwin

package gpu

import "testing"

func TestMetalDarwinDestroyContextByID_FilePair(t *testing.T) { metalDestroyContextByID(0) }

