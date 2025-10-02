#!/usr/bin/env bash
set -euo pipefail

# check-single-declaration-per-file.sh
#
# Ensures Go files follow our "single declaration per file" guideline and that
# each non-test source file has a corresponding _test.go.
#
# Rules (default, cohesive mode):
# - A file may contain exactly one of these cohesive units:
#   - A single type plus any number of methods with that type as receiver
#     and optional const blocks whose specs are typed to that type (enum values).
#   - Only methods, all with the same receiver type (type declared elsewhere).
#   - A single non-method function.
#   - Only const blocks (treated as one enum “unit”).
# - doc.go and files with //go:embed are ignored.
# - Generated files (containing both "Code generated" and "DO NOT EDIT") are ignored.
# - By default, each non-test source file that declares a type or function must
#   have a corresponding sibling test file named <name>_test.go.
#
# Configuration via env vars:
# - CHECK_MODE:     "cohesive" (default) or "strict".
#                   strict = exactly one top-level decl (func/type/const-block).
# - CHECK_TEST_MODE: "per-file" (default) or "package".
#                   package = require at least one *_test.go in same dir.
# - CHECK_VERBOSE:  set to 1 for verbose output.
#
# Usage: scripts/check-single-declaration-per-file.sh [path ...]
# If no paths given, scans the repo for .go files (excluding vendor/build/.git).

CHECK_MODE=${CHECK_MODE:-cohesive}
CHECK_TEST_MODE=${CHECK_TEST_MODE:-per-file}
CHECK_VERBOSE=${CHECK_VERBOSE:-0}

if [[ "${CHECK_MODE}" != "cohesive" && "${CHECK_MODE}" != "strict" ]]; then
  echo "ERROR: CHECK_MODE must be 'cohesive' or 'strict' (got ${CHECK_MODE})" >&2
  exit 2
fi

if [[ "${CHECK_TEST_MODE}" != "per-file" && "${CHECK_TEST_MODE}" != "package" ]]; then
  echo "ERROR: CHECK_TEST_MODE must be 'per-file' or 'package' (got ${CHECK_TEST_MODE})" >&2
  exit 2
fi

if ! command -v go >/dev/null 2>&1; then
  echo "ERROR: Go toolchain not found in PATH" >&2
  exit 2
fi

# Determine repo root (prefer git; fallback to script dir parent)
REPO_ROOT=""
if command -v git >/dev/null 2>&1 && git rev-parse --show-toplevel >/dev/null 2>&1; then
  REPO_ROOT=$(git rev-parse --show-toplevel)
else
  # Resolve to this script's directory, then go up to repo root (assumed)
  SCRIPT_DIR=$(cd -- "$(dirname -- "$0")" && pwd)
  REPO_ROOT=$(cd -- "${SCRIPT_DIR}/.." && pwd)
fi
cd "${REPO_ROOT}"

