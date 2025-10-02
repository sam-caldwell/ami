package ascii

import (
    "strings"
    "github.com/sam-caldwell/ami/src/schemas/graph"
)

// edgeTag formats selected edge attributes for inline display.
func edgeTag(e graph.Edge) string {
    if len(e.Attrs) == 0 { return "" }
    var tags []string
    if v, ok := e.Attrs["bounded"].(bool); ok && v { tags = append(tags, "bounded") }
    if v, ok := e.Attrs["delivery"].(string); ok && v != "" && v != "atLeastOnce" { tags = append(tags, v) }
    if v, ok := e.Attrs["type"].(string); ok && v != "" { tags = append(tags, "type:"+v) }
    if v, ok := e.Attrs["multipath"].(string); ok && v != "" { tags = append(tags, "mp:"+v) }
    if len(tags) == 0 { return "" }
    return strings.Join(tags, ",")
}

