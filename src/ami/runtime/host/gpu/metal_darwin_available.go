//go:build darwin

package gpu

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Metal -framework Foundation
#include <stdbool.h>

bool AmiMetalAvailable(void);
*/
import "C"

// MetalAvailable reports whether a Metal device is present.
func MetalAvailable() bool { return bool(C.AmiMetalAvailable()) }

