package root

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "github.com/spf13/cobra"
    "github.com/sam-caldwell/ami/src/ami/workspace"
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
            resolved := sch.SourcesV1{Schema: "sources.v1", Units: []sch.SourceUnit{}}
            b, _ := json.MarshalIndent(resolved, "", "  ")
            _ = os.WriteFile(filepath.Join("build", "debug", "source", "resolved.json"), b, 0644)
        }
        logger.Info("build completed (scaffold)", map[string]interface{}{"targets": len(plan.Targets)})
    },
}

func init() {
    cmdBuild.Flags().BoolVar(&buildVerbose, "verbose", false, "emit debug artifacts")
}

