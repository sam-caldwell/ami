//go:build darwin

package gpu

import "testing"

func TestMetal_Available_And_Devices(t *testing.T) {
    if !MetalAvailable() { t.Fatalf("MetalAvailable() should be true on darwin host with Metal GPU") }
    devs := MetalDevices()
    if len(devs) == 0 { t.Fatalf("MetalDevices() returned empty list") }
    for i, d := range devs {
        if d.Backend != "metal" { t.Fatalf("device[%d].Backend=%q want metal", i, d.Backend) }
        if d.ID < 0 { t.Fatalf("device[%d].ID < 0", i) }
        if d.Name == "" { t.Fatalf("device[%d] has empty Name", i) }
    }
}

