#!/usr/bin/env bash
set -euo pipefail

# Usage: check-single-test-per-file.sh [dir]
# Scans Go test files under the given directory (default: src/)
# and fails if any *_test.go file contains more than one Test* function.

DIR=${1:-src/}

if [ ! -d "$DIR" ]; then
  echo "ERROR: directory not found: $DIR" >&2
  exit 2
fi

errors=0

# Build a tiny analyzer to count top-level Test* functions via go/parser (ignores strings)
TMPDIR=$(mktemp -d 2>/dev/null || mktemp -d -t amitestcheck)
cleanup(){ rm -rf "$TMPDIR"; }
trap cleanup EXIT
cat >"$TMPDIR"/testcheck.go <<'EOF'
package main

import (
  "flag"
  "fmt"
  "go/ast"
  "go/parser"
  "go/token"
  "os"
)

func main(){
  flag.Parse()
  if flag.NArg() == 0 { return }
  fset := token.NewFileSet()
  for _, path := range flag.Args() {
    src, err := os.ReadFile(path)
    if err != nil { fmt.Fprintf(os.Stderr, "read %s: %v\n", path, err); os.Exit(1) }
    f, err := parser.ParseFile(fset, path, src, 0)
    if err != nil { fmt.Fprintf(os.Stderr, "parse %s: %v\n", path, err); os.Exit(1) }
    var count int
    for _, d := range f.Decls {
      if fn, ok := d.(*ast.FuncDecl); ok {
        if fn.Recv == nil && len(fn.Name.Name) >= 4 && fn.Name.Name[:4] == "Test" && fn.Name.Name != "TestMain" {
          count++
        }
      }
    }
    if count > 1 {
      fmt.Printf("%s: %d Test functions (expected at most 1)\n", path, count)
      os.Exit(2)
    }
  }
}
EOF

while IFS= read -r -d '' f; do
  if ! go run "$TMPDIR"/testcheck.go -- "$f" >/dev/null 2>&1; then
    echo "ERROR: $f contains more than one Test* function" >&2
    errors=1
  fi
done < <(find "$DIR" -type f -name '*_test.go' -print0 | sort -z) || true

exit "$errors"
