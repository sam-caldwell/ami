//go:build darwin

package gpu

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Metal -framework Foundation

void AmiMetalReleasePipeline(int pipeId);
*/
import "C"

func metalReleasePipeline(id int) { C.AmiMetalReleasePipeline(C.int(id)) }

