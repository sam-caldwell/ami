package driver

import (
    "encoding/json"
    "os"
    "path/filepath"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

type ssaUnit struct {
    Schema   string     `json:"schema"`
    Package  string     `json:"package"`
    Unit     string     `json:"unit"`
    Functions []ssaFunc `json:"functions"`
}

type ssaFunc struct {
    Name string   `json:"name"`
    Defs []ssaDef `json:"defs"`
}

type ssaDef struct {
    Name    string `json:"name"`
    Version int    `json:"version"`
    SSAName string `json:"ssaName"`
    Type    string `json:"type"`
}

// writeSSADebug emits a simple SSA debug view for straight-line code by versioning
// variable definitions found in Var and Assign instructions. This is a scaffold
// to aid future SSA construction and testing.
func writeSSADebug(pkg, unit string, m ir.Module) (string, error) {
    out := ssaUnit{Schema: "ssa.v1", Package: pkg, Unit: unit}
    for _, f := range m.Functions {
        sf := ssaFunc{Name: f.Name}
        versions := map[string]int{}
        // walk blocks/instructions; assign version numbers on each definition
        for _, b := range f.Blocks {
            for _, ins := range b.Instr {
                switch v := ins.(type) {
                case ir.Var:
                    n := versions[v.Name] // zero if missing
                    sf.Defs = append(sf.Defs, ssaDef{Name: v.Name, Version: n, SSAName: v.Name + "#" + itoa(n), Type: v.Type})
                    versions[v.Name] = n + 1
                case ir.Assign:
                    // destination gets a new version; use source type as hint
                    name := v.DestID
                    n := versions[name]
                    typ := ""
                    if v.Src.Type != "" { typ = v.Src.Type }
                    sf.Defs = append(sf.Defs, ssaDef{Name: name, Version: n, SSAName: name + "#" + itoa(n), Type: typ})
                    versions[name] = n + 1
                }
            }
        }
        out.Functions = append(out.Functions, sf)
    }
    dir := filepath.Join("build", "debug", "ir", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    b, err := json.MarshalIndent(out, "", "  ")
    if err != nil { return "", err }
    path := filepath.Join(dir, unit+".ssa.json")
    if err := os.WriteFile(path, b, 0o644); err != nil { return "", err }
    return path, nil
}

func itoa(n int) string {
    // small itoa without fmt for determinism/perf
    if n == 0 { return "0" }
    var buf [20]byte
    i := len(buf)
    for n > 0 {
        i--
        buf[i] = byte('0' + (n % 10))
        n /= 10
    }
    return string(buf[i:])
}

