package astjson

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    sch "github.com/sam-caldwell/ami/src/schemas"
)

func buildPipelineNode(n astpkg.PipelineDecl) sch.ASTNode {
    pd := sch.ASTNode{Kind: "PipelineDecl", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
    pd.Fields = map[string]interface{}{"name": n.Name, "connectors": n.Connectors}
    for _, st := range n.Steps {
        call := sch.ASTNode{Kind: "NodeCall", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
        fields := map[string]interface{}{"name": st.Name}
        if len(st.Args) > 0 {
            fields["args"] = st.Args
        }
        if st.Attrs != nil && len(st.Attrs) > 0 {
            fields["attrs"] = st.Attrs
        }
        if st.InlineWorker != nil {
            fields["inlineWorker"] = map[string]interface{}{"kind": "FuncLit"}
        }
        if len(st.Workers) > 0 {
            var ws []map[string]string
            for _, w := range st.Workers {
                ws = append(ws, map[string]string{"name": w.Name, "kind": w.Kind})
            }
            fields["workers"] = ws
        }
        call.Fields = fields
        pd.Children = append(pd.Children, call)
    }
    if len(n.ErrorSteps) > 0 {
        errNode := sch.ASTNode{Kind: "ErrorPipeline", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
        errNode.Fields = map[string]interface{}{"connectors": n.ErrorConnectors}
        for _, st := range n.ErrorSteps {
            call := sch.ASTNode{Kind: "NodeCall", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
            fields := map[string]interface{}{"name": st.Name}
            if len(st.Args) > 0 {
                fields["args"] = st.Args
            }
            if st.Attrs != nil && len(st.Attrs) > 0 {
                fields["attrs"] = st.Attrs
            }
            if st.InlineWorker != nil {
                fields["inlineWorker"] = map[string]interface{}{"kind": "FuncLit"}
            }
            if len(st.Workers) > 0 {
                var ws []map[string]string
                for _, w := range st.Workers {
                    ws = append(ws, map[string]string{"name": w.Name, "kind": w.Kind})
                }
                fields["workers"] = ws
            }
            call.Fields = fields
            errNode.Children = append(errNode.Children, call)
        }
        pd.Children = append(pd.Children, errNode)
    }
    return pd
}
