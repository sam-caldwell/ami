package manifest

import "testing"

func TestManifest_ZeroValueAndSetters(t *testing.T) {
    var m Manifest
    if m.Schema != "" || m.Data != nil { t.Fatalf("zero value mismatch: %+v", m) }
    m.Schema = "ami.manifest/v1"
    m.Data = map[string]any{"k": "v"}
    if m.Schema != "ami.manifest/v1" || m.Data["k"].(string) != "v" {
        t.Fatalf("unexpected values: %+v", m)
    }
}

