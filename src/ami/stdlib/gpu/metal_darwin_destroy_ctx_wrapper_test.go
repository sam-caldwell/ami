//go:build darwin

package gpu

import "testing"

func TestMetalDarwinDestroyContext_FilePair(t *testing.T) { _ = MetalDestroyContext(Context{backend:"metal", valid:true, ctxId:1}) }

