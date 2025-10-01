package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

type irSymbolsIndexUnit struct {
    Unit    string   `json:"unit"`
    Exports []string `json:"exports,omitempty"`
    Externs []string `json:"externs,omitempty"`
}

type irSymbolsIndex struct {
    Schema  string               `json:"schema"`
    Package string               `json:"package"`
    Units   []irSymbolsIndexUnit `json:"units"`
}

func collectExports(m ir.Module) []string {
    var names []string
    for _, f := range m.Functions { names = append(names, f.Name) }
    sortStrings(names)
    return names
}

// collectExterns scans IR for operations that imply runtime extern usage and
// returns a deterministic list of extern symbol names (not full LLVM decls).
// This mirrors the minimal extern collection in the LLVM emitter.
func collectExterns(m ir.Module) []string {
    seen := map[string]bool{}
    add := func(s string) { if s != "" { seen[s] = true } }
    for _, f := range m.Functions {
        for _, b := range f.Blocks {
            for _, ins := range b.Instr {
                if ex, ok := ins.(ir.Expr); ok {
                    op := ex.Op
                    if op == "panic" { add("ami_rt_panic") }
                    if op == "alloc" || ex.Callee == "ami_rt_alloc" { add("ami_rt_alloc") }
                    // Owned/zeroize helpers surfaced as calls
                    switch ex.Callee {
                    case "ami_rt_zeroize":
                        add("ami_rt_zeroize")
                    case "ami_rt_owned_len":
                        add("ami_rt_owned_len")
                    case "ami_rt_owned_ptr":
                        add("ami_rt_owned_ptr")
                    case "ami_rt_owned_new":
                        add("ami_rt_owned_new")
                    case "ami_rt_zeroize_owned":
                        add("ami_rt_zeroize_owned")
                    case "ami_rt_sleep_ms":
                        add("ami_rt_sleep_ms")
                    case "ami_rt_time_now":
                        add("ami_rt_time_now")
                    case "ami_rt_time_add":
                        add("ami_rt_time_add")
                    case "ami_rt_time_delta":
                        add("ami_rt_time_delta")
                    case "ami_rt_time_unix":
                        add("ami_rt_time_unix")
                    case "ami_rt_time_unix_nano":
                        add("ami_rt_time_unix_nano")
                    }
                } else if d, ok := ins.(ir.Defer); ok {
                    // Include deferred calls in extern set
                    ex := d.Expr
                    if strings.ToLower(ex.Op) == "call" {
                        switch ex.Callee {
                        case "ami_rt_panic":
                            add("ami_rt_panic")
                        case "ami_rt_alloc":
                            add("ami_rt_alloc")
                        case "ami_rt_zeroize":
                            add("ami_rt_zeroize")
                        case "ami_rt_owned_len":
                            add("ami_rt_owned_len")
                        case "ami_rt_owned_ptr":
                            add("ami_rt_owned_ptr")
                        case "ami_rt_owned_new":
                            add("ami_rt_owned_new")
                        case "ami_rt_zeroize_owned":
                            add("ami_rt_zeroize_owned")
                        case "ami_rt_sleep_ms":
                            add("ami_rt_sleep_ms")
                        case "ami_rt_time_now":
                            add("ami_rt_time_now")
                        case "ami_rt_time_add":
                            add("ami_rt_time_add")
                        case "ami_rt_time_delta":
                            add("ami_rt_time_delta")
                        case "ami_rt_time_unix":
                            add("ami_rt_time_unix")
                        case "ami_rt_time_unix_nano":
                            add("ami_rt_time_unix_nano")
                        }
                    }
                }
            }
        }
    }
    var out []string
    for k := range seen { out = append(out, k) }
    sortStrings(out)
    return out
}

// writeIRSymbolsIndex writes ir.symbols.index.json for a package.
func writeIRSymbolsIndex(pkg string, units []irSymbolsIndexUnit) (string, error) {
    idx := irSymbolsIndex{Schema: "ir.symbols.index.v1", Package: pkg, Units: units}
    dir := filepath.Join("build", "debug", "ir", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    b, err := json.MarshalIndent(idx, "", "  ")
    if err != nil { return "", err }
    out := filepath.Join(dir, "ir.symbols.index.json")
    if err := os.WriteFile(out, b, 0o644); err != nil { return "", err }
    return out, nil
}
