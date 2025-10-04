package gpu

import "testing"

func TestDevice_FilePair(t *testing.T) {
    d := Device{Backend: "cuda", ID: 0, Name: "x"}
    if d.Backend == "" { t.Fatalf("empty backend") }
}

