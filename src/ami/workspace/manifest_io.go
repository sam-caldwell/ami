package workspace

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
    "sort"
)

// Load reads ami.sum from path and populates the manifest, accepting either:
// - object form: packages: { "name": {"version": "v1.2.3", "sha256": "...", "source":"..."} }
// - nested form: packages: { "name": { "v1.2.3": "<sha256>", ... } }
func (m *Manifest) Load(path string) error {
    b, err := os.ReadFile(path)
    if err != nil { return err }
    var raw map[string]any
    if err := json.Unmarshal(b, &raw); err != nil { return fmt.Errorf("invalid ami.sum: %w", err) }
    schema, _ := raw["schema"].(string)
    if schema == "" { return fmt.Errorf("missing schema") }
    m.Schema = schema
    m.Packages = make(map[string]map[string]string)
    pkgs, ok := raw["packages"]
    if !ok { return nil }
    switch t := pkgs.(type) {
    case map[string]any:
        for name, v := range t {
            if mm, ok := v.(map[string]any); ok {
                // object or nested
                if ver, okv := mm["version"].(string); okv {
                    sha, _ := mm["sha256"].(string)
                    if m.Packages[name] == nil { m.Packages[name] = map[string]string{} }
                    m.Packages[name][ver] = sha
                    continue
                }
                // nested by version
                for ver, x := range mm {
                    if sha, ok := x.(string); ok {
                        if m.Packages[name] == nil { m.Packages[name] = map[string]string{} }
                        m.Packages[name][ver] = sha
                    }
                }
            }
        }
    case []any:
        // array form: [{name, version, sha256, ...}]
        for _, el := range t {
            if mm, ok := el.(map[string]any); ok {
                name, _ := mm["name"].(string)
                ver, _ := mm["version"].(string)
                sha, _ := mm["sha256"].(string)
                if name != "" && ver != "" {
                    if m.Packages[name] == nil { m.Packages[name] = map[string]string{} }
                    m.Packages[name][ver] = sha
                }
            }
        }
    }
    return nil
}

// Save writes the manifest to path using a canonical, deterministic JSON layout:
//   schema: ami.sum/v1
//   packages: { packageName: { version: sha256, ... }, ... }
// Keys are sorted lexicographically.
func (m *Manifest) Save(path string) error {
    if m.Schema == "" { m.Schema = "ami.sum/v1" }
    if m.Packages == nil { m.Packages = map[string]map[string]string{} }
    // Build JSON manually to ensure deterministic key order
    f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
    if err != nil { return err }
    w := bufio.NewWriter(f)
    _, _ = w.WriteString("{\n  \"schema\": \"")
    _, _ = w.WriteString(m.Schema)
    _, _ = w.WriteString("\",\n  \"packages\": {")
    // sort package names
    var names []string
    for name := range m.Packages { names = append(names, name) }
    sort.Strings(names)
    for i, name := range names {
        if i > 0 { _, _ = w.WriteString(",") }
        _, _ = w.WriteString("\n    \"")
        _, _ = w.WriteString(name)
        _, _ = w.WriteString("\": {")
        // sort versions
        var versions []string
        for ver := range m.Packages[name] { versions = append(versions, ver) }
        sort.Strings(versions)
        for j, ver := range versions {
            if j > 0 { _, _ = w.WriteString(",") }
            _, _ = w.WriteString("\n      \"")
            _, _ = w.WriteString(ver)
            _, _ = w.WriteString("\": \"")
            _, _ = w.WriteString(m.Packages[name][ver])
            _, _ = w.WriteString("\"")
        }
        if len(versions) > 0 { _, _ = w.WriteString("\n    }") } else { _, _ = w.WriteString("}") }
    }
    if len(names) > 0 { _, _ = w.WriteString("\n  }") } else { _, _ = w.WriteString("}") }
    _, _ = w.WriteString("\n}\n")
    if err := w.Flush(); err != nil { _ = f.Close(); return err }
    return f.Close()
}

