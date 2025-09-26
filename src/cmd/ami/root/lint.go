package root

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
	"github.com/sam-caldwell/ami/src/ami/compiler/diag"
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/sem"
	srcset "github.com/sam-caldwell/ami/src/ami/compiler/source"
	tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
	"github.com/sam-caldwell/ami/src/ami/workspace"
	"github.com/sam-caldwell/ami/src/internal/logger"
	sch "github.com/sam-caldwell/ami/src/schemas"
)

// types and flags moved to lint_types.go and lint_flags.go

func newLintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint AMI sources in the workspace",
		Example: `  ami lint
  ami --json lint`,
		Run: func(cmd *cobra.Command, args []string) {
			runLint()
		},
	}
	// Flags for 5.1 extensions
	cmd.Flags().BoolVar(&lintStrict, "strict", false, "treat warnings as errors (exit non-zero in strict mode)")
	cmd.Flags().StringVar(&lintRules, "rules", "", "only include rules matching pattern(s), comma-separated (case-insensitive substring match)")
	cmd.Flags().IntVar(&lintMaxWarn, "max-warn", 0, "maximum warnings to emit (0 = unlimited)")
	return cmd
}

// runLint discovers AMI source entrypoints using the workspace, orders
// them so that imported workspace packages are linted before the main
// package, and emits a sources.v1 summary (JSON when --json).
func runLint() {
	// Load workspace; on failure, emit diagnostic consistent with build.
	const wsPath = "ami.workspace"
	ws, err := workspace.Load(wsPath)
	if err != nil {
		if flagJSON {
			d := sch.DiagV1{Schema: "diag.v1", Timestamp: sch.FormatTimestamp(nowUTC()), Level: "error", Code: "E_WS_SCHEMA", Message: fmt.Sprintf("workspace validation failed: %v", err), File: wsPath}
			if verr := d.Validate(); verr == nil {
				_ = json.NewEncoder(os.Stdout).Encode(d)
			}
		} else {
			fmt.Fprintln(os.Stderr, fmt.Sprintf("workspace error: %v", err))
		}
		return
	}

	// Derive linter config and package -> root map from workspace
	lcfg := parseLinterConfig(ws)
	pkgRoots := parseWorkspacePackages(ws)
	// Workspace package rules (names, versions)
	diags := lintWorkspacePackages(ws)
	// Discover entry units ('.ami' files) per package
	// For main, prefer src/main.ami when present.
	// Lint order: imports (workspace-local) first, then main.
	order := lintOrder(pkgRoots)

	// Build sources summary in the discovered order
	sources := sch.SourcesV1{Schema: "sources.v1", Timestamp: sch.FormatTimestamp(nowUTC())}
	// Deterministic file collection per package (sorted)
	var ulist []lintUnitRec
	for _, pkg := range order {
		root := pkgRoots[pkg]
		files, _ := filepath.Glob(filepath.Join(root, "*.ami"))
		sort.Strings(files)
		for _, f := range files {
			b, _ := os.ReadFile(f)
			src := string(b)
            imports := parser.ExtractImports(src)
            // also capture detailed imports (alias/constraint) for sources.v1
            var importsDetailed []sch.SourceImport
            for _, it := range parser.ExtractImportItems(src) {
                importsDetailed = append(importsDetailed, sch.SourceImport{Path: it.Path, Alias: it.Alias, Constraint: it.Constraint})
            }
            // keep AST for lint rules
            p := parser.New(src)
            ast := p.ParseFile()
            ulist = append(ulist, lintUnitRec{pkg: pkg, file: f, src: src, imports: imports, ast: ast})
            sources.Units = append(sources.Units, sch.SourceUnit{Package: pkg, File: f, Imports: imports, ImportsDetailed: importsDetailed, Source: src})
		}
	}

	// Lint diagnostics across units
	for _, u := range ulist {
		diags = append(diags, lintUnit(u.pkg, u.file, u.src, u.ast, lcfg)...)
	}
	// Cross-unit hygiene within a package: duplicate function declarations across files
	diags = append(diags, lintDuplicateFunctionsAcrossUnits(ulist)...)
	// Cross-package constraint checks
	diags = append(diags, lintCrossPackageConstraints(ws, pkgRoots, ulist)...)

	// Output
	// Effective strictness considers workspace preset and CLI flag
	effectiveStrict := lintStrict || lcfg.strict
	if flagJSON {
		// Emit sources summary first for deterministic tooling
		if err := sources.Validate(); err == nil {
			_ = json.NewEncoder(os.Stdout).Encode(sources)
		}
		// Emit lint diagnostics as JSON lines
		errs := 0
		warns := 0
		warnEmitted := 0
		for _, d := range diags {
			if !ruleSelected(d.Code) {
				continue
			}
			if d.Level == diag.Warn && lintMaxWarn > 0 && warnEmitted >= lintMaxWarn {
				continue
			}
			level := d.Level
			if effectiveStrict && level == diag.Warn {
				level = diag.Error
			}
			if level == diag.Error {
				errs++
			} else if level == diag.Warn {
				warns++
				warnEmitted++
			}
			sd := d.ToSchema()
			sd.Level = string(level)
			if sd.Package == "" {
				sd.Package = uPackageFromPath(d.File, pkgRoots)
			}
			// Add LINT_* alias in data for forward-compat namespace
			if sd.Data == nil {
				sd.Data = map[string]any{}
			}
			sd.Data["lint_code"] = lintAlias(sd.Code)
			_ = json.NewEncoder(os.Stdout).Encode(sd)
		}
		// Summary record
		summary := sch.DiagV1{Schema: "diag.v1", Timestamp: sch.FormatTimestamp(nowUTC()), Level: "info", Code: "LINT_SUMMARY", Message: fmt.Sprintf("%d warnings, %d errors", warns, errs)}
		_ = json.NewEncoder(os.Stdout).Encode(summary)
		return
	}
	// Human mode: concise, ordered list with filenames inline
	logger.Info(fmt.Sprintf("lint: discovered %d units", len(sources.Units)), nil)
	for _, u := range sources.Units {
		logger.Info(fmt.Sprintf("lint: unit %s (%s)", u.File, u.Package), nil)
	}
	// Human diagnostics
	errs := 0
	warns := 0
	warnEmitted := 0
	for _, d := range diags {
		if !ruleSelected(d.Code) {
			continue
		}
		if d.Level == diag.Warn && lintMaxWarn > 0 && warnEmitted >= lintMaxWarn {
			continue
		}
		msg := fmt.Sprintf("%s: %s", d.Code, d.Message)
		if d.File != "" {
			msg = d.File + ": " + msg
		}
		level := d.Level
		if effectiveStrict && level == diag.Warn {
			level = diag.Error
		}
		switch level {
		case diag.Error:
			logger.Error(msg, nil)
			errs++
		case diag.Warn:
			logger.Warn(msg, nil)
			warns++
			warnEmitted++
		default:
			logger.Info(msg, nil)
		}
	}
	logger.Info(fmt.Sprintf("lint: summary: %d warnings, %d errors", warns, errs), nil)
}

