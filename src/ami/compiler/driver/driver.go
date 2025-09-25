package driver

import (
    "os"
    cg "github.com/sam-caldwell/ami/src/ami/compiler/codegen"
    irpkg "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    astjson "github.com/sam-caldwell/ami/src/ami/compiler/astjson"
    sch "github.com/sam-caldwell/ami/src/schemas"
)

// Options placeholder for future flags
type Options struct{}

// ASMUnit contains generated assembly text for a compilation unit.
type ASMUnit struct {
    Package string
    Unit    string // file path
    Text    string
}

// Result holds compiler outputs for scaffolding
type Result struct {
    AST []sch.ASTV1
    IR  []sch.IRV1
    Pipelines []sch.PipelinesV1
    EventMeta []sch.EventMetaV1
    ASM []ASMUnit // assembly text per unit
}

// Compile parses, lowers to IR, and generates assembly text deterministically.
func Compile(files []string, opts Options) (Result, error) {
    res := Result{}
    for _, f := range files {
        // Parse to internal AST
        p := parser.New(mustReadFile(f))
        fileAST := p.ParseFile()
        pkgName := fileAST.Package
        if pkgName == "" { pkgName = "main" }

        // Richer AST schema output
        astOut := astjson.ToSchemaAST(fileAST, f)
        res.AST = append(res.AST, astOut)

        // Lower to IR and convert to schema
        irMod := irpkg.FromASTFile(pkgName, f, fileAST)
        // Apply pragma-derived attributes and lower pipelines for worker/factory info
        irMod.ApplyDirectives(fileAST.Directives)
        irMod.LowerPipelines(fileAST)
        irOut := irMod.ToSchema()
        res.IR = append(res.IR, irOut)
        // Pipelines debug IR
        pipes := irMod.ToPipelinesSchema()
        res.Pipelines = append(res.Pipelines, pipes)

        // Generate assembly text
        asmText := cg.GenerateASM(irMod)
        res.ASM = append(res.ASM, ASMUnit{Package: pkgName, Unit: f, Text: asmText})

        // Event metadata debug artifact
        em := sch.EventMetaV1{Schema: "eventmeta.v1", Package: pkgName, File: f, ImmutablePayload: true,
            Fields: []sch.EventMetaFieldV1{
                {Name: "id", Type: "string", Note: "unique event identifier"},
                {Name: "timestamp", Type: "iso8601", Note: "creation time in UTC"},
                {Name: "attempt", Type: "int", Note: "delivery/retry attempt count"},
            },
            Trace: &sch.TraceContextV1{
                Traceparent: sch.EventMetaFieldV1{Name: "traceparent", Type: "string", Note: "W3C traceparent header (version-traceid-spanid-flags)"},
                Tracestate:  sch.EventMetaFieldV1{Name: "tracestate", Type: "string", Note: "W3C tracestate header (vendor extensions)"},
            },
        }
        res.EventMeta = append(res.EventMeta, em)
    }
    return res, nil
}

// mustReadFile returns source or empty string on error; build path already validated by caller.
func mustReadFile(path string) string {
    // Keep this local to avoid introducing CLI deps or fs mocks; callers supply existing files.
    // If read fails, return empty source to keep determinism in scaffolding.
    b, err := osReadFile(path)
    if err != nil { return "" }
    return string(b)
}

// indirection for testing/mocking if needed in future
var osReadFile = func(path string) ([]byte, error) { return os.ReadFile(path) }
