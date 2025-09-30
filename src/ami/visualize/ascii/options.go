package ascii

// Options carries ASCII renderer options. This is a placeholder for
// future width/color/legend knobs; kept minimal for now.
type Options struct {
    // Width is the maximum line width for wrapping. When 0, no wrapping occurs.
    Width int
    // Focus highlights nodes whose label/kind contains this substring (case-insensitive).
    Focus string
    // Legend, when true, renders a one-line legend above the block.
    Legend bool
    // Color enables ANSI color for focus highlighting (human mode only).
    Color bool
    // ShowIDs prefixes node labels with their instance IDs (e.g., 00:ingress).
    ShowIDs bool
    // EdgeIDs labels arrows with fromId->toId (e.g., [00->01]).
    EdgeIDs bool
}
