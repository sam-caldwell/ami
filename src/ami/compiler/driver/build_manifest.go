package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type bmUnit struct {
    Unit      string `json:"unit"`
    IR        string `json:"ir,omitempty"`
    Pipelines string `json:"pipelines,omitempty"`
    EventMeta string `json:"eventmeta,omitempty"`
    ASM       string `json:"asm,omitempty"`
    AST       string `json:"ast,omitempty"`
    Sources   string `json:"sources,omitempty""
}

type bmPackage struct {
    Name       string   `json:"name"`
    Units      []bmUnit `json:"units"`
    EdgesIndex string   `json:"edgesIndex,omitempty"`
    AsmIndex   string   `json:"asmIndex,omitempty"`
}

type BuildManifest struct {
    Schema   string      `json:"schema"`
    Packages []bmPackage `json:"packages"`
}

func writeBuildManifest(m BuildManifest) (string, error) {
    if m.Schema == "" { m.Schema = "manifest.v1" }
    dir := filepath.Join("build", "debug")
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    b, err := json.MarshalIndent(m, "", "  ")
    if err != nil { return "", err }
    out := filepath.Join(dir, "manifest.json")
    if err := os.WriteFile(out, b, 0o644); err != nil { return "", err }
    return out, nil
}