# Collect Go files
declare -a GO_FILES
if [[ $# -gt 0 ]]; then
  # Use provided paths
  while IFS= read -r -d '' f; do GO_FILES+=("$f"); done < <(\
    rg --files -g '!**/vendor/**' -g '!**/build/**' -g '!**/.git/**' -t go "$@" -0 || true)
else
  while IFS= read -r -d '' f; do GO_FILES+=("$f"); done < <(\
    rg --files -g '!**/vendor/**' -g '!**/build/**' -g '!**/.git/**' -t go -0 || true)
fi

if [[ ${#GO_FILES[@]} -eq 0 ]]; then
  [[ "${CHECK_VERBOSE}" == "1" ]] && echo "No Go files found. Nothing to check." >&2
  exit 0
fi

# Build a small analyzer in a temp dir and run it with the file list on stdin.
TMPDIR=$(mktemp -d 2>/dev/null || mktemp -d -t amideclcheck)
cleanup() {
  rm -rf "${TMPDIR}";
}
trap cleanup EXIT

cat >"${TMPDIR}/declcheck.go" <<'EOF'
package main

import (
  "bufio"
  "errors"
  "flag"
  "fmt"
  "go/ast"
  "go/parser"
  "go/token"
  "os"
  "path/filepath"
  "strings"
)

var (
  mode       = flag.String("mode", "cohesive", "checking mode: cohesive|strict")
  testMode   = flag.String("test", "per-file", "test requirement: per-file|package")
  verbose    = flag.Bool("v", false, "verbose output")
)

type Result struct {
  Path               string
  IsTest             bool
  IsDoc              bool
  IsGenerated        bool
  HasGoEmbed         bool
  TypeNames          []string
  MethodRecvNames    []string
  NonMethodFuncNames []string
  ConstTypedNames    []string // types referenced by typed const specs
  ConstBlocks        int
}

func main() {
  flag.Parse()
  files := readInputFiles()
  if len(files) == 0 {
    if *verbose { fmt.Fprintln(os.Stderr, "no input files") }
    return
  }

  var violations int

  // Precompute per-directory test presence if package mode
  dirHasTests := map[string]bool{}
  if *testMode == "package" {
    for _, p := range files {
      if strings.HasSuffix(p, "_test.go") {
        dirHasTests[filepath.Dir(p)] = true
      }
    }
  }

  for _, path := range files {
    res, err := analyze(path)
    if err != nil {
      fmt.Fprintf(os.Stderr, "%s: parse error: %v\n", path, err)
      violations++
      continue
    }

    // Skip ignored categories
    if res.IsDoc || res.IsGenerated || res.HasGoEmbed {
      if *verbose { fmt.Fprintf(os.Stderr, "skip: %s\n", path) }
      continue
    }

    // Single-declaration check
    if !res.IsTest {
      if !checkSingleDecl(*mode, res) {
        reportSingleDeclViolation(res)
        violations++
      }
      // Test presence check
      if requiresTest(res) {
        if *testMode == "per-file" {
          exp := strings.TrimSuffix(res.Path, ".go") + "_test.go"
          if _, err := os.Stat(exp); err != nil {
            if errors.Is(err, os.ErrNotExist) {
              fmt.Printf("%s: missing corresponding test file %s\n", res.Path, exp)
              violations++
            } else {
              fmt.Fprintf(os.Stderr, "%s: test check error: %v\n", res.Path, err)
              violations++
            }
          }
        } else { // package
          if !dirHasTests[filepath.Dir(res.Path)] {
            fmt.Printf("%s: package has no *_test.go files\n", res.Path)
            violations++
          }
        }
      }
    }
  }

  if violations > 0 {
    fmt.Fprintf(os.Stderr, "Found %d violation(s)\n", violations)
    os.Exit(1)
  }
}

func readInputFiles() []string {
  // Prefer args; else read stdin newline-delimited
  var out []string
  if flag.NArg() > 0 {
    out = append(out, flag.Args()...)
    return out
  }
  s := bufio.NewScanner(os.Stdin)
  for s.Scan() {
    line := strings.TrimSpace(s.Text())
    if line != "" {
      out = append(out, line)
    }
  }
  return out
}

func analyze(path string) (Result, error) {
  res := Result{Path: path, IsTest: strings.HasSuffix(path, "_test.go")}
  base := filepath.Base(path)
  if strings.EqualFold(base, "doc.go") { res.IsDoc = true }

  src, err := os.ReadFile(path)
  if err != nil { return res, err }
  content := string(src)
  if strings.Contains(content, "Code generated") && strings.Contains(content, "DO NOT EDIT") {
    res.IsGenerated = true
  }
  if strings.Contains(content, "//go:embed") {
    res.HasGoEmbed = true
  }

  fset := token.NewFileSet()
  f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
  if err != nil { return res, err }

  // Walk top-level declarations
  for _, d := range f.Decls {
    switch decl := d.(type) {
    case *ast.FuncDecl:
      if decl.Recv != nil && len(decl.Recv.List) > 0 {
        res.MethodRecvNames = append(res.MethodRecvNames, recvTypeName(decl.Recv.List[0].Type))
      } else {
        res.NonMethodFuncNames = append(res.NonMethodFuncNames, decl.Name.Name)
      }
    case *ast.GenDecl:
      switch decl.Tok {
      case token.TYPE:
        for _, s := range decl.Specs {
          if ts, ok := s.(*ast.TypeSpec); ok {
            res.TypeNames = append(res.TypeNames, ts.Name.Name)
          }
        }
      case token.CONST:
        res.ConstBlocks++
        for _, s := range decl.Specs {
          if vs, ok := s.(*ast.ValueSpec); ok && vs.Type != nil {
            if tn := typeExprName(vs.Type); tn != "" { res.ConstTypedNames = append(res.ConstTypedNames, tn) }
          }
        }
      default:
        // ignore imports/var
      }
    }
  }
  return res, nil
}

func checkSingleDecl(mode string, r Result) bool {
  switch mode {
  case "strict":
    // Exactly one of: any func (methods count individually), type spec, const block
    entities := len(r.TypeNames) + len(r.MethodRecvNames) + len(r.NonMethodFuncNames) + r.ConstBlocks
    return entities == 1
  default: // cohesive
    // Case 1: has exactly one type, allow methods for that type and consts typed as that type.
    if len(r.TypeNames) == 1 {
      t := r.TypeNames[0]
      // no non-method functions
      if len(r.NonMethodFuncNames) > 0 { return false }
      // all methods receivers match type
      for _, rt := range r.MethodRecvNames {
        if baseType(rt) != t { return false }
      }
      // if consts exist and are typed, they must match the type
      for _, ct := range r.ConstTypedNames {
        if baseType(ct) != t { return false }
      }
      // multiple types => not allowed
      return true // one cohesive unit
    }
    // Case 2: only methods, all receivers same type
    if len(r.TypeNames) == 0 && len(r.MethodRecvNames) > 0 && len(r.NonMethodFuncNames) == 0 {
      // ensure all receivers same base type
      uniq := map[string]struct{}{}
      for _, rt := range r.MethodRecvNames { uniq[baseType(rt)] = struct{}{} }
      return len(uniq) == 1 && len(r.ConstTypedNames) == 0
    }
    // Case 3: single non-method function only
    if len(r.TypeNames) == 0 && len(r.MethodRecvNames) == 0 && len(r.NonMethodFuncNames) == 1 && r.ConstBlocks == 0 {
      return true
    }
    // Case 4: only const blocks (typed or untyped)
    if len(r.TypeNames) == 0 && len(r.MethodRecvNames) == 0 && len(r.NonMethodFuncNames) == 0 && r.ConstBlocks >= 1 {
      // treat as one enum unit
      // if there are typed consts, ensure all typed const specs (across blocks) share one type
      if len(r.ConstTypedNames) <= 1 { return true }
      uniq := map[string]struct{}{}
      for _, ct := range r.ConstTypedNames { uniq[baseType(ct)] = struct{}{} }
      return len(uniq) == 1
    }
    return false
  }
}

func requiresTest(r Result) bool {
  // Skip for docs/generated/embed-only files
  if r.IsDoc || r.IsGenerated || r.HasGoEmbed { return false }
  if r.IsTest { return false }
  // Require tests when the file defines a type or any function (method or not)
  if len(r.TypeNames) > 0 || len(r.MethodRecvNames) > 0 || len(r.NonMethodFuncNames) > 0 {
    return true
  }
  return false
}

func reportSingleDeclViolation(r Result) {
  var parts []string
  if n := len(r.TypeNames); n > 0 { parts = append(parts, fmt.Sprintf("types=%d[%s]", n, strings.Join(r.TypeNames, ","))) }
  if n := len(r.MethodRecvNames); n > 0 { parts = append(parts, fmt.Sprintf("methods=%d[%s]", n, strings.Join(r.MethodRecvNames, ","))) }
  if n := len(r.NonMethodFuncNames); n > 0 { parts = append(parts, fmt.Sprintf("funcs=%d[%s]", n, strings.Join(r.NonMethodFuncNames, ","))) }
  if r.ConstBlocks > 0 { parts = append(parts, fmt.Sprintf("const-blocks=%d", r.ConstBlocks)) }
  if len(parts) == 0 { parts = append(parts, "no-declarations") }
  fmt.Printf("%s: expected single cohesive declaration; found %s\n", r.Path, strings.Join(parts, ","))
}

func recvTypeName(expr ast.Expr) string {
  switch t := expr.(type) {
  case *ast.StarExpr:
    return recvTypeName(t.X)
  case *ast.Ident:
    return t.Name
  case *ast.IndexExpr:
    return recvTypeName(t.X)
  case *ast.IndexListExpr:
    return recvTypeName(t.X)
  case *ast.SelectorExpr:
    return t.Sel.Name
  default:
    return fmt.Sprintf("%T", expr)
  }
}

func typeExprName(expr ast.Expr) string {
  switch t := expr.(type) {
  case *ast.Ident:
    return t.Name
  case *ast.SelectorExpr:
    return t.Sel.Name
  case *ast.StarExpr:
    return typeExprName(t.X)
  case *ast.ArrayType:
    return typeExprName(t.Elt)
  default:
    return ""
  }
}

func baseType(s string) string {
  // strip pointer/generic markers if any leaked in text
  return strings.TrimLeft(s, "*[]")
}

// no extra error helpers needed
EOF

# Run analyzer
ARGS=("-mode=${CHECK_MODE}" "-test=${CHECK_TEST_MODE}")
[[ "${CHECK_VERBOSE}" == "1" ]] && ARGS+=("-v")

if ! printf '%s
' "${GO_FILES[@]}" | go run "${TMPDIR}/declcheck.go" "${ARGS[@]}"; then
  exit 1
fi

[[ "${CHECK_VERBOSE}" == "1" ]] && echo "All files pass single-declaration and test checks."
exit 0
