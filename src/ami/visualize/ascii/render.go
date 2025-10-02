package ascii

import (
    "fmt"
    "github.com/sam-caldwell/ami/src/schemas/graph"
)

// Render returns a two-part ASCII representation: a header line and the graph body.
// The header follows the CLI expectation: "package: <pkg>  pipeline: <name>\n".
// The body uses RenderLine or RenderBlock depending on options.
func Render(g graph.Graph, opt Options) (string, string) {
    header := fmt.Sprintf("package: %s  pipeline: %s\n", g.Package, g.Name)
    // Prefer block rendering when a width is given or when legend/focus options requested.
    if opt.Width > 0 || opt.Legend || opt.Focus != "" || opt.ShowIDs || opt.EdgeIDs || opt.Color {
        return header, RenderBlock(g, opt)
    }
    return header, RenderLine(g, opt) + "\n"
}

