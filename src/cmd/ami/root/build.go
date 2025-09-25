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
    "sort"
    "runtime"
    "time"
    "strings"
)

var buildVerbose bool

func newBuildCmd() *cobra.Command {
    cmd := &cobra.Command{
	    Use:   "build",
	    Short: "Build the workspace",
        Example: `  # Build project
  ami build

  # Build with debug artifacts (AST/IR/ASM)
  ami build --verbose

  # Emit JSON diagnostics
  ami --json build`,
        Run: func(cmd *cobra.Command, args []string) {
        wsPath := "ami.workspace"
        ws, err := workspace.Load(wsPath)
        if err != nil {
            // Build-time enforcement of workspace schema constraints
            if flagJSON {
                d := sch.DiagV1{Schema: "diag.v1", Timestamp: sch.FormatTimestamp(nowUTC()), Level: "error", Code: "E_WS_SCHEMA", Message: fmt.Sprintf("workspace validation failed: %v", err), File: wsPath}
                if verr := d.Validate(); verr == nil {
                    enc := json.NewEncoder(os.Stdout)
                    _ = enc.Encode(d)
                } else {
                    // fallback: plain message to stdout if schema couldn't validate (shouldn't happen)
                    fmt.Fprintln(os.Stdout, "{\"schema\":\"diag.v1\",\"level\":\"error\",\"message\":\"workspace validation failed\"}")
                }
            } else {
                // Human mode: plain text to stderr and exit USER_ERROR
                fmt.Fprintln(os.Stderr, fmt.Sprintf("workspace error: %v", err))
            }
            os.Exit(ex.UserError)
        }
        _ = ws
        // If an existing ami.manifest is present, cross-check it against ami.sum
        if _, err := os.Stat("ami.manifest"); err == nil {
            if _, serr := os.Stat("ami.sum"); serr == nil {
                if existing, lerr := manifest.Load("ami.manifest"); lerr == nil {
                    if cerr := existing.CrossCheckWithSumFile("ami.sum"); cerr != nil {
                        logger.Error("integrity: existing manifest vs ami.sum mismatch", map[string]interface{}{"error": cerr.Error()})
                        os.Stderr.WriteString("integrity violation: existing manifest and ami.sum do not match\n")
                        os.Exit(ex.IntegrityViolationError)
                    }
                }
            }
        }
        plan := sch.BuildPlanV1{
            Schema:    "buildplan.v1",
            Timestamp: sch.FormatTimestamp(nowUTC()),
            Workspace: ".",
            Toolchain: sch.ToolchainV1{AmiVersion: version, GoVersion: runtime.Version()},
            Targets:   []sch.BuildTarget{},
        }
        // Discover source files and populate plan with deterministic ordering
        allFiles, _ := filepath.Glob(filepath.Join("src", "*.ami"))
        sort.Strings(allFiles)
        for _, f := range allFiles {
            pkg := filepath.Base(filepath.Dir(f))
            if pkg == "." || pkg == "src" { pkg = "main" }
            outAsm := filepath.Join("build","debug","asm", pkg, filepath.Base(f)+".s")
            outAST := filepath.Join("build","debug","ast", pkg, filepath.Base(f)+".ast.json")
            outIR := filepath.Join("build","debug","ir", pkg, filepath.Base(f)+".ir.json")
            plan.Targets = append(plan.Targets, sch.BuildTarget{
                Package: pkg,
                Unit:    f,
                Inputs:  []string{f},
                Outputs: []string{outAST, outIR, outAsm},
                Steps:   []string{"parse","typecheck","ir","codegen"},
            })
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
                    out := filepath.Join(pkgDir, unit+".ast.json")
                    _ = os.WriteFile(out, b, 0644)
                    logger.Info("build.debug.artifact", map[string]interface{}{"kind":"ast", "path": out})
                }
                // IR per package/unit
                for _, ir := range res.IR {
                    pkgDir := filepath.Join("build","debug","ir", ir.Package)
                    _ = os.MkdirAll(pkgDir, 0755)
                    unit := filepath.Base(ir.File)
                    b, _ := json.MarshalIndent(ir, "", "  ")
                    out := filepath.Join(pkgDir, unit+".ir.json")
                    _ = os.WriteFile(out, b, 0644)
                    logger.Info("build.debug.artifact", map[string]interface{}{"kind":"ir", "path": out})
                }
                // Pipelines debug IR per package/unit
                for _, p := range res.Pipelines {
                    pkgDir := filepath.Join("build","debug","ir", p.Package)
                    _ = os.MkdirAll(pkgDir, 0755)
                    unit := filepath.Base(p.File)
                    b, _ := json.MarshalIndent(p, "", "  ")
                    out := filepath.Join(pkgDir, unit+".pipelines.json")
                    _ = os.WriteFile(out, b, 0644)
                    logger.Info("build.debug.artifact", map[string]interface{}{"kind":"pipelines", "path": out})
                }
                // Event metadata per package/unit
                for _, em := range res.EventMeta {
                    pkgDir := filepath.Join("build","debug","ir", em.Package)
                    _ = os.MkdirAll(pkgDir, 0755)
                    unit := filepath.Base(em.File)
                    b, _ := json.MarshalIndent(em, "", "  ")
                    out := filepath.Join(pkgDir, unit+".eventmeta.json")
                    _ = os.WriteFile(out, b, 0644)
                    logger.Info("build.debug.artifact", map[string]interface{}{"kind":"eventmeta", "path": out})
                }
                // ASM per package/unit + per-package index (use compiler codegen output)
                asmByPkg := map[string][]sch.ASMFile{}
                edgesByPkg := map[string][]sch.EdgeInitV1{}
                for _, unit := range res.ASM {
                    pkg := unit.Package
                    pkgDir := filepath.Join("build","debug","asm", pkg)
                    _ = os.MkdirAll(pkgDir, 0755)
                    base := filepath.Base(unit.Unit)
                    asmPath := filepath.Join(pkgDir, base+".s")
                    content := []byte(unit.Text)
                    _ = os.WriteFile(asmPath, content, 0644)
                    logger.Info("build.debug.artifact", map[string]interface{}{"kind":"asm", "path": asmPath})
                    // In human verbose mode, echo discovered edge_init stubs to stdout for visibility
                    if !flagJSON {
                        lines := strings.Split(unit.Text, "\n")
                        for _, ln := range lines {
                            l := strings.TrimSpace(ln)
                            if strings.HasPrefix(l, "edge_init ") {
                                logger.Info("edge-init "+l, nil)
                            }
                        }
                    }
                    size := int64(len(content))
                    sum := sha256.Sum256(content)
                    asmByPkg[pkg] = append(asmByPkg[pkg], sch.ASMFile{Unit: unit.Unit, Path: asmPath, Size: size, Sha256: hex.EncodeToString(sum[:])})
                }
                // Collect edge summary per package from pipelines debug IR
                for _, p := range res.Pipelines {
                    // capture edges for this unit
                    for _, pipe := range p.Pipelines {
                        // Steps
                        for i, st := range pipe.Steps {
                            if st.InEdge == nil { continue }
                            label := fmt.Sprintf("%s.step%d.in", pipe.Name, i)
                            ei := sch.EdgeInitV1{Unit: p.File, Pipeline: pipe.Name, Segment: "normal", Step: i, Node: st.Node, Label: label, Kind: st.InEdge.Kind, MinCapacity: st.InEdge.MinCapacity, MaxCapacity: st.InEdge.MaxCapacity, Backpressure: st.InEdge.Backpressure, Type: st.InEdge.Type, UpstreamName: st.InEdge.UpstreamName,
                                Bounded: st.InEdge.Bounded, Delivery: st.InEdge.Delivery}
                            edgesByPkg[p.Package] = append(edgesByPkg[p.Package], ei)
                        }
                        // Error steps
                        for i, st := range pipe.ErrorSteps {
                            if st.InEdge == nil { continue }
                            label := fmt.Sprintf("%s.step%d.in", pipe.Name, i)
                            ei := sch.EdgeInitV1{Unit: p.File, Pipeline: pipe.Name, Segment: "error", Step: i, Node: st.Node, Label: label, Kind: st.InEdge.Kind, MinCapacity: st.InEdge.MinCapacity, MaxCapacity: st.InEdge.MaxCapacity, Backpressure: st.InEdge.Backpressure, Type: st.InEdge.Type, UpstreamName: st.InEdge.UpstreamName,
                                Bounded: st.InEdge.Bounded, Delivery: st.InEdge.Delivery}
                            edgesByPkg[p.Package] = append(edgesByPkg[p.Package], ei)
                        }
                    }
                }
                // write per-package indexes with deterministic package ordering
                var pkgs []string
                for pkg := range asmByPkg { pkgs = append(pkgs, pkg) }
                if len(pkgs) > 1 {
                    // simple selection sort to avoid importing sort; but importing sort is fine
                }
                // use sort.Strings for clarity
                // (import already present? we'll add)
                sort.Strings(pkgs)
                for _, pkg := range pkgs {
                    asmIdx := sch.ASMIndexV1{Schema: "asm.v1", Package: pkg, Files: asmByPkg[pkg]}
                    if items, ok := edgesByPkg[pkg]; ok && len(items) > 0 {
                        asmIdx.Edges = items
                    }
                    b, _ := json.MarshalIndent(asmIdx, "", "  ")
                    out := filepath.Join("build","debug","asm", pkg, "index.json")
                    _ = os.WriteFile(out, b, 0644)
                    logger.Info("build.debug.artifact", map[string]interface{}{"kind":"asmIndex", "path": out})
                    // Also write per-package edge summary if available
                    if items, ok := edgesByPkg[pkg]; ok && len(items) > 0 {
                        ed := sch.EdgesV1{Schema: "edges.v1", Timestamp: sch.FormatTimestamp(nowUTC()), Package: pkg, Items: items}
                        if err := ed.Validate(); err == nil {
                            eb, _ := json.MarshalIndent(ed, "", "  ")
                            epath := filepath.Join("build","debug","asm", pkg, "edges.json")
                            _ = os.WriteFile(epath, eb, 0644)
                            logger.Info("build.debug.artifact", map[string]interface{}{"kind":"edges", "path": epath})
                        }
                    }
                }
			}
				b, _ := json.MarshalIndent(resolved, "", "  ")
				out := filepath.Join("build", "debug", "source", "resolved.json")
				_ = os.WriteFile(out, b, 0644)
                logger.Info("build.debug.artifact", map[string]interface{}{"kind":"resolved", "path": out})

                // Write build plan file and log location (human)
                planPath := filepath.Join("build","debug","buildplan.json")
                if pb, err := json.MarshalIndent(plan, "", "  "); err == nil {
                    _ = os.WriteFile(planPath, pb, 0644)
                    logger.Info(fmt.Sprintf("build plan written: %s", planPath), map[string]interface{}{"targets": len(plan.Targets)})
                }
		}
        // Emit build plan as a JSON record in --json mode
        if flagJSON {
            if err := plan.Validate(); err == nil {
                _ = json.NewEncoder(os.Stdout).Encode(plan)
            }
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
                    // Emit JSON diagnostic summary in JSON mode
                    if flagJSON {
                        d := sch.DiagV1{Schema: "diag.v1", Timestamp: sch.FormatTimestamp(nowUTC()), Level: "error", Code: "E_INTEGRITY", Message: "integrity violation: ami.sum does not match cache"}
                        if verr := d.Validate(); verr == nil { _ = json.NewEncoder(os.Stdout).Encode(d) }
                    }
                    // Fail build with integrity violation exit code
                    os.Stderr.WriteString("integrity violation: ami.sum does not match cache\n")
                    os.Exit(ex.IntegrityViolationError)
                }
            }
        }

        // Write ami.manifest with artifacts/toolchain and cross-check ami.sum
        artifacts := []manifest.Artifact{}
        for _, path := range []struct{p,kind string}{{"build/debug/source/resolved.json","resolved"},{"build/debug/ast/main/main.ami.ast.json","ast"},{"build/debug/ir/main/main.ami.ir.json","ir"},{"build/debug/ir/main/main.ami.pipelines.json","pipelines"},{"build/debug/ir/main/main.ami.eventmeta.json","eventmeta"},{"build/debug/asm/main/main.ami.s","asm"},{"build/debug/asm/index.json","asmIndex"}} {
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
        // Cross-check manifest packages against ami.sum if present (completeness)
        if _, err := os.Stat("ami.sum"); err == nil {
            if err := man.CrossCheckWithSumFile("ami.sum"); err != nil {
                logger.Error("integrity: manifest vs ami.sum mismatch", map[string]interface{}{"error": err.Error()})
                if flagJSON {
                    d := sch.DiagV1{Schema: "diag.v1", Timestamp: sch.FormatTimestamp(nowUTC()), Level: "error", Code: "E_INTEGRITY_MANIFEST", Message: "integrity violation: manifest and ami.sum do not match"}
                    if verr := d.Validate(); verr == nil { _ = json.NewEncoder(os.Stdout).Encode(d) }
                }
                os.Stderr.WriteString("integrity violation: manifest and ami.sum do not match\n")
                os.Exit(ex.IntegrityViolationError)
            }
        }
        if err := manifest.Save("ami.manifest", &man); err != nil {
            logger.Error(fmt.Sprintf("failed to write ami.manifest: %v", err), nil)
            return
        }
        logger.Info("build completed (scaffold)", map[string]interface{}{"targets": len(plan.Targets)})
	    },
    }
    cmd.Flags().BoolVar(&buildVerbose, "verbose", false, "emit debug artifacts")
    return cmd
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

func nowUTC() time.Time { return time.Now().UTC() }
