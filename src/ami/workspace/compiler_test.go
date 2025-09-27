package workspace

import (
    "gopkg.in/yaml.v3"
    "strings"
    "testing"
)

func TestCompiler_YAMLKeys_BasenamePair(t *testing.T) {
    c := Compiler{Concurrency: "NUM_CPU", Target: "./build", Env: []string{"darwin/arm64"}}
    b, err := yaml.Marshal(c)
    if err != nil { t.Fatalf("yaml: %v", err) }
    s := string(b)
    if !strings.Contains(s, "concurrency:") || !strings.Contains(s, "target:") { t.Fatalf("missing keys: %s", s) }
}

