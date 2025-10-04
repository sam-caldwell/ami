//go:build darwin

package gpu

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Metal -framework Foundation

void AmiMetalContextDestroy(int ctxId);
*/
import "C"

func metalDestroyContextByID(id int) { C.AmiMetalContextDestroy(C.int(id)) }

