package sem

import (
    "strconv"
    "strings"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeConcurrency validates concurrency pragmas:
//   #pragma concurrency:workers N
//   #pragma concurrency:schedule <policy>
// workers must be >=1; schedule in {fifo, lifo, fair, worksteal} (extensible).
func AnalyzeConcurrency(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    for _, pr := range f.Pragmas {
        if pr.Domain != "concurrency" { continue }
        switch pr.Key {
        case "workers":
            // Value or params["count"]
            s := strings.TrimSpace(pr.Value)
            if s == "" { if v, ok := pr.Params["count"]; ok { s = strings.TrimSpace(v) } }
            n, err := strconv.Atoi(s)
            if err != nil || n < 1 {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_CONCURRENCY_WORKERS_INVALID", Message: "concurrency workers must be >=1", Pos: &diag.Position{Line: pr.Pos.Line, Column: pr.Pos.Column, Offset: pr.Pos.Offset}})
            }
        case "schedule":
            pol := strings.ToLower(strings.TrimSpace(pr.Value))
            switch pol {
            case "fifo", "lifo", "fair", "worksteal":
                // ok
            default:
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_CONCURRENCY_SCHEDULE_INVALID", Message: "invalid concurrency schedule", Pos: &diag.Position{Line: pr.Pos.Line, Column: pr.Pos.Column, Offset: pr.Pos.Offset}})
            }
        }
    }
    return out
}

