package root

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "github.com/spf13/cobra"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    "github.com/sam-caldwell/ami/src/ami/compiler/driver"
    "github.com/sam-caldwell/ami/src/internal/logger"
    sch "github.com/sam-caldwell/ami/src/schemas"
)

var buildVerbose bool

var cmdBuild = &cobra.Command{
    Use:   "build",
    Short: "Build the workspace",
    Run: func(cmd *cobra.Command, args []string) {
        wsPath := "ami.workspace"
        ws, err := workspace.Load(wsPath)
        if err != nil {
            logger.Error(fmt.Sprintf("failed to load %s: %v", wsPath, err), nil)
            return
        }
        _ = ws
        plan := sch.BuildPlanV1{
            Schema:    "buildplan.v1",
            Workspace: ".",
            Toolchain: sch.ToolchainV1{AmiVersion: "v0.0.0-dev", GoVersion: "1.25"},
            Targets:   []sch.BuildTarget{},
        }
        if buildVerbose {
            _ = os.MkdirAll("build/debug/source", 0755)
            _ = os.MkdirAll("build/debug/ast", 0755)
            _ = os.MkdirAll("build/debug/ir", 0755)
            _ = os.MkdirAll("build/debug/asm", 0755)
            // Resolved sources (scaffold for src/main.ami)
            resolved := sch.SourcesV1{Schema: "sources.v1", Units: []sch.SourceUnit{}}
            if _, err := os.Stat("src/main.ami"); err == nil {
                resolved.Units = append(resolved.Units, sch.SourceUnit{Package: "main", File: "src/main.ami", Imports: []string{}, Source: ""})
                // AST scaffold
                ast := sch.ASTV1{Schema: "ast.v1", Package: "main", File: "src/main.ami", Root: sch.ASTNode{Kind: "File", Pos: sch.Position{Line:1,Column:1,Offset:0}}}
                b, _ := json.MarshalIndent(ast, "", "  ")
                _ = os.WriteFile(filepath.Join("build","debug","ast","main.ast.json"), b, 0644)
                // IR scaffold
                ir := sch.IRV1{Schema: "ir.v1", Package: "main", File: "src/main.ami", Functions: []sch.IRFunction{{Name:"main", Blocks: []sch.IRBlock{{Label:"entry"}}}}}
                b, _ = json.MarshalIndent(ir, "", "  ")
                _ = os.WriteFile(filepath.Join("build","debug","ir","main.ir.json"), b, 0644)
                // ASM scaffold
                _ = os.WriteFile(filepath.Join("build","debug","asm","main.s"), []byte("; AMI-IR assembly scaffold\n"), 0644)
                asmIdx := sch.ASMIndexV1{Schema: "asm.v1", Package: "main", Files: []sch.ASMFile{{Unit:"src/main.ami", Path:"build/debug/asm/main.s", Size:0, Sha256:""}}}
                b, _ = json.MarshalIndent(asmIdx, "", "  ")
                _ = os.WriteFile(filepath.Join("build","debug","asm","index.json"), b, 0644)
            }
            b, _ := json.MarshalIndent(resolved, "", "  ")
            _ = os.WriteFile(filepath.Join("build", "debug", "source", "resolved.json"), b, 0644)
        }
        logger.Info("build completed (scaffold)", map[string]interface{}{"targets": len(plan.Targets)})
    },
}

func init() {
    cmdBuild.Flags().BoolVar(&buildVerbose, "verbose", false, "emit debug artifacts")
}