// lintDuplicateFunctionsAcrossUnits warns when the same function name is declared
// in multiple units within the same package.
func lintDuplicateFunctionsAcrossUnits(units []lintUnitRec) []diag.Diagnostic {
	var out []diag.Diagnostic
	// pkg -> name -> []unit indices
	idx := map[string]map[string][]int{}
	for i, u := range units {
		if idx[u.pkg] == nil {
			idx[u.pkg] = map[string][]int{}
		}
		for _, d := range u.ast.Decls {
			if fd, ok := d.(astpkg.FuncDecl); ok && fd.Name != "" {
				idx[u.pkg][fd.Name] = append(idx[u.pkg][fd.Name], i)
			}
		}
	}
	for pkg, names := range idx {
		for name, list := range names {
			if len(list) <= 1 {
				continue
			}
			// emit a warning for each duplicate beyond the first occurrence
			for _, i := range list[1:] {
				u := units[i]
				// find position via AST
				var pos *srcset.Position
				for _, d := range u.ast.Decls {
					if fd, ok := d.(astpkg.FuncDecl); ok && fd.Name == name {
						fs := srcset.NewFileSet()
						sf := fs.AddFileFromSource(u.file, u.src)
						p := sf.PositionFor(fd.Pos.Offset)
						pos = &p
						break
					}
				}
				out = append(out, diag.Diagnostic{Level: diag.Warn, Code: "W_DUP_FUNC_ACROSS_FILES", Message: "duplicate function declared across multiple units: " + name, Package: pkg, File: u.file, Pos: pos})
			}
		}
	}
	return out
}

// parseWorkspacePackages extracts a map of package name -> root directory
// from the loosely-typed workspace.Packages field.
/* moved to parse_workspace_packages.go
func parseWorkspacePackages(ws *workspace.Workspace) map[string]string {
	out := map[string]string{}
	for _, p := range ws.Packages {
		m, ok := p.(map[string]any)
		if !ok {
			continue
		}
		for name, v := range m {
			// value is expected to be a map with at least 'root'
			if vm, ok := v.(map[string]any); ok {
				if r, ok := vm["root"].(string); ok && strings.TrimSpace(r) != "" {
					out[name] = r
				}
			}
		}
	}
	// If no packages parsed, default to main: ./src (common scaffold)
	if len(out) == 0 {
		out["main"] = "./src"
	}
	return out
}
*/

// lintOrder returns packages in the order they should be linted:
// imported workspace-local packages first, then main. If a local
// package itself imports other local packages, those are ordered
// before it (DFS). Packages without resolvable imports are included
// once. Unknown imports (external) are ignored for ordering.
/* moved to parse_workspace_packages.go
func lintOrder(pkgRoots map[string]string) []string {
	// Helper to read a representative unit for a package to extract imports.
	importsFor := func(pkg string) []string {
		root, ok := pkgRoots[pkg]
		if !ok {
			return nil
		}
		// Prefer main.ami; otherwise first *.ami by name.
		mainPath := filepath.Join(root, "main.ami")
		var path string
		if fi, err := os.Stat(mainPath); err == nil && !fi.IsDir() {
			path = mainPath
		} else {
			list, _ := filepath.Glob(filepath.Join(root, "*.ami"))
			sort.Strings(list)
			if len(list) > 0 {
				path = list[0]
			}
		}
		if path == "" {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		return parser.ExtractImports(string(b))
	}

	// DFS
	visited := map[string]bool{}
	order := []string{}
	var visit func(string)
	visit = func(pkg string) {
		if visited[pkg] {
			return
		}
		visited[pkg] = true
		// For each import that is a workspace-local package, visit first
		for _, imp := range importsFor(pkg) {
			if _, ok := pkgRoots[imp]; ok {
				visit(imp)
			}
		}
		order = append(order, pkg)
	}
	// Always start from main if present
	if _, ok := pkgRoots["main"]; ok {
		visit("main")
	}
	// Include any remaining packages that weren't reachable from main
	// in stable order (sorted by name)
	var rest []string
	for k := range pkgRoots {
		if !visited[k] {
			rest = append(rest, k)
		}
	}
	sort.Strings(rest)
	for _, k := range rest {
		visit(k)
	}
	return order
}
*/

// lintWorkspacePackages enforces workspace-level package naming and version rules.
/* moved to lint_workspace_packages.go
func lintWorkspacePackages(ws *workspace.Workspace) []diag.Diagnostic {
	var out []diag.Diagnostic
	semverRe := regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)
	for _, p := range ws.Packages {
		m, ok := p.(map[string]any)
		if !ok {
			continue
		}
		for name, v := range m {
			// package key name should be a valid import path (Chapter 3.0 composition)
			if !parser.ValidateImportPath(name) {
				out = append(out, diag.Diagnostic{Level: diag.Error, Code: "E_WS_PKG_NAME", Message: fmt.Sprintf("invalid workspace package name: %q", name)})
			}
			vm, ok := v.(map[string]any)
			if !ok {
				continue
			}
			if ver, ok := vm["version"].(string); ok {
				if !semverRe.MatchString(ver) {
					out = append(out, diag.Diagnostic{Level: diag.Error, Code: "E_WS_PKG_VERSION", Message: fmt.Sprintf("package %q has invalid semantic version: %q", name, ver)})
				}
			}
		}
	}
	return out
}
*/

