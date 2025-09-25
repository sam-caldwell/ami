package root

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/sam-caldwell/ami/src/ami/compiler/driver"
	"github.com/sam-caldwell/ami/src/ami/workspace"
	"github.com/sam-caldwell/ami/src/internal/logger"
	sch "github.com/sam-caldwell/ami/src/schemas"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
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
			// Resolved sources and compiler driver scaffolds
			resolved := sch.SourcesV1{Schema: "sources.v1", Units: []sch.SourceUnit{}}
			var files []string
            if _, err := os.Stat("src/main.ami"); err == nil { files = append(files, "src/main.ami") }
            if len(files) > 0 {
				srcBytes, _ := os.ReadFile(files[0])
                resolved.Units = append(resolved.Units, sch.SourceUnit{Package: "main", File: files[0], Imports: []string{}, Source: string(srcBytes)})
				// Use compiler driver to create AST/IR
                res, _ := driver.Compile(files, driver.Options{})
                // AST per package/unit
                for _, a := range res.AST {
                    pkgDir := filepath.Join("build","debug","ast", a.Package)
                    _ = os.MkdirAll(pkgDir, 0755)
                    unit := filepath.Base(a.File)
                    b, _ := json.MarshalIndent(a, "", "  ")
                    _ = os.WriteFile(filepath.Join(pkgDir, unit+".ast.json"), b, 0644)
                }
				// IR per package/unit
                for _, ir := range res.IR {
                    pkgDir := filepath.Join("build","debug","ir", ir.Package)
                    _ = os.MkdirAll(pkgDir, 0755)
                    unit := filepath.Base(ir.File)
                    b, _ := json.MarshalIndent(ir, "", "  ")
                    _ = os.WriteFile(filepath.Join(pkgDir, unit+".ir.json"), b, 0644)
                }
				// ASM per package/unit + index
                asmFiles := []sch.ASMFile{}
                for _, f := range files {
                    pkgDir := filepath.Join("build","debug","asm", "main")
                    _ = os.MkdirAll(pkgDir, 0755)
                    unit := filepath.Base(f)
                    asmPath := filepath.Join(pkgDir, unit+".s")
                    _ = os.WriteFile(asmPath, []byte("; AMI-IR assembly scaffold\n"), 0644)
                    content := []byte("; AMI-IR assembly scaffold\n")
                    _ = os.WriteFile(asmPath, content, 0644)
                    size := int64(len(content))
                    sum := sha256.Sum256(content)
                    asmFiles = append(asmFiles, sch.ASMFile{Unit: f, Path: asmPath, Size: size, Sha256: hex.EncodeToString(sum[:])})
                }
                asmIdx := sch.ASMIndexV1{Schema: "asm.v1", Package: "main", Files: asmFiles}
                b, _ := json.MarshalIndent(asmIdx, "", "  ")
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
