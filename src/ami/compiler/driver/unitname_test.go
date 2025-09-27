package driver

import "testing"

func TestUnitName_StripsExtension(t *testing.T) {
    if unitName("/x/y/z/main.ami") != "main" { t.Fatalf("basic strip") }
    if unitName("/x/y/z/Makefile") != "Makefile" { t.Fatalf("no dot remains unchanged") }
    if unitName("main.ami.go") != "main.ami" { t.Fatalf("last dot only") }
}