// lintUnit returns diagnostics for a single unit.
func lintUnit(pkgName, filePath, src string, f *astpkg.File, cfg lintConfig) []diag.Diagnostic {
	var out []diag.Diagnostic
	// Set up positions from source
	fs := srcset.NewFileSet()
	sf := fs.AddFileFromSource(filePath, src)
	toPos := func(offset int) *srcset.Position { p := sf.PositionFor(offset); return &p }
	// File-level suppression via pragmas: #pragma lint:disable RULE[,RULE2]
	// For this phase, apply file-wide disable; enable resets if seen later.
	disabled := map[string]bool{}
	for _, d := range f.Directives {
		name := strings.ToLower(strings.TrimSpace(d.Name))
		payload := strings.TrimSpace(d.Payload)
		if name == "lint:disable" {
			for _, part := range strings.Split(payload, ",") {
				code := strings.TrimSpace(part)
				if code == "" {
					continue
				}
				disabled[code] = true
			}
		}
		if name == "lint:enable" {
			for _, part := range strings.Split(payload, ",") {
				code := strings.TrimSpace(part)
				if code == "" {
					continue
				}
				delete(disabled, code)
			}
		}
	}
	// helper to test config-based suppression
	isSuppressed := func(code string, pkg, file string) bool {
		// package-based
		if m, ok := cfg.suppressPkg[pkg]; ok {
			if m[code] || m["*"] {
				return true
			}
		}
		// path-based globs (match on file path)
		for _, ent := range cfg.suppressPaths {
			if ent.rules[code] || ent.rules["*"] {
				if ok, _ := filepath.Match(ent.glob, file); ok {
					return true
				}
				// also try matching directory prefixes (dir/**)
				if strings.HasSuffix(ent.glob, "/**") {
					g := strings.TrimSuffix(ent.glob, "/**")
					if strings.HasPrefix(filepath.Clean(file), filepath.Clean(g)+string(os.PathSeparator)) {
						return true
					}
				}
			}
		}
		return false
	}

    apply := func(level diag.Level, code, message string, pkg, file string, pos *srcset.Position) {
		// file-local suppression
		if disabled[code] {
			return
		}
		// config-based suppression
		if isSuppressed(code, pkg, file) {
			return
		}
		// workspace severity overrides
		if sev, ok := cfg.severity[code]; ok {
			switch strings.ToLower(sev) {
			case "off":
				return
			case "error":
				level = diag.Error
			case "warn":
				level = diag.Warn
			case "info":
				level = diag.Info
			}
		}
		d := diag.Diagnostic{Level: level, Code: code, Message: message, Package: pkg, File: file}
		if pos != nil {
			d.Pos = pos
		}
        out = append(out, d)
    }
    // Ambient state: flag pointer State parameters in function signatures
    // We scan source since parser strips '*' before building TypeRef in params.
    if strings.Contains(src, "*State") {
        if idx := strings.Index(src, "*State"); idx >= 0 {
            apply(diag.Warn, "W_STATE_PARAM_POINTER", "do not use *State parameters; state is ambient. Use state.get/set/update/list", pkgName, filePath, toPos(idx))
        } else {
            apply(diag.Warn, "W_STATE_PARAM_POINTER", "do not use *State parameters; state is ambient. Use state.get/set/update/list", pkgName, filePath, nil)
        }
    }
	// Language reminders and compatibility hints
	if off := findPackageIdentOffset(src); off >= 0 {
		apply(diag.Warn, "W_LANG_NOT_GO", "AMI is not Go; semantics differ (no address-of '&', '*' is not dereference). Use '*' on assignment LHS and mutate(expr) for side effects.", pkgName, filePath, toPos(off))
	} else {
		apply(diag.Warn, "W_LANG_NOT_GO", "AMI is not Go; semantics differ (no address-of '&', '*' is not dereference). Use '*' on assignment LHS and mutate(expr) for side effects.", pkgName, filePath, nil)
	}
	// Heuristic detection of common Go-specific syntax
	var goHits []string
	patterns := []string{" var ", "\nvar ", ":=", " chan ", "\nchan ", " go ", " range ", "\ninterface ", " map[", " make(", " new("}
	for _, ptn := range patterns {
		if strings.Contains(src, ptn) {
			goHits = append(goHits, strings.TrimSpace(ptn))
		}
	}
	if len(goHits) > 0 {
		msg := "Go syntax detected (e.g., " + strings.Join(goHits, ", ") + "); AMI is not Go. See docs/language-mutability.md and docs/language-pointers.md."
		apply(diag.Warn, "W_GO_SYNTAX_DETECTED", msg, pkgName, filePath, nil)
	}
	// Parser-level errors captured already during AST creation
	p := parser.New(src)
	_ = p.ParseFile()
	for _, e := range p.Errors() {
		if e.File == "" || e.File == "input.ami" {
			e.File = filePath
		}
		if e.Package == "" {
			e.Package = pkgName
		}
		// do not rewrite parser error severity/code; still honor file-local disables and workspace 'off'
		if disabled[e.Code] {
			continue
		}
		if sev, ok := cfg.severity[e.Code]; ok {
			switch strings.ToLower(sev) {
			case "off":
				continue
			case "error":
				e.Level = diag.Error
			case "warn":
				e.Level = diag.Warn
			case "info":
				e.Level = diag.Info
			}
		}
		out = append(out, e)
	}
	// Semantic analyzer
	semres := sem.AnalyzeFile(f)
	for _, e := range semres.Diagnostics {
		if e.File == "" {
			e.File = filePath
		}
		if e.Package == "" {
			e.Package = pkgName
		}
		if disabled[e.Code] {
			continue
		}
		if sev, ok := cfg.severity[e.Code]; ok {
			switch strings.ToLower(sev) {
			case "off":
				continue
			case "error":
				e.Level = diag.Error
			case "warn":
				e.Level = diag.Warn
			case "info":
				e.Level = diag.Info
			}
		}
		out = append(out, e)
	}
	// Naming: package should be lowercase per style
	if f.Package != strings.ToLower(f.Package) && f.Package != "" {
		// find the position of the package identifier via scanning tokens
		if poff := findPackageIdentOffset(src); poff >= 0 {
			apply(diag.Warn, "W_PKG_LOWERCASE", "package name should be lowercase", pkgName, filePath, toPos(poff))
		} else {
			apply(diag.Warn, "W_PKG_LOWERCASE", "package name should be lowercase", pkgName, filePath, nil)
		}
	}
	// Hygiene: TODO/FIXME policy in comments (line comments only)
	// We conservatively scan for // and then look for TODO/FIXME tokens.
	base := 0
	for _, line := range strings.Split(src, "\n") {
		if idx := strings.Index(line, "//"); idx >= 0 {
			c := line[idx+2:]
			lower := strings.ToLower(c)
			if j := strings.Index(lower, "todo"); j >= 0 {
				apply(diag.Warn, "W_TODO_COMMENT", "TODO found in comment", pkgName, filePath, toPos(base+idx+2+j))
			}
			if j := strings.Index(lower, "fixme"); j >= 0 {
				apply(diag.Warn, "W_FIXME_COMMENT", "FIXME found in comment", pkgName, filePath, toPos(base+idx+2+j))
			}
		}
		base += len(line) + 1
	}
	// Imports: ordering, duplicates, relative paths, and unused
	// Collect imports with alias
	type impInfo struct{ path, alias string }
	var imports []impInfo
	seenPath := map[string]bool{}
	aliasToPath := map[string]string{}
	for _, d := range f.Decls {
		if id, ok := d.(astpkg.ImportDecl); ok {
			alias := id.Alias
			if alias == "" {
				alias = pathBase(id.Path)
			}
			imports = append(imports, impInfo{path: id.Path, alias: alias})
			// relative imports are disallowed
			if strings.HasPrefix(id.Path, "./") {
				if pos := firstImportPathOffset(src, id.Path); pos >= 0 {
					apply(diag.Warn, "W_REL_IMPORT", "relative import paths are disallowed; use workspace package name", pkgName, filePath, toPos(pos))
				} else {
					apply(diag.Warn, "W_REL_IMPORT", "relative import paths are disallowed; use workspace package name", pkgName, filePath, nil)
				}
			}
			if seenPath[id.Path] {
				// attach pos of this duplicate occurrence (second and later)
				if pos := firstImportPathOffset(src, id.Path); pos >= 0 {
					apply(diag.Warn, "W_DUP_IMPORT", fmt.Sprintf("duplicate import %q", id.Path), pkgName, filePath, toPos(pos))
				} else {
					apply(diag.Warn, "W_DUP_IMPORT", fmt.Sprintf("duplicate import %q", id.Path), pkgName, filePath, nil)
				}
			}
			seenPath[id.Path] = true
			// duplicate alias mapped to different paths
			if prev, ok := aliasToPath[alias]; ok && prev != id.Path {
				// position of the alias token for this import if available
				if pos := importAliasOffset(src, alias, id.Path); pos >= 0 {
					apply(diag.Warn, "W_DUP_IMPORT_ALIAS", fmt.Sprintf("alias %q used for multiple imports (%s, %s)", alias, prev, id.Path), pkgName, filePath, toPos(pos))
				} else {
					apply(diag.Warn, "W_DUP_IMPORT_ALIAS", fmt.Sprintf("alias %q used for multiple imports (%s, %s)", alias, prev, id.Path), pkgName, filePath, nil)
				}
			} else if !ok {
				aliasToPath[alias] = id.Path
			}
		}
	}
	// Import order rule: paths should be lexicographically sorted
	if len(imports) >= 2 {
		var paths []string
		for _, im := range imports {
			paths = append(paths, im.path)
		}
		exp := append([]string(nil), paths...)
		sort.Strings(exp)
		equal := len(paths) == len(exp)
		for i := range paths {
			if i >= len(exp) || paths[i] != exp[i] {
				equal = false
				break
			}
		}
		if !equal {
			// find first mismatch and use position of that path
			var bad string
			for i := range paths {
				if paths[i] != exp[i] {
					bad = paths[i]
					break
				}
			}
			if pos := firstImportPathOffset(src, bad); pos >= 0 {
				apply(diag.Warn, "W_IMPORT_ORDER", "imports are not ordered lexicographically by path", pkgName, filePath, toPos(pos))
			} else {
				apply(diag.Warn, "W_IMPORT_ORDER", "imports are not ordered lexicographically by path", pkgName, filePath, nil)
			}
		}
	}
	// Build used identifier set from function bodies and pipeline args
	used := map[string]bool{}
	for _, d := range f.Decls {
		if fd, ok := d.(astpkg.FuncDecl); ok {
			for _, t := range fd.Body {
				if t.Kind == tok.IDENT {
					used[t.Lexeme] = true
				}
			}
		}
		if pd, ok := d.(astpkg.PipelineDecl); ok {
            markArgs := func(steps []astpkg.NodeCall) {
                for _, st := range steps {
                    for _, a := range st.Args {
                        for _, tok := range splitIdents(a) {
                            used[tok] = true
                        }
                    }
                    // also scan structured attribute values for identifiers (e.g., worker=Name, in=edge.Pipeline(name=Upstream))
                    for _, v := range st.Attrs {
                        for _, tok := range splitIdents(v) { used[tok] = true }
                    }
                }
            }
			markArgs(pd.Steps)
			markArgs(pd.ErrorSteps)
		}
	}
	for _, im := range imports {
		if im.alias == "_" {
			continue
		} // parser/sem already error; skip here
		if !used[im.alias] {
			// position: prefer alias token if present, otherwise the path
			if ao := importAliasOffset(src, im.alias, im.path); ao >= 0 {
				apply(diag.Warn, "W_UNUSED_IMPORT", fmt.Sprintf("import %q (%s) not used", im.path, im.alias), pkgName, filePath, toPos(ao))
			} else if po := firstImportPathOffset(src, im.path); po >= 0 {
				apply(diag.Warn, "W_UNUSED_IMPORT", fmt.Sprintf("import %q (%s) not used", im.path, im.alias), pkgName, filePath, toPos(po))
			} else {
				apply(diag.Warn, "W_UNUSED_IMPORT", fmt.Sprintf("import %q (%s) not used", im.path, im.alias), pkgName, filePath, nil)
			}
		}
	}
	// Formatting: final newline and CRLF
	if !strings.HasSuffix(src, "\n") {
		apply(diag.Warn, "W_FILE_NO_NEWLINE", "file does not end with newline", pkgName, filePath, toPos(len(src)))
	}
	if strings.Contains(src, "\r\n") {
		if i := strings.Index(src, "\r\n"); i >= 0 {
			apply(diag.Warn, "W_FILE_CRLF", "file contains CRLF line endings; use LF", pkgName, filePath, toPos(i))
		} else {
			apply(diag.Warn, "W_FILE_CRLF", "file contains CRLF line endings; use LF", pkgName, filePath, nil)
		}
	}
	// Pipeline hints: attach positions for ingress/egress when present
	// Find first KW_INGRESS and first KW_EGRESS tokens.
	{
		s := scannerFor(src)
		seenIngress := false
		for {
			t := s.Next()
			if t.Kind == tok.EOF {
				break
			}
			if !seenIngress && t.Kind == tok.KW_INGRESS {
				p := srcset.NewFileSet()
				sf := p.AddFileFromSource(filePath, src)
				pos := sf.PositionFor(t.Offset)
				out = append(out, diag.Diagnostic{Level: diag.Info, Code: "W_PIPELINE_INGRESS_POS", Message: "ingress position", Package: pkgName, File: filePath, Pos: &pos})
				seenIngress = true
			}
			if t.Kind == tok.KW_EGRESS {
				p := srcset.NewFileSet()
				sf := p.AddFileFromSource(filePath, src)
				pos := sf.PositionFor(t.Offset)
				out = append(out, diag.Diagnostic{Level: diag.Info, Code: "W_PIPELINE_EGRESS_POS", Message: "egress position", Package: pkgName, File: filePath, Pos: &pos})
				// don't break; we want both if both appear
			}
		}
	}
	// Language-specific hints based on simple type usage in decls
	// Utility: render a type reference back to source-ish text
	var renderType func(tr astpkg.TypeRef) string
	renderType = func(tr astpkg.TypeRef) string {
		s := strings.Builder{}
		if tr.Ptr {
			s.WriteString("*")
		}
		if tr.Slice {
			s.WriteString("[]")
		}
		name := tr.Name
		s.WriteString(name)
		if len(tr.Args) > 0 {
			s.WriteString("<")
			for i, a := range tr.Args {
				if i > 0 {
					s.WriteString(", ")
				}
				s.WriteString(renderType(a))
			}
			s.WriteString(">")
		}
		return s.String()
	}
	// Utility: find offset of the rendered type text in source (best effort)
	findTypeOffset := func(src, typ string) int {
		if typ == "" {
			return -1
		}
		if i := strings.Index(src, typ); i >= 0 {
			return i
		}
		// Try without spaces in generics as fallback
		compact := strings.ReplaceAll(typ, ", ", ",")
		return strings.Index(src, compact)
	}
	// de-duplicate hints per file by (code,typeString)
	emitted := map[string]bool{}
	var hintType func(tr astpkg.TypeRef)
	hintType = func(tr astpkg.TypeRef) {
		tstr := renderType(tr)
		key := ""
		// prefer type-based dedup key when non-empty
		if tstr != "" {
			key = "|" + tstr
		}
		off := findTypeOffset(src, tstr)
		var p *srcset.Position
		// prefer exact AST offset when available
		if tr.Offset >= 0 {
			p = toPos(tr.Offset)
		} else if off >= 0 {
			p = toPos(off)
		}
		// no pointer type hints in AMI 2.3.2
		name := strings.ToLower(tr.Name)
		switch name {
		case "owned":
			if !emitted["W_RAII_OWNED_HINT"+key] {
				apply(diag.Info, "W_RAII_OWNED_HINT", "Owned<T> should be explicitly released: use mutate(release(x)) or an equivalent explicit release call.", pkgName, filePath, p)
				emitted["W_RAII_OWNED_HINT"+key] = true
			}
		case "map":
			if len(tr.Args) != 2 {
				if !emitted["W_MAP_ARITY_HINT"+key] {
					apply(diag.Warn, "W_MAP_ARITY_HINT", "map requires two type parameters: map<K,V>", pkgName, filePath, p)
					emitted["W_MAP_ARITY_HINT"+key] = true
				}
			} else {
				k := tr.Args[0]
				if k.Ptr || k.Slice || strings.ToLower(k.Name) == "map" || strings.ToLower(k.Name) == "slice" {
					if !emitted["W_MAP_KEY_TYPE_HINT"+key] {
						apply(diag.Warn, "W_MAP_KEY_TYPE_HINT", "map key should be scalar (string/int); avoid pointers, slices, or maps", pkgName, filePath, p)
						emitted["W_MAP_KEY_TYPE_HINT"+key] = true
					}
				}
			}
		case "set":
			if len(tr.Args) != 1 {
				if !emitted["W_SET_ARITY_HINT"+key] {
					apply(diag.Warn, "W_SET_ARITY_HINT", "set requires one type parameter: set<T>", pkgName, filePath, p)
					emitted["W_SET_ARITY_HINT"+key] = true
				}
			} else {
				e := tr.Args[0]
				if e.Ptr || e.Slice || strings.ToLower(e.Name) == "map" || strings.ToLower(e.Name) == "slice" {
					if !emitted["W_SET_ELEM_TYPE_HINT"+key] {
						apply(diag.Warn, "W_SET_ELEM_TYPE_HINT", "set element should be scalar; avoid pointers, slices, or maps", pkgName, filePath, p)
						emitted["W_SET_ELEM_TYPE_HINT"+key] = true
					}
				}
			}
		case "slice":
			if len(tr.Args) != 1 {
				if !emitted["W_SLICE_ARITY_HINT"+key] {
					apply(diag.Warn, "W_SLICE_ARITY_HINT", "slice requires one type parameter: slice<T>", pkgName, filePath, p)
					emitted["W_SLICE_ARITY_HINT"+key] = true
				}
			}
		}
		for _, a := range tr.Args {
			hintType(a)
		}
	}
    // Pointer dereference is not part of AMI (2.3.2); no deref hints
    for _, d := range f.Decls {
        if fd, ok := d.(astpkg.FuncDecl); ok {
            // Soft migration: suggest removing explicit State parameter (ambient state)
            // Be tolerant of parser idiosyncrasies: detect legacy 3-parameter worker signature
            legacy3Param := func(fd astpkg.FuncDecl) (bool, *srcset.Position) {
                if len(fd.Params) != 3 { return false, nil }
                p1 := fd.Params[0].Type
                p2 := fd.Params[1].Type
                p3 := fd.Params[2]
                if !(p1.Name == "Context" && !p1.Ptr && !p1.Slice) { return false, nil }
                if !(p2.Name == "Event" && len(p2.Args) == 1 && !p2.Ptr) { return false, nil }
                // Third param: type may be mis-attributed; accept either type name State or identifier/name 'st'
                isStateType := strings.EqualFold(p3.Type.Name, "state")
                isStateIdent := strings.EqualFold(p3.Name, "st") || strings.EqualFold(p3.Type.Name, "st")
                if !(isStateType || isStateIdent) { return false, nil }
                // Prefer type offset when available; otherwise function-level position
                if p3.Type.Offset >= 0 { return true, toPos(p3.Type.Offset) }
                return true, nil
            }
            if ok, pos := legacy3Param(fd); ok {
                apply(diag.Info, "W_STATE_PARAM_AMBIENT_SUGGEST", "State is ambient; prefer ambient access (state.get/set/update/list) over passing State parameter", pkgName, filePath, pos)
            } else {
                // Fallback: direct scan for explicit non-pointer State params when signature is not the legacy worker form
                for _, prm := range fd.Params {
                    if strings.EqualFold(prm.Type.Name, "state") && !prm.Type.Ptr {
                        var p *srcset.Position
                        if prm.Type.Offset >= 0 { p = toPos(prm.Type.Offset) }
                        apply(diag.Info, "W_STATE_PARAM_AMBIENT_SUGGEST", "State is ambient; prefer ambient access (state.get/set/update/list) over passing State parameter", pkgName, filePath, p)
                    }
                }
            }
            for _, p := range fd.Params { hintType(p.Type) }
            for _, r := range fd.Result { hintType(r) }
        }
		if sd, ok := d.(astpkg.StructDecl); ok {
			for _, fld := range sd.Fields {
				hintType(fld.Type)
			}
		}
		if pd, ok := d.(astpkg.PipelineDecl); ok {
			// Node with no workers on transform/fanout: warn
            for _, st := range pd.Steps {
                kind := strings.ToLower(st.Name)
                if (kind == "transform" || kind == "fanout") && len(st.Workers) == 0 && st.InlineWorker == nil && strings.TrimSpace(st.Attrs["worker"]) == "" {
                    apply(diag.Warn, "W_NODE_NO_WORKERS", "node has no workers; verify configuration", pkgName, filePath, nil)
                }
                // MultiPath hints (tolerant)
                if v := strings.TrimSpace(st.Attrs["in"]); strings.HasPrefix(v, "edge.MultiPath(") {
                    mp, ok := parseMultiPathFromValueLint(v)
                    if !ok {
                        apply(diag.Warn, "W_MP_INVALID", "invalid edge.MultiPath specification", pkgName, filePath, nil)
                    } else {
                        if kind != "collect" {
                            apply(diag.Warn, "W_MP_ONLY_COLLECT", "edge.MultiPath should be used on Collect nodes", pkgName, filePath, nil)
                        }
                        if len(mp.Inputs) == 0 {
                            apply(diag.Warn, "W_MP_INPUTS_EMPTY", "edge.MultiPath has no inputs", pkgName, filePath, nil)
                        } else {
                            if mp.Inputs[0].Kind != "fifo" {
                                apply(diag.Warn, "W_MP_INPUT0_KIND", "first MultiPath input should be a FIFO default upstream edge", pkgName, filePath, nil)
                            }
                            // Smells on inputs
                            for _, in := range mp.Inputs {
                                if in.Backpressure == "block" && in.MaxCapacity == 0 {
                                    apply(diag.Warn, "W_MP_EDGE_SMELL_UNBOUNDED_BLOCK", "MultiPath input uses block with unbounded capacity", pkgName, filePath, nil)
                                }
                                if in.Backpressure == "drop" && in.MaxCapacity > 0 && in.MaxCapacity <= 1 {
                                    apply(diag.Warn, "W_MP_EDGE_SMELL_TINY_BOUNDED_DROP", "MultiPath input uses 'drop' with tiny bounded capacity (<=1)", pkgName, filePath, nil)
                                }
                            }
                            // Type mismatch hint when provided
                            base := ""
                            for _, in := range mp.Inputs {
                                if in.Type == "" { continue }
                                if base == "" { base = in.Type; continue }
                                if base != in.Type { apply(diag.Warn, "W_MP_INPUT_TYPE_MISMATCH", "MultiPath inputs have mismatched payload types", pkgName, filePath, nil); break }
                            }
                        }
                        // Merge op hints
                        if len(mp.Merge) == 0 {
                            apply(diag.Info, "W_MP_MERGE_SUGGEST", "consider specifying merge behavior (e.g., merge.Sort, merge.Stable)", pkgName, filePath, nil)
                        } else {
                            for _, op := range mp.Merge {
                                if !mpAllowedMerge(op.Name) { apply(diag.Warn, "W_MP_MERGE_INVALID", "invalid merge operator in MultiPath", pkgName, filePath, nil) }
                            }
                        }
                    }
                }
                // Edge smell: unbounded capacity with block backpressure
                if spec, ok := parseEdgeSpecFromNodeLint(st); ok {
                    if (spec.Kind == "fifo" || spec.Kind == "lifo" || spec.Kind == "pipeline") && strings.ToLower(spec.Backpressure) == "block" && spec.MaxCapacity == 0 {
                        apply(diag.Warn, "W_EDGE_SMELL_UNBOUNDED_BLOCK", "edge configured with block backpressure and unbounded capacity", pkgName, filePath, nil)
                    }
                    // Edge smell: tiny bounded capacity with drop backpressure (likely data loss amplification)
                    if (spec.Kind == "fifo" || spec.Kind == "lifo" || spec.Kind == "pipeline") && strings.ToLower(spec.Backpressure) == "drop" && spec.MaxCapacity > 0 && spec.MaxCapacity <= 1 {
                        apply(diag.Warn, "W_EDGE_SMELL_TINY_BOUNDED_DROP", "edge uses 'drop' backpressure with tiny bounded capacity (<=1)", pkgName, filePath, nil)
                    }
                    // Backpressure tokens alignment: warn on legacy ambiguous 'drop'
                    if strings.ToLower(spec.Backpressure) == "drop" {
                        apply(diag.Warn, "W_EDGE_BP_AMBIGUOUS_DROP", "backpressure=drop is ambiguous; use dropOldest or dropNewest", pkgName, filePath, nil)
                    }
                }
            }
            for _, st := range pd.ErrorSteps {
                kind := strings.ToLower(st.Name)
                if (kind == "transform" || kind == "fanout") && len(st.Workers) == 0 && st.InlineWorker == nil && strings.TrimSpace(st.Attrs["worker"]) == "" {
                    apply(diag.Warn, "W_NODE_NO_WORKERS", "error-path node has no workers; verify configuration", pkgName, filePath, nil)
                }
                if v := strings.TrimSpace(st.Attrs["in"]); strings.HasPrefix(v, "edge.MultiPath(") {
                    mp, ok := parseMultiPathFromValueLint(v)
                    if !ok {
                        apply(diag.Warn, "W_MP_INVALID", "invalid edge.MultiPath specification", pkgName, filePath, nil)
                    } else {
                        if kind != "collect" {
                            apply(diag.Warn, "W_MP_ONLY_COLLECT", "edge.MultiPath should be used on Collect nodes", pkgName, filePath, nil)
                        }
                        if len(mp.Inputs) == 0 { apply(diag.Warn, "W_MP_INPUTS_EMPTY", "edge.MultiPath has no inputs", pkgName, filePath, nil) }
                        for _, in := range mp.Inputs {
                            if in.Backpressure == "block" && in.MaxCapacity == 0 { apply(diag.Warn, "W_MP_EDGE_SMELL_UNBOUNDED_BLOCK", "MultiPath input uses block with unbounded capacity (error path)", pkgName, filePath, nil) }
                            if in.Backpressure == "drop" && in.MaxCapacity > 0 && in.MaxCapacity <= 1 { apply(diag.Warn, "W_MP_EDGE_SMELL_TINY_BOUNDED_DROP", "MultiPath input uses 'drop' with tiny bounded capacity (<=1) (error path)", pkgName, filePath, nil) }
                        }
                    }
                }
                if spec, ok := parseEdgeSpecFromNodeLint(st); ok {
                    if (spec.Kind == "fifo" || spec.Kind == "lifo" || spec.Kind == "pipeline") && strings.ToLower(spec.Backpressure) == "block" && spec.MaxCapacity == 0 {
                        apply(diag.Warn, "W_EDGE_SMELL_UNBOUNDED_BLOCK", "edge configured with block backpressure and unbounded capacity (error path)", pkgName, filePath, nil)
                    }
                    if (spec.Kind == "fifo" || spec.Kind == "lifo" || spec.Kind == "pipeline") && strings.ToLower(spec.Backpressure) == "drop" && spec.MaxCapacity > 0 && spec.MaxCapacity <= 1 {
                        apply(diag.Warn, "W_EDGE_SMELL_TINY_BOUNDED_DROP", "edge uses 'drop' backpressure with tiny bounded capacity (<=1) (error path)", pkgName, filePath, nil)
                    }
                    if strings.ToLower(spec.Backpressure) == "drop" {
                        apply(diag.Warn, "W_EDGE_BP_AMBIGUOUS_DROP", "backpressure=drop is ambiguous; use dropOldest or dropNewest", pkgName, filePath, nil)
                    }
                }
            }
        }
	}
	return out
}

