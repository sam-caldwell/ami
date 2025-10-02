package exec

import ev "github.com/sam-caldwell/ami/src/schemas/events"

// DSL stubs
func applyFilter(expr string, e ev.Event) bool {
    switch expr {
    case "", "none":
        return true
    case "drop_even":
        // drops events where payload["i"] is even
        if m, ok := e.Payload.(map[string]any); ok {
            if v, ok := m["i"].(int); ok { return v%2 != 0 }
            if f, ok := m["i"].(float64); ok { return int(f)%2 != 0 }
        }
        return true
    default:
        return true
    }
}

