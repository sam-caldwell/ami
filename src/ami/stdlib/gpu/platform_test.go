package gpu

import "testing"

func TestPlatform_FilePair(t *testing.T) {
    p := Platform{Vendor: "v", Name: "n", Version: "1"}
    if p.Name == "" { t.Fatalf("empty name") }
}