// --- Minimal edge spec parser for lint smells (tolerant) ---
type edgeSpecLint struct {
	Kind         string
	Name         string
	MinCapacity  int
	MaxCapacity  int
	Backpressure string
	Type         string
}

func parseEdgeSpecFromNodeLint(st astpkg.NodeCall) (edgeSpecLint, bool) {
    // Prefer structured attribute if present
    if v := strings.TrimSpace(st.Attrs["in"]); v != "" {
        return parseEdgeSpecFromValueLint(v)
    }
    return parseEdgeSpecFromArgsLint(st.Args)
}

func parseEdgeSpecFromValueLint(v string) (edgeSpecLint, bool) {
    s := strings.TrimSpace(v)
    switch {
    case strings.HasPrefix(s, "edge.FIFO(") && strings.HasSuffix(s, ")"):
        m := parseKVListLint(s[len("edge.FIFO(") : len(s)-1])
        return edgeSpecLint{Kind: "fifo", MinCapacity: atoi0(m["minCapacity"]), MaxCapacity: atoi0(m["maxCapacity"]), Backpressure: m["backpressure"], Type: m["type"]}, true
    case strings.HasPrefix(s, "edge.LIFO(") && strings.HasSuffix(s, ")"):
        m := parseKVListLint(s[len("edge.LIFO(") : len(s)-1])
        return edgeSpecLint{Kind: "lifo", MinCapacity: atoi0(m["minCapacity"]), MaxCapacity: atoi0(m["maxCapacity"]), Backpressure: m["backpressure"], Type: m["type"]}, true
    case strings.HasPrefix(s, "edge.Pipeline(") && strings.HasSuffix(s, ")"):
        m := parseKVListLint(s[len("edge.Pipeline(") : len(s)-1])
        return edgeSpecLint{Kind: "pipeline", Name: m["name"], MinCapacity: atoi0(m["minCapacity"]), MaxCapacity: atoi0(m["maxCapacity"]), Backpressure: m["backpressure"], Type: m["type"]}, true
    }
    return edgeSpecLint{}, false
}

