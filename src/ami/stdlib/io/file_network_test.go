package io

import (
    "testing"
)

func TestHostnameAndInterfaces(t *testing.T) {
    hn, err := Hostname()
    if err != nil { t.Fatalf("Hostname error: %v", err) }
    if hn == "" { t.Fatalf("Hostname should not be empty") }

    ifs, err := Interfaces()
    if err != nil { t.Fatalf("Interfaces error: %v", err) }
    if len(ifs) == 0 { t.Fatalf("Interfaces should not be empty") }
}

