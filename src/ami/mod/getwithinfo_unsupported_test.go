package mod

import "testing"

func TestGetWithInfo_UnsupportedURL_Error(t *testing.T) {
    if _, _, _, err := GetWithInfo("http://example.com/repo#v1.0.0"); err == nil {
        t.Fatalf("expected error for unsupported scheme")
    }
}

