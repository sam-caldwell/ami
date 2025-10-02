package driver

import "github.com/sam-caldwell/ami/src/ami/compiler/ir"

func collectExports(m ir.Module) []string {
    var names []string
    for _, f := range m.Functions { names = append(names, f.Name) }
    sortStrings(names)
    return names
}

