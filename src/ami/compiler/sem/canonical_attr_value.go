package sem

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

func canonicalAttrValue(name string, args []ast.Arg) string {
    // normalize value strings per attribute for conflict checks
    if name == "merge.Sort" {
        // field[/order]
        f := ""
        ord := "asc" // default asc when unspecified
        if len(args) > 0 { f = args[0].Text }
        if len(args) > 1 { ord = args[1].Text }
        return f + "/" + ord
    }
    if name == "merge.Buffer" {
        cap := ""
        pol := ""
        if len(args) > 0 { cap = args[0].Text }
        if len(args) > 1 { pol = args[1].Text }
        return cap + "/" + pol
    }
    if len(args) > 0 { return args[0].Text }
    return ""
}

