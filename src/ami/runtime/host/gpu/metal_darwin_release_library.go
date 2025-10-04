//go:build darwin

package gpu

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Metal -framework Foundation

void AmiMetalReleaseLibrary(int libId);
*/
import "C"

// internal helper for Go-level Release()
func metalReleaseLibrary(id int) { C.AmiMetalReleaseLibrary(C.int(id)) }

