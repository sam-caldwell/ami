package root

import (
	"crypto/sha256"
	"io"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/sam-caldwell/ami/src/ami/compiler/driver"
	"github.com/sam-caldwell/ami/src/ami/workspace"
    "github.com/sam-caldwell/ami/src/ami/manifest"
    ammod "github.com/sam-caldwell/ami/src/ami/mod"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/internal/logger"
	ex "github.com/sam-caldwell/ami/src/internal/exit"
	sch "github.com/sam-caldwell/ami/src/schemas"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
    "runtime"
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
            Toolchain: sch.ToolchainV1{AmiVersion: "v0.0.0-dev", GoVersion: "1.25.1"},
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
                imports := parser.ExtractImports(string(srcBytes))
                resolved.Units = append(resolved.Units, sch.SourceUnit{Package: "main", File: files[0], Imports: imports, Source: string(srcBytes)})
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
        // Validate cache integrity against ami.sum (fail build on mismatch)
        if sum, err := ammod.LoadSumForCLI("ami.sum"); err == nil {
            cacheDir, cerr := ammod.CacheDir()
            if cerr == nil {
                ok := true
                for pkg, vers := range sum.Packages {
                    base := filepath.Base(pkg)
                    for ver, digest := range vers {
                        entry := filepath.Join(cacheDir, base+"@"+ver)
                        if fi, e := os.Stat(entry); e != nil || !fi.IsDir() {
                            ok = false
                            logger.Error("integrity: cache entry missing", map[string]interface{}{"pkg": pkg, "version": ver, "path": entry})
                            continue
                        }
                        d2, e := ammod.CommitDigestForCLI(entry, ver)
                        if e != nil {
                            ok = false
                            logger.Error("integrity: digest compute failed", map[string]interface{}{"pkg": pkg, "version": ver, "error": e.Error()})
                            continue
                        }
                        if d2 != digest {
                            ok = false
                            logger.Error("integrity: digest mismatch", map[string]interface{}{"pkg": pkg, "version": ver})
                        }
                    }
                }
                if !ok {
                    // Fail build with integrity violation exit code
                    os.Stderr.WriteString("integrity violation: ami.sum does not match cache\n")
                    os.Exit(ex.IntegrityViolationError)
                }
            }
        }

        // Write ami.manifest with artifacts/toolchain and cross-check ami.sum
        artifacts := []manifest.Artifact{}
        for _, path := range []struct{p,kind string}{{"build/debug/source/resolved.json","resolved"},{"build/debug/ast/main/main.ami.ast.json","ast"},{"build/debug/ir/main/main.ami.ir.json","ir"},{"build/debug/asm/main/main.ami.s","asm"},{"build/debug/asm/index.json","asmIndex"}} {
            if fi, err := os.Stat(path.p); err==nil && !fi.IsDir() {
                sha, size, _ := fileSHA256(path.p)
                artifacts = append(artifacts, manifest.Artifact{Path: path.p, Kind: path.kind, Size: size, Sha256: sha})
            }
        }
        wd, _ := os.Getwd()
        projName := filepath.Base(wd)
        projVersion := "0.0.0"
        amiVer := version
        goVer := runtime.Version()
        pkgs := []manifest.Package{}
        if sum, err := ammod.LoadSumForCLI("ami.sum"); err == nil {
            for name, vers := range sum.Packages {
                for ver, digest := range vers {
                    cache, _ := ammod.CacheDir()
                    base := filepath.Base(name)
                    src := filepath.Join(cache, base+"@"+ver)
                    pkgs = append(pkgs, manifest.Package{Name: name, Version: ver, Digest: digest, Source: src})
                }
            }
        }
        man := manifest.Manifest{Schema: "ami.manifest/v1", Project: manifest.Project{Name: projName, Version: projVersion}, Packages: pkgs, Artifacts: artifacts, Toolchain: manifest.Toolchain{AmiVersion: amiVer, GoVersion: goVer}}
        if err := manifest.Save("ami.manifest", &man); err != nil {
            logger.Error(fmt.Sprintf("failed to write ami.manifest: %v", err), nil)
            return
        }
        logger.Info("build completed (scaffold)", map[string]interface{}{"targets": len(plan.Targets)})
	},
}

func init() {
	cmdBuild.Flags().BoolVar(&buildVerbose, "verbose", false, "emit debug artifacts")
}


func fileSHA256(path string) (string, int64, error) {
    f, err := os.Open(path)
    if err != nil { return "", 0, err }
    defer f.Close()
    h := sha256.New()
    n, err := io.Copy(h, f)
    if err != nil { return "", 0, err }
    return hex.EncodeToString(h.Sum(nil)), n, nil
}
