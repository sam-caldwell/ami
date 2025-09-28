package sem

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// ComputeSCC returns a map from function name to the set of functions in its SCC (including itself).
// It builds a simple call graph from the AST file and uses Kosaraju's algorithm.
func ComputeSCC(f *ast.File) map[string]map[string]bool {
    scc := map[string]map[string]bool{}
    if f == nil { return scc }
    // Collect function names and adjacency
    var names []string
    index := map[string]int{}
    for _, d := range f.Decls {
        if fn, ok := d.(*ast.FuncDecl); ok {
            index[fn.Name] = len(names)
            names = append(names, fn.Name)
        }
    }
    n := len(names)
    adj := make([][]int, n)
    radj := make([][]int, n)
    // Build edges
    for _, d := range f.Decls {
        fn, ok := d.(*ast.FuncDecl)
        if !ok || fn.Body == nil { continue }
        from, ok := index[fn.Name]; if !ok { continue }
        // Collect calls in body
        var walkExpr func(e ast.Expr)
        walkExpr = func(e ast.Expr) {
            switch v := e.(type) {
            case *ast.CallExpr:
                if to, ok := index[v.Name]; ok {
                    adj[from] = append(adj[from], to)
                    radj[to] = append(radj[to], from)
                }
                for _, a := range v.Args { walkExpr(a) }
            case *ast.UnaryExpr:
                walkExpr(v.X)
            case *ast.BinaryExpr:
                walkExpr(v.X); walkExpr(v.Y)
            case *ast.SliceLit:
                for _, el := range v.Elems { walkExpr(el) }
            case *ast.SetLit:
                for _, el := range v.Elems { walkExpr(el) }
            case *ast.MapLit:
                for _, kv := range v.Elems { walkExpr(kv.Key); walkExpr(kv.Val) }
            }
        }
        for _, st := range fn.Body.Stmts {
            switch v := st.(type) {
            case *ast.ExprStmt:
                if v.X != nil { walkExpr(v.X) }
            case *ast.AssignStmt:
                if v.Value != nil { walkExpr(v.Value) }
            case *ast.VarDecl:
                if v.Init != nil { walkExpr(v.Init) }
            case *ast.ReturnStmt:
                for _, e := range v.Results { walkExpr(e) }
            case *ast.DeferStmt:
                if v.Call != nil { walkExpr(v.Call) }
            }
        }
    }
    // Kosaraju: order by finish time on adj, then dfs on radj
    vis := make([]bool, n)
    order := []int{}
    var dfs1 func(v int)
    dfs1 = func(v int) {
        vis[v] = true
        for _, w := range adj[v] { if !vis[w] { dfs1(w) } }
        order = append(order, v)
    }
    for v := 0; v < n; v++ { if !vis[v] { dfs1(v) } }
    comp := make([]int, n)
    for i := range comp { comp[i] = -1 }
    var dfs2 func(v, c int)
    dfs2 = func(v, c int) {
        comp[v] = c
        for _, w := range radj[v] { if comp[w] == -1 { dfs2(w, c) } }
    }
    c := 0
    for i := n-1; i >= 0; i-- {
        v := order[i]
        if comp[v] == -1 { dfs2(v, c); c++ }
    }
    // collect sets
    sets := make([]map[string]bool, c)
    for i := 0; i < c; i++ { sets[i] = map[string]bool{} }
    for i, name := range names { sets[comp[i]][name] = true }
    for i, name := range names { scc[name] = sets[comp[i]] }
    return scc
}

