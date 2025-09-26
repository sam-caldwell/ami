package sem

import (
    "fmt"
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    "strings"
)

func analyzePipeline(pd astpkg.PipelineDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    // Basic pipeline shape checks
    if len(pd.Steps) == 0 {
        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_PIPELINE_EMPTY", Message: fmt.Sprintf("pipeline %q has no steps", pd.Name)})
        return diags
    }
    allowed := map[string]bool{"ingress": true, "transform": true, "fanout": true, "collect": true, "egress": true}
    ingressCount := 0
    egressCount := 0
    // validate step names and counts
    for _, s := range pd.Steps {
        n := strings.ToLower(s.Name)
        if !allowed[n] {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_UNKNOWN_NODE", Message: "invalid pipeline step: " + s.Name})
        }
        if n == "ingress" {
            ingressCount++
        }
        if n == "egress" {
            egressCount++
        }
    }
    if ingressCount == 0 || strings.ToLower(pd.Steps[0].Name) != "ingress" {
        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_PIPELINE_START_INGRESS", Message: "pipeline must start with ingress"})
    }
    if egressCount == 0 || strings.ToLower(pd.Steps[len(pd.Steps)-1].Name) != "egress" {
        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_PIPELINE_END_EGRESS", Message: "pipeline must end with egress"})
    }
    // error path checks
    for _, s := range pd.ErrorSteps {
        n := strings.ToLower(s.Name)
        if !allowed[n] {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_UNKNOWN_NODE", Message: "invalid pipeline step: " + s.Name})
        }
    }
    // Step placement rules
    for i, st := range pd.Steps {
        n := strings.ToLower(st.Name)
        if n == "ingress" && i != 0 {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_INGRESS_POSITION", Message: "ingress must be the first step"})
        }
        if n == "egress" && i != len(pd.Steps)-1 {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EGRESS_POSITION", Message: "egress must be the last step"})
        }
    }
    for i, st := range pd.ErrorSteps {
        n := strings.ToLower(st.Name)
        if n == "ingress" && i != 0 {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_INGRESS_POSITION", Message: "ingress must be the first step"})
        }
        if n == "egress" && i != len(pd.ErrorSteps)-1 {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ERRPIPE_END_EGRESS", Message: "error path must end with egress"})
        }
    }
    // Required workers presence for transform/fanout/collect
    checkWorkers := func(step astpkg.NodeCall) {
        n := strings.ToLower(step.Name)
        switch n {
        case "transform", "fanout", "collect":
            if len(step.Workers) == 0 {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_PIPELINE_WORKERS", Message: n + " step requires at least one worker"})
            }
        }
    }
    for _, s := range pd.Steps {
        checkWorkers(s)
    }
    for _, s := range pd.ErrorSteps {
        checkWorkers(s)
    }
    // egress cannot have workers (sink) — skip strict enforcement at this stage
    // Duplicate ingress/egress detection
    if ingressCount > 1 {
        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_DUP_INGRESS", Message: "pipeline has multiple ingress steps"})
    }
    if egressCount > 1 {
        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_DUP_EGRESS", Message: "pipeline has multiple egress steps"})
    }
    // Error path rules: cannot start with ingress; must end with egress
    if len(pd.ErrorSteps) > 0 {
        if strings.ToLower(pd.ErrorSteps[0].Name) == "ingress" {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ERRPIPE_START_INVALID", Message: "error path cannot start with ingress"})
        }
        if strings.ToLower(pd.ErrorSteps[len(pd.ErrorSteps)-1].Name) != "egress" {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ERRPIPE_END_EGRESS", Message: "error path must end with egress"})
        }
    }
    return diags
}
