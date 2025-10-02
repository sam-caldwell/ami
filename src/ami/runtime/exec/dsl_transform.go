package exec

import ev "github.com/sam-caldwell/ami/src/schemas/events"

func applyTransform(expr string, e ev.Event) ev.Event {
    switch expr {
    case "", "none":
        return e
    default:
        // add_field:name sets payload[name]=true
        if len(expr) > 10 && expr[:10] == "add_field:" {
            key := expr[10:]
            if m, ok := e.Payload.(map[string]any); ok && key != "" { m[key] = true; e.Payload = m }
        }
        return e
    }
}

