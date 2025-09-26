package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    "strings"
)

// analyzeIOPermissions enforces that only ingress/egress nodes may perform I/O
// when step arguments indicate I/O usage via simple attributes.
// Detection rules (scaffold):
// - Any argument starting with "io=" (e.g., io=read, io=write)
// - Any argument containing "io.read(" or "io.write("
// These forms are only allowed on ingress/egress. Others emit E_IO_PERMISSION.
func analyzeIOPermissions(pd astpkg.PipelineDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    isIOArg := func(s string) bool {
        s = strings.TrimSpace(s)
        if s == "" {
            return false
        }
        if strings.HasPrefix(s, "io=") {
            return true
        }
        if strings.Contains(s, "io.read(") || strings.Contains(s, "io.write(") {
            return true
        }
        return false
    }
    check := func(name string, args []string) {
        n := strings.ToLower(name)
        allowed := (n == "ingress" || n == "egress")
        if allowed {
            return
        }
        for _, a := range args {
            if isIOArg(a) {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_IO_PERMISSION", Message: "I/O operations are only allowed in ingress/egress nodes"})
                break
            }
        }
    }
    for _, st := range pd.Steps {
        check(st.Name, st.Args)
    }
    for _, st := range pd.ErrorSteps {
        check(st.Name, st.Args)
    }
    return diags
}

