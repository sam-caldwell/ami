package workspace

import (
    "gopkg.in/yaml.v3"
    "strings"
    "testing"
)

func TestCompiler_YAMLKeysPresent(t *testing.T) {
    c := Compiler{Concurrency: "1", Target: "./out", Env: []string{"linux/amd64"}, Options: []string{"verbose"}}
    b, err := yaml.Marshal(c)
    if err != nil { t.Fatalf("yaml marshal: %v", err) }
    s := string(b)
    for _, k := range []string{"concurrency:", "target:", "env:", "options:"} {
        if !strings.Contains(s, k) { t.Fatalf("expected %q in YAML: %s", k, s) }
    }
}

