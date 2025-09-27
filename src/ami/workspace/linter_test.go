package workspace

import (
    "gopkg.in/yaml.v3"
    "strings"
    "testing"
)

func TestLinter_YAMLShape(t *testing.T) {
    l := Linter{Options: []string{"strict"}, Rules: map[string]string{"W_IMPORT_ORDER":"warn"}}
    b, err := yaml.Marshal(l)
    if err != nil { t.Fatalf("yaml: %v", err) }
    s := string(b)
    if !strings.Contains(s, "options:") || !strings.Contains(s, "rules:") {
        t.Fatalf("expected options/rules keys: %s", s)
    }
}

