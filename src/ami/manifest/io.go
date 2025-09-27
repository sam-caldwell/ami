package manifest

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
    "sort"
)

// Load reads an ami.manifest JSON document from path into the receiver.
// Unknown fields are captured into Data for forward compatibility.
func (m *Manifest) Load(path string) error {
    b, err := os.ReadFile(path)
    if err != nil { return err }
    var raw map[string]any
    if err := json.Unmarshal(b, &raw); err != nil {
        return fmt.Errorf("invalid ami.manifest: %w", err)
    }
    schema, _ := raw["schema"].(string)
    if schema == "" {
        return fmt.Errorf("missing schema")
    }
    m.Schema = schema
    // Copy remaining keys into Data
    data := make(map[string]any)
    for k, v := range raw {
        if k == "schema" { continue }
        data[k] = v
    }
    m.Data = data
    return nil
}

// Save writes the manifest to path using deterministic top-level key ordering.
// The file is created or truncated with 0644 permissions.
func (m Manifest) Save(path string) error {
    if m.Schema == "" {
        m.Schema = "ami.manifest/v1"
    }
    if m.Data == nil {
        m.Data = map[string]any{}
    }
    f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
    if err != nil { return err }
    w := bufio.NewWriter(f)
    // Write opening and schema first
    _, _ = w.WriteString("{\n  \"schema\": \"")
    _, _ = w.WriteString(m.Schema)
    _, _ = w.WriteString("\"")
    // Sort remaining top-level keys
    keys := make([]string, 0, len(m.Data))
    for k := range m.Data { keys = append(keys, k) }
    sort.Strings(keys)
    for _, k := range keys {
        _, _ = w.WriteString(",\n  \"")
        _, _ = w.WriteString(k)
        _, _ = w.WriteString("\": ")
        // Marshal each value deterministically via json (nested map ordering is not guaranteed).
        enc, err := json.Marshal(m.Data[k])
        if err != nil { _ = f.Close(); return err }
        _, _ = w.Write(enc)
    }
    _, _ = w.WriteString("\n}\n")
    if err := w.Flush(); err != nil { _ = f.Close(); return err }
    return f.Close()
}

