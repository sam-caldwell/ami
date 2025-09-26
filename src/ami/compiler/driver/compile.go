package driver

import (
    astjson "github.com/sam-caldwell/ami/src/ami/compiler/astjson"
    cg "github.com/sam-caldwell/ami/src/ami/compiler/codegen"
    cdiag "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    irpkg "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/sem"
)

// CompileWithDiagnostics parses each file, lowers to IR, generates debug schemas
// and ASM listings, and accumulates parser/semantic diagnostics.
func CompileWithDiagnostics(files []string, opts Options) (Result, []cdiag.Diagnostic, error) {
    res := Result{}
    var diags []cdiag.Diagnostic
    for _, f := range files {
        // Parse to internal AST
        p := parser.New(mustReadFile(f))
        fileAST := p.ParseFile()
        // collect any parse diagnostics and attach file path
        if errs := p.Errors(); len(errs) > 0 {
            for _, d := range errs {
                d.File = f
                diags = append(diags, d)
            }
        }
        // Optional: include semantic diagnostics
        if opts.SemDiags {
            if sres := sem.AnalyzeFile(fileAST); len(sres.Diagnostics) > 0 {
                for _, d := range sres.Diagnostics {
                    if d.File == "" {
                        d.File = f
                    }
                    diags = append(diags, d)
                }
            }
        }
        pkgName := fileAST.Package
        if pkgName == "" {
            pkgName = "main"
        }

        // Richer AST schema output
        astOut := astjson.ToSchemaAST(fileAST, f)
        res.AST = append(res.AST, astOut)

        // Lower to IR and convert to schema
        irMod := irpkg.FromASTFile(pkgName, fileAST.Version, f, fileAST)
        // Apply pragma-derived attributes and lower pipelines for worker/factory info
        irMod.ApplyDirectives(fileAST.Directives)
        if opts.EffectiveConcurrency > 0 {
            irMod.Concurrency = opts.EffectiveConcurrency
        }
        irMod.LowerPipelines(fileAST)
        irOut := irMod.ToSchema()
        res.IR = append(res.IR, irOut)
        // Pipelines debug IR
        pipes := irMod.ToPipelinesSchema()
        res.Pipelines = append(res.Pipelines, pipes)

        // Generate assembly text
        asmText := cg.GenerateASM(irMod)
        res.ASM = append(res.ASM, ASMUnit{Package: pkgName, Unit: f, Text: asmText})

        // Event metadata debug artifact (scaffold)
        res.EventMeta = append(res.EventMeta, buildEventMeta(pkgName, f))
    }
    return res, diags, nil
}

// Compile parses, lowers to IR, and generates assembly text deterministically.
func Compile(files []string, opts Options) (Result, error) {
    res, _, err := CompileWithDiagnostics(files, opts)
    return res, err
}

