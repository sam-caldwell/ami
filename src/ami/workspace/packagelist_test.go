package workspace

import (
    "gopkg.in/yaml.v3"
    "testing"
)

func TestPackageList_MarshalYAML_SequenceOfSingleEntryMaps(t *testing.T) {
    l := PackageList{
        {Key: "main", Package: Package{Name: "app", Version: "0.0.1", Root: "./src"}},
        {Key: "util", Package: Package{Name: "util", Version: "1.2.3", Root: "./util"}},
    }
    b, err := yaml.Marshal(map[string]any{"packages": l})
    if err != nil { t.Fatalf("yaml marshal: %v", err) }
    s := string(b)
    // Expect structure:
    // packages:
    //   - main:
    //   - util:
    if !(containsLine(s, "packages:") && containsLine(s, "- main:") && containsLine(s, "- util:")) {
        t.Fatalf("unexpected YAML shape:\n%s", s)
    }
}

func TestPackageList_UnmarshalYAML_RoundTrip(t *testing.T) {
    in := []byte("packages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n  - util:\n      name: util\n      version: 1.2.3\n      root: ./util\n")
    var tmp struct{ Packages PackageList `yaml:"packages"` }
    if err := yaml.Unmarshal(in, &tmp); err != nil { t.Fatalf("yaml unmarshal: %v", err) }
    if len(tmp.Packages) != 2 || tmp.Packages[0].Key != "main" || tmp.Packages[1].Key != "util" {
        t.Fatalf("unexpected parsed packages: %#v", tmp.Packages)
    }
}

func containsLine(s, line string) bool {
    for _, ln := range splitLines(s) {
        if ln == line { return true }
    }
    return false
}
func splitLines(s string) []string {
    var out []string
    start := 0
    for i := 0; i < len(s); i++ {
        if s[i] == '\n' {
            out = append(out, s[start:i])
            start = i+1
        }
    }
    if start <= len(s) { out = append(out, s[start:]) }
    return out
}