func parseEdgeSpecFromArgsLint(args []string) (edgeSpecLint, bool) {
	for _, a := range args {
		s := strings.TrimSpace(a)
		if !strings.HasPrefix(s, "in=") {
			continue
		}
		v := strings.TrimPrefix(s, "in=")
		switch {
		case strings.HasPrefix(v, "edge.FIFO(") && strings.HasSuffix(v, ")"):
			m := parseKVListLint(v[len("edge.FIFO(") : len(v)-1])
			return edgeSpecLint{Kind: "fifo", MinCapacity: atoi0(m["minCapacity"]), MaxCapacity: atoi0(m["maxCapacity"]), Backpressure: m["backpressure"], Type: m["type"]}, true
		case strings.HasPrefix(v, "edge.LIFO(") && strings.HasSuffix(v, ")"):
			m := parseKVListLint(v[len("edge.LIFO(") : len(v)-1])
			return edgeSpecLint{Kind: "lifo", MinCapacity: atoi0(m["minCapacity"]), MaxCapacity: atoi0(m["maxCapacity"]), Backpressure: m["backpressure"], Type: m["type"]}, true
		case strings.HasPrefix(v, "edge.Pipeline(") && strings.HasSuffix(v, ")"):
			m := parseKVListLint(v[len("edge.Pipeline(") : len(v)-1])
			return edgeSpecLint{Kind: "pipeline", Name: m["name"], MinCapacity: atoi0(m["minCapacity"]), MaxCapacity: atoi0(m["maxCapacity"]), Backpressure: m["backpressure"], Type: m["type"]}, true
		}
	}
	return edgeSpecLint{}, false
}

