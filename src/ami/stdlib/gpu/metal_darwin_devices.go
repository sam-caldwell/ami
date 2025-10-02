//go:build darwin

package gpu

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Metal -framework Foundation
#include <stdbool.h>
#include <stdlib.h>

int AmiMetalDeviceCount(void);
char* AmiMetalDeviceNameAt(int idx);
void AmiMetalFreeCString(char* p);
*/
import "C"

// MetalDevices enumerates Metal devices by index with names.
func MetalDevices() []Device {
    n := int(C.AmiMetalDeviceCount())
    if n <= 0 { return nil }
    out := make([]Device, 0, n)
    for i := 0; i < n; i++ {
        nameC := C.AmiMetalDeviceNameAt(C.int(i))
        name := ""
        if nameC != nil {
            name = C.GoString(nameC)
            C.AmiMetalFreeCString(nameC)
        }
        out = append(out, Device{Backend: "metal", ID: i, Name: name})
    }
    return out
}

