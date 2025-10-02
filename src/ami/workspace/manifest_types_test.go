package workspace

import "testing"

func TestManifest_SetHasVersions(t *testing.T) {
    var m Manifest
    if m.Has("x", "1.0.0") { t.Fatalf("unexpected Has before Set") }
    m.Set("x", "1.0.0", "deadbeef")
    if !m.Has("x", "1.0.0") { t.Fatalf("expected Has after Set") }
    vs := m.Versions("x")
    if len(vs) != 1 || vs[0] != "1.0.0" { t.Fatalf("unexpected versions: %v", vs) }
}