// --- MultiPath tolerant parser for lint hints ---
type mpInputLint struct{ edgeSpecLint }
type mpSpecLint struct{
    Inputs []edgeSpecLint
    Merge  []mergeOpLint
}
type mergeOpLint struct{ Name string }

func parseMultiPathFromValueLint(v string) (mpSpecLint, bool) {
    s := strings.TrimSpace(v)
    if !strings.HasPrefix(s, "edge.MultiPath(") || !strings.HasSuffix(s, ")") { return mpSpecLint{}, false }
    inner := s[len("edge.MultiPath(") : len(s)-1]
    idx := strings.Index(inner, "inputs=")
    if idx < 0 { return mpSpecLint{}, false }
    after := strings.TrimSpace(inner[idx+len("inputs="):])
    if len(after) == 0 || after[0] != '[' { return mpSpecLint{}, false }
    // capture bracket block
    i := 1
    depth := 1
    for i < len(after) && depth > 0 {
        switch after[i] {
        case '[': depth++
        case ']': depth--
        }
        i++
    }
    if depth != 0 { return mpSpecLint{}, false }
    list := after[1 : i-1]
    parts := splitTopLevelCommasLint(list)
    var inputs []edgeSpecLint
    for _, p := range parts {
        p = strings.TrimSpace(p)
        if p == "" { continue }
        if es, ok := parseEdgeSpecFromValueLint(p); ok { inputs = append(inputs, es) }
    }
    // optional merge
    rest := strings.TrimSpace(after[i:])
    var merge []mergeOpLint
    if j := strings.Index(rest, "merge="); j >= 0 {
        mv := strings.TrimSpace(rest[j+len("merge="):])
        if k := strings.IndexByte(mv, '('); k > 0 {
            name := strings.TrimSpace(mv[:k])
            merge = append(merge, mergeOpLint{Name: name})
        }
    }
    return mpSpecLint{Inputs: inputs, Merge: merge}, true
}

