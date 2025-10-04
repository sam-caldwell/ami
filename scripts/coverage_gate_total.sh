#!/usr/bin/env bash
set -euo pipefail

# Enforce an overall coverage threshold across ./src/... packages.
# Usage:
#   scripts/coverage_gate_total.sh
#   THRESHOLD=0.85 scripts/coverage_gate_total.sh

threshold=${THRESHOLD:-0.80}

mkdir -p build
profile="build/coverage.out"
echo "Running coverage across ./src/... (threshold=${threshold})"
GOFLAGS= go test -count=1 -covermode=atomic -coverprofile="${profile}" ./src/...

line=$(go tool cover -func="${profile}" | tail -n 1)
pct=$(echo "${line}" | awk '{print $NF}' | tr -d '%')
frac=$(awk -v n="${pct}" 'BEGIN { printf "%.4f", n/100.0 }')
echo "TOTAL COVERAGE: ${pct}%"

awk -v c="${frac}" -v t="${threshold}" 'BEGIN { if (c+0.00001 < t) exit 1; else exit 0 }' || {
  echo "FAIL: total coverage ${frac} below threshold ${threshold}" >&2
  exit 1
}

echo "Coverage gate (total) passed."
exit 0

