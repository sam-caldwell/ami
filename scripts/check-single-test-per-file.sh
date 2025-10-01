#!/usr/bin/env bash
set -euo pipefail

# Usage: check-single-test-per-file.sh [dir]
# Scans Go test files under the given directory (default: src/ami/compiler/parser)
# and fails if any *_test.go file contains more than one Test* function.

DIR=${1:-src}

if [ ! -d "$DIR" ]; then
  echo "ERROR: directory not found: $DIR" >&2
  exit 2
fi

errors=0
while IFS= read -r -d '' f; do
  # Count public test functions in the file (functions starting with Test), excluding TestMain
  count=$(grep -E '^[[:space:]]*func[[:space:]]+Test[[:alnum:]_]*[[:space:]]*\(' "$f" | grep -v -E '^[[:space:]]*func[[:space:]]+TestMain[[:space:]]*\(' | wc -l | tr -d ' ')
  if [ "$count" -gt 1 ]; then
    echo "ERROR: $f contains $count Test functions (expected at most 1)" >&2
    errors=1
  fi
done < <(find "$DIR" -type f -name '*_test.go' -print0 | sort -z) || true

exit "$errors"