func mpAllowedMerge(name string) bool {
    n := strings.ToLower(strings.TrimPrefix(name, "merge."))
    switch n {
    case "", "sort", "stable", "key", "dedup", "window", "watermark", "timeout", "buffer", "partitionby":
        return n != ""
    default:
        return false
    }
}

func parseKVListLint(s string) map[string]string {
	out := map[string]string{}
	parts := splitTopLevelCommasLint(s)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if eq := strings.IndexByte(p, '='); eq >= 0 {
			k := strings.TrimSpace(p[:eq])
			v := strings.TrimSpace(p[eq+1:])
			if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
				v = v[1 : len(v)-1]
			}
			out[k] = v
		}
	}
	return out
}

func splitTopLevelCommasLint(s string) []string {
	var out []string
	depth := 0
	last := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				out = append(out, s[last:i])
				last = i + 1
			}
		}
	}
	out = append(out, s[last:])
	return out
}

func atoi0(s string) int { n, _ := strconv.Atoi(s); return n }

func pathBase(p string) string {
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}

// splitIdents returns identifier-like tokens from a string (A-Za-z0-9_)
// splitIdents moved to split_idents.go

// uPackageFromPath maps a file path back to a package name using known roots.
func uPackageFromPath(file string, roots map[string]string) string {
	// Attempt: if file path has a root prefix, return that package
	// Iterate stable by package name for determinism
	var names []string
	for name := range roots {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		root := filepath.Clean(roots[name])
		if strings.HasPrefix(filepath.Clean(file), filepath.Clean(root)+string(os.PathSeparator)) {
			return name
		}
	}
	return ""
}

// ruleSelected applies the --rules filter (comma-separated substrings).
// ruleSelected moved to rule_selected.go

// findPackageIdentOffset returns the byte offset of the package identifier token in src.
func findPackageIdentOffset(src string) int {
	s := scannerFor(src)
	for {
		t := s.Next()
		if t.Kind == tok.EOF {
			break
		}
		if t.Kind == tok.KW_PACKAGE || (t.Kind == tok.IDENT && strings.ToLower(t.Lexeme) == "package") {
			t2 := s.Next()
			if t2.Kind == tok.IDENT {
				return t2.Offset
			}
			return -1
		}
	}
	return -1
}

// firstImportPathOffset returns the offset of the first occurrence of an import path.
// firstImportPathOffset moved to import_path_helpers.go

// importAliasOffset returns the offset of an import alias on the line importing the given path.
// importAliasOffset moved to import_path_helpers.go

// scannerFor moved to scanner_for.go

func unquoteSimple(s string) string {
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		return s[1 : len(s)-1]
	}
	return s
}

// parseLinterConfig extracts rule severity overrides from workspace: toolchain.linter.rules
// parseLinterConfig moved to parse_linter_config.go

