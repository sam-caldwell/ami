package ascii

import (
    "sort"
    "strings"

    "github.com/sam-caldwell/ami/src/schemas/graph"
)

// RenderBlock renders a minimal multi-line ASCII block for a graph.
// Line 1 is always the primary straight chain (best-effort). Additional
// lines render simple branches from nodes with multiple outgoing edges.
// Layout is deterministic and stable across runs.
func RenderBlock(g graph.Graph, opt Options) string {
    // Build adjacency list (multi)
    outs := map[string][]string{}
    ids := map[string]graph.Node{}
    for _, n := range g.Nodes { ids[n.ID] = n }
    for _, e := range g.Edges { outs[e.From] = append(outs[e.From], e.To) }
    // Choose a start: node with in-degree 0 if present
    indeg := map[string]int{}
    for _, n := range g.Nodes { indeg[n.ID] = 0 }
    for _, e := range g.Edges { indeg[e.To]++ }
    start := ""
    for id := range ids { if indeg[id] == 0 { start = id; break } }
    if start == "" && len(g.Nodes) > 0 { start = g.Nodes[0].ID }
    // Determine primary chain by following the lexicographically-smallest next
    chain := []string{}
    seen := map[string]bool{}
    cur := start
    for cur != "" && !seen[cur] {
        seen[cur] = true
        chain = append(chain, cur)
        nexts := outs[cur]
        if len(nexts) == 0 { break }
        sort.Strings(nexts)
        cur = nexts[0]
    }
    // Fallback: if we couldn't build a chain, just sort ids and render line
    if len(chain) == 0 {
        return RenderLine(g, opt)
    }
    // Build the primary chain text and track token start columns
    token := func(id string) string {
        n := ids[id]
        lbl := strings.TrimSpace(n.Label)
        if lbl == "" { lbl = n.Kind }
        if opt.ShowIDs {
            // include the instance ID as prefix: e.g., 00:ingress
            lbl = id + ":" + lbl
        }
        // focus highlighting
        if opt.Focus != "" && containsFold(lbl, opt.Focus) {
            if opt.Color {
                // bright yellow
                lbl = "\x1b[33m" + lbl + "\x1b[0m"
            } else {
                lbl = "*" + lbl + "*"
            }
        }
        switch strings.ToLower(n.Kind) {
        case "ingress", "egress":
            return "[" + lbl + "]"
        default:
            return "(" + lbl + ")"
        }
    }
    parts := make([]string, 0, len(chain))
    starts := make([]int, 0, len(chain))
    col := 0
    for i, id := range chain {
        if i > 0 {
            // choose arrow; include edge attrs when available
            arrow := " --> "
            from := chain[i-1]
            to := id
            if e, ok := findEdge(g, from, to); ok {
                tags := edgeTag(e)
                if opt.EdgeIDs { if tags != "" { tags += ";" }; tags += from + "->" + to }
                if tags != "" { arrow = " --[" + tags + "]--> " }
            }
            parts = append(parts, arrow)
            col += len(arrow)
        }
        starts = append(starts, col)
        t := token(id)
        parts = append(parts, t)
        col += len(t)
    }
    var lines []string
    if opt.Width > 0 {
        lines = append(lines, wrapTokens(parts, opt.Width))
    } else {
        lines = append(lines, strings.Join(parts, ""))
    }
    // Render simple branches: for any chain node with extra outs beyond the primary
    for idx, id := range chain {
        nexts := outs[id]
        if len(nexts) <= 1 { continue }
        // branch targets excluding the chosen primary next
        sort.Strings(nexts)
        primary := nexts[0]
        for _, to := range nexts[1:] {
            // two-line branch: '|' line then '+--> (label)'
            pad := strings.Repeat(" ", starts[idx])
            lines = append(lines, wrapLine(pad+"|", opt.Width))
            // branch arrow may include tags from edge
            arr := "+--> "
            if e, ok := findEdge(g, id, to); ok {
                if tag := edgeTag(e); tag != "" { arr = "+-[" + tag + "]-> " }
            }
            // extend dashed alignment towards the target when target is on the primary chain
            // determine target start column if present
            targetStart := -1
            for j, cid := range chain {
                if cid == to { targetStart = starts[j]; break }
            }
            if targetStart > 0 {
                dashLen := targetStart - starts[idx] - 3
                if dashLen < 1 { dashLen = 1 }
                lines = append(lines, wrapLine(strings.Repeat(" ", starts[idx])+"+"+strings.Repeat("-", dashLen)+"> "+token(to), opt.Width))
            } else {
                lines = append(lines, wrapLine(pad+arr+token(to), opt.Width))
            }
        }
        _ = primary
    }
    if opt.Legend {
        legend := "legend: [] ingress/egress, () worker; * focus"
        lines = append([]string{legend}, lines...)
    }
    return strings.Join(lines, "\n") + "\n"
}

