//go:build darwin

package gpu

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Metal -framework Foundation

void AmiMetalFreeBuffer(int bufId);
*/
import "C"

func metalFreeBufferByID(id int) { C.AmiMetalFreeBuffer(C.int(id)) }

