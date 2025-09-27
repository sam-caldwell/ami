package workspace

import "testing"

func TestNormalizeImports_DedupAndTrim(t *testing.T) {
    p := &Package{Import: []string{"  a  ", "b", "a", "", "b"}}
    NormalizeImports(p)
    if len(p.Import) != 2 || p.Import[0] != "a" || p.Import[1] != "b" {
        t.Fatalf("unexpected: %+v", p.Import)
    }
}

func TestParseImportEntry_SplitsConstraint(t *testing.T) {
    path, c := ParseImportEntry("mod ^1.2.3")
    if path != "mod" || c != "^1.2.3" { t.Fatalf("got %q %q", path, c) }
    path, c = ParseImportEntry("./local")
    if path != "./local" || c != "" { t.Fatalf("got %q %q", path, c) }
    path, c = ParseImportEntry(">= 1.2.3")
    if path != ">=" || c != "1.2.3" { t.Fatalf("got %q %q", path, c) }
}