// lintCrossPackageConstraints checks that workspace-local imports satisfy the
// version constraints declared in the importing package's workspace entry.
func lintCrossPackageConstraints(ws *workspace.Workspace, pkgRoots map[string]string, units []lintUnitRec) []diag.Diagnostic {
	var out []diag.Diagnostic
	// Build local package versions and importer constraints maps
	localVer := map[string]string{}
	importerCons := map[string]map[string]string{}
	for _, p := range ws.Packages {
		m, ok := p.(map[string]any)
		if !ok {
			continue
		}
		for name, v := range m {
			vm, ok := v.(map[string]any)
			if !ok {
				continue
			}
			if ver, ok := vm["version"].(string); ok {
				if strings.TrimSpace(ver) != "" {
					localVer[name] = normSemVer(ver)
				}
			}
			if imv, ok := vm["import"]; ok && imv != nil {
				if lst, ok := imv.([]any); ok {
					for _, item := range lst {
						s, ok := item.(string)
						if !ok {
							continue
						}
						fields := strings.Fields(s)
						if len(fields) == 0 {
							continue
						}
						path := fields[0]
						cons := "==latest"
						if len(fields) > 1 {
							cons = strings.ReplaceAll(strings.Join(fields[1:], ""), " ", "")
						}
						if importerCons[name] == nil {
							importerCons[name] = map[string]string{}
						}
						importerCons[name][path] = cons
					}
				}
			}
		}
	}
	// For each unit's imports, if it's a local package and a constraint exists, verify
	consByTarget := map[string]map[string]bool{}
	for _, u := range units {
		for _, imp := range u.imports {
			if _, ok := pkgRoots[imp]; !ok {
				continue
			} // external or unknown
			cons, hasCons := importerCons[u.pkg][imp]
			if !hasCons {
				continue
			}
			ver, ok := localVer[imp]
			if !ok {
				continue
			}
			if !satisfiesConstraint(ver, cons) {
				// attach position at import path occurrence when possible
				pos := firstImportPathOffset(u.src, imp)
				d := diag.Diagnostic{Level: diag.Error, Code: "E_IMPORT_CONSTRAINT", Message: fmt.Sprintf("import %q version %s does not satisfy constraint %q", imp, ver, cons), Package: u.pkg, File: u.file}
				if pos >= 0 {
					fs := srcset.NewFileSet()
					sf := fs.AddFileFromSource(u.file, u.src)
					p := sf.PositionFor(pos)
					d.Pos = &p
				}
				out = append(out, d)
				continue
			}
			// prerelease forbidden when constraint omits prerelease
			if _, _, _, pre := parseSemVer(ver); pre != "" {
				if cons != "==latest" && !strings.Contains(cons, "-") {
					pos := firstImportPathOffset(u.src, imp)
					d := diag.Diagnostic{Level: diag.Error, Code: "E_IMPORT_PRERELEASE_FORBIDDEN", Message: fmt.Sprintf("import %q prerelease %s is forbidden by constraint %q (no prerelease allowed)", imp, ver, cons), Package: u.pkg, File: u.file}
					if pos >= 0 {
						fs := srcset.NewFileSet()
						sf := fs.AddFileFromSource(u.file, u.src)
						p := sf.PositionFor(pos)
						d.Pos = &p
					}
					out = append(out, d)
				}
			}
			if consByTarget[imp] == nil {
				consByTarget[imp] = map[string]bool{}
			}
			consByTarget[imp][cons] = true
		}
	}
	// Consistency check: same package should have a single constraint across importers
	for target, set := range consByTarget {
		if len(set) > 1 {
			var list []string
			for c := range set {
				list = append(list, c)
			}
			sort.Strings(list)
			canonical := list[0]
			for _, u := range units {
				cons := importerCons[u.pkg][target]
				if cons == "" || cons == canonical {
					continue
				}
				pos := firstImportPathOffset(u.src, target)
				d := diag.Diagnostic{Level: diag.Error, Code: "E_IMPORT_CONSISTENCY", Message: fmt.Sprintf("workspace imports %q with conflicting constraints (%q vs %q)", target, cons, canonical), Package: u.pkg, File: u.file}
				if pos >= 0 {
					fs := srcset.NewFileSet()
					sf := fs.AddFileFromSource(u.file, u.src)
					p := sf.PositionFor(pos)
					d.Pos = &p
				}
				out = append(out, d)
			}
		}
	}
	return out
}

// normSemVer ensures a leading 'v' for compare.
func normSemVer(v string) string {
	if strings.HasPrefix(v, "v") {
		return v
	}
	return "v" + v
}

// parseSemVer returns [major,minor,patch] and prerelease string.
func parseSemVer(v string) (int, int, int, string) {
	v = strings.TrimPrefix(v, "v")
	// strip build metadata
	if i := strings.IndexByte(v, '+'); i >= 0 {
		v = v[:i]
	}
	pre := ""
	if i := strings.IndexByte(v, '-'); i >= 0 {
		pre = v[i+1:]
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	if len(parts) < 3 {
		return 0, 0, 0, pre
	}
	a := atoi(parts[0])
	b := atoi(parts[1])
	c := atoi(parts[2])
	return a, b, c, pre
}

func atoi(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func semverLess(a, b string) bool {
	ma, na, pa, _ := parseSemVer(a)
	mb, nb, pb, _ := parseSemVer(b)
	if ma != mb {
		return ma < mb
	}
	if na != nb {
		return na < nb
	}
	if pa != pb {
		return pa < pb
	}
	return false
}

func satisfiesConstraint(ver, cons string) bool {
	// Always true for latest macro
	if cons == "==latest" {
		return true
	}
	ver = normSemVer(ver)
	cons = strings.TrimSpace(cons)
	// Exact version (semver)
	if isSemVer(cons) {
		return !semverLess(ver, normSemVer(cons)) && !semverLess(normSemVer(cons), ver)
	}
	// >=x.y.z
	if strings.HasPrefix(cons, ">=") {
		base := normSemVer(strings.TrimPrefix(cons, ">="))
		return !semverLess(ver, base)
	}
	// >x.y.z
	if strings.HasPrefix(cons, ">") {
		base := normSemVer(strings.TrimPrefix(cons, ">"))
		return semverLess(base, ver)
	}
	// ^x.y.z: same major, >= base
	if strings.HasPrefix(cons, "^") {
		base := normSemVer(strings.TrimPrefix(cons, "^"))
		mv, mn, mp, _ := parseSemVer(ver)
		bv, bn, bp, _ := parseSemVer(base)
		if mv != bv {
			return false
		}
		// ver >= base
		if mn != bn {
			return mn > bn
		}
		return mp >= bp
	}
	// ~x.y.z: same major+minor, >= base
	if strings.HasPrefix(cons, "~") {
		base := normSemVer(strings.TrimPrefix(cons, "~"))
		mv, mn, mp, _ := parseSemVer(ver)
		bv, bn, bp, _ := parseSemVer(base)
		if mv != bv || mn != bn {
			return false
		}
		return mp >= bp
	}
	return true
}

func isSemVer(s string) bool {
	// copied from workspace semantic: ^v?\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$
	re := regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)
	return re.MatchString(s)
}

// toStringSlice helps parse list-like values from workspace config
func toStringSlice(v any) []string {
	switch a := v.(type) {
	case []any:
		var out []string
		for _, e := range a {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return append([]string(nil), a...)
	default:
		return nil
	}
}

// lintAlias maps existing rule codes to a forward-compat LINT_* namespace
func lintAlias(code string) string {
	if strings.HasPrefix(code, "LINT_") {
		return code
	}
	return "LINT_" + code
}
