package graph

import (
    "bytes"
    "encoding/json"
    "sort"
)

// Graph is a pipeline graph representation for JSON emission.
// Field ordering in MarshalJSON is deterministic.
type Graph struct {
    Package string
    Unit    string
    Name    string
    Nodes   []Node
    Edges   []Edge
}

// MarshalJSON renders Graph with stable key order and sorted nodes/edges.
func (g Graph) MarshalJSON() ([]byte, error) {
    // Copy and sort to avoid mutating the receiver
    nodes := make([]Node, len(g.Nodes))
    copy(nodes, g.Nodes)
    sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
    edges := make([]Edge, len(g.Edges))
    copy(edges, g.Edges)
    sort.Slice(edges, func(i, j int) bool {
        if edges[i].From == edges[j].From { return edges[i].To < edges[j].To }
        return edges[i].From < edges[j].From
    })

    var buf bytes.Buffer
    buf.WriteByte('{')
    // schema first
    buf.WriteString("\"schema\":\"")
    buf.WriteString(Schema)
    buf.WriteString("\"")
    // package
    buf.WriteString(",\"package\":")
    pb, _ := json.Marshal(g.Package)
    buf.Write(pb)
    // unit
    buf.WriteString(",\"unit\":")
    ub, _ := json.Marshal(g.Unit)
    buf.Write(ub)
    // name
    buf.WriteString(",\"name\":")
    nb, _ := json.Marshal(g.Name)
    buf.Write(nb)
    // nodes
    buf.WriteString(",\"nodes\":[")
    for i, n := range nodes {
        if i > 0 { buf.WriteByte(',') }
        buf.WriteByte('{')
        buf.WriteString("\"id\":")
        b, _ := json.Marshal(n.ID)
        buf.Write(b)
        buf.WriteString(",\"kind\":")
        b, _ = json.Marshal(n.Kind)
        buf.Write(b)
        buf.WriteString(",\"label\":")
        b, _ = json.Marshal(n.Label)
        buf.Write(b)
        buf.WriteByte('}')
    }
    buf.WriteByte(']')
    // edges
    buf.WriteString(",\"edges\":[")
    for i, e := range edges {
        if i > 0 { buf.WriteByte(',') }
        buf.WriteByte('{')
        buf.WriteString("\"from\":")
        b, _ := json.Marshal(e.From)
        buf.Write(b)
        buf.WriteString(",\"to\":")
        b, _ = json.Marshal(e.To)
        buf.Write(b)
        if len(e.Attrs) > 0 {
            buf.WriteString(",\"attrs\":{")
            keys := make([]string, 0, len(e.Attrs))
            for k := range e.Attrs { keys = append(keys, k) }
            sort.Strings(keys)
            for j, k := range keys {
                if j > 0 { buf.WriteByte(',') }
                kb, _ := json.Marshal(k)
                buf.Write(kb)
                buf.WriteByte(':')
                vb, _ := json.Marshal(e.Attrs[k])
                buf.Write(vb)
            }
            buf.WriteByte('}')
        }
        buf.WriteByte('}')
    }
    buf.WriteByte(']')
    buf.WriteByte('}')
    return buf.Bytes(), nil
}

