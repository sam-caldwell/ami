#!/usr/bin/env bash
set -euo pipefail

echo "Scanning src/ for test coverage hotspots..." >&2
no_tests=0

# Enforce: packages with no *_test.go (exclude test fixtures)
find src -type d | grep -v "^src/cmd/ami/build/test/" | while read -r d; do
  c_go=$(find "$d" -maxdepth 1 -type f -name '*.go' | wc -l | tr -d ' ')
  c_test=$(find "$d" -maxdepth 1 -type f -name '*_test.go' | wc -l | tr -d ' ')
  if [ "$c_go" != "0" ] && [ "$c_test" = "0" ]; then
    echo "NO_TESTS  $d"
    no_tests=1
  fi
done

# Advisory: files with no paired *_test.go (ignore embeds, docs, fixtures)
find src -type f -name "*.go" ! -name "*_test.go" | while read -r f; do
  case "$f" in
    *schema_embed.go|*doc.go|src/cmd/ami/build/test/*) continue;;
  esac
  base=$(basename "$f" .go)
  dir=$(dirname "$f")
  if [ ! -f "$dir/${base}_test.go" ]; then
    echo "MISSING_PAIR  $f  (expect: ${dir}/${base}_test.go)"
  fi
done

if [ "$no_tests" = "1" ]; then
  echo "One or more packages have no tests." >&2
  exit 1
fi
exit 0
