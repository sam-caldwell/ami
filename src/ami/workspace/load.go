package workspace

import (
    "os"
    "gopkg.in/yaml.v3"
)

// Load reads YAML from path into the workspace receiver.
func (w *Workspace) Load(path string) error {
    b, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    if err := yaml.Unmarshal(b, w); err != nil { return err }
    // Normalize compiler.env: eliminate duplicates preserving first occurrence order.
    if len(w.Toolchain.Compiler.Env) > 1 {
        seen := make(map[string]struct{}, len(w.Toolchain.Compiler.Env))
        out := make([]string, 0, len(w.Toolchain.Compiler.Env))
        for _, e := range w.Toolchain.Compiler.Env {
            if _, ok := seen[e]; ok { continue }
            seen[e] = struct{}{}
            out = append(out, e)
        }
        w.Toolchain.Compiler.Env = out
    }
    return nil
}
