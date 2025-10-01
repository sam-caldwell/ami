#!/usr/bin/env bash
set -euo pipefail

# Coverage gate for changed packages.
# - Determines changed Go packages between BASE..HEAD and enforces per-package coverage >= THRESHOLD.
# - Fails if any changed package has no *_test.go files.
#
# Environment:
#   THRESHOLD   - required coverage as a fraction (default: 0.80)
#   BASE_SHA    - explicit base commit SHA (optional)
#   BASE_REF    - explicit base ref/branch (e.g., origin/main) if BASE_SHA unset
#   HEAD_SHA    - explicit head commit SHA (defaults to HEAD)
#   SKIP_PKGS   - regex; skip packages whose import path matches this pattern (optional)
#   PKG_THRESHOLDS - comma-separated regex=threshold overrides (e.g., "pkg/unstable/.*=0.70,.*parser.*=0.90")
#   EXTRA_EXCLUDES - space/comma-separated path patterns to exclude from changed files (e.g., "src/**/third_party/**")
#
# Usage:
#   scripts/coverage_gate.sh            # auto-detect base
#   THRESHOLD=0.85 scripts/coverage_gate.sh

threshold=${THRESHOLD:-0.80}
head_sha=${HEAD_SHA:-HEAD}

# Resolve base to compare against
base_sha=${BASE_SHA:-}
if [[ -z "${base_sha}" ]]; then
  base_ref=${BASE_REF:-}
  if [[ -n "${base_ref}" ]]; then
    # If it's a bare branch name, prefix with origin/
    if [[ "${base_ref}" != origin/* ]]; then base_ref="origin/${base_ref}"; fi
    git fetch --no-tags --prune origin "+refs/heads/*:refs/remotes/origin/*" >/dev/null 2>&1 || true
    base_sha=$(git rev-parse --verify "${base_ref}" 2>/dev/null || echo "")
  fi
fi
if [[ -z "${base_sha}" ]]; then
  # Fallback: try origin/main
  git fetch --no-tags --prune origin "+refs/heads/*:refs/remotes/origin/*" >/dev/null 2>&1 || true
  if git rev-parse --verify origin/main >/dev/null 2>&1; then
    base_sha=$(git rev-parse origin/main)
  else
    # Last resort: previous commit
    base_sha=$(git rev-parse HEAD~1)
  fi
fi

echo "Coverage gate threshold: ${threshold}"
echo "Comparing changes: ${base_sha}..${head_sha}"

# Collect changed Go package directories under src/, respecting excludes
read -r -a excludes <<< ""
excludes=(
  ":!**/*.pb.go"
  ":!tools/**"
  ":!examples/**"
  ":!tests/**"
  ":!src/**/testdata/**"
  ":!src/**/generated/**"
  ":!src/**/mocks/**"
  ":!**/zz_generated*.go"
  ":!**/*_mock.go"
)
if [[ -n "${EXTRA_EXCLUDES:-}" ]]; then
  # split on spaces or commas
  IFS=', ' read -r -a extra <<< "${EXTRA_EXCLUDES}"
  for p in "${extra[@]}"; do
    [[ -z "${p}" ]] && continue
    excludes+=(":!${p}")
  done
fi

changed_dirs=$(git diff --name-only "${base_sha}..${head_sha}" -- 'src/**/*.go' "${excludes[@]}" |
  xargs -n1 dirname | sort -u)

if [[ -z "${changed_dirs}" ]]; then
  echo "No changed Go packages; skipping coverage gate."
  exit 0
fi

fail=0
echo "Changed package directories:" >&2
for d in ${changed_dirs}; do echo "  - ${d}" >&2; done

tmpdir=$(mktemp -d)
trap 'rm -rf "${tmpdir}"' EXIT

for dir in ${changed_dirs}; do
  # Convert directory to Go package import path
  if ! pkg=$(go list "./${dir}" 2>/dev/null); then
    echo "SKIP: not a Go package: ${dir}" >&2
    continue
  fi
  # Package skip by regex
  if [[ -n "${SKIP_PKGS:-}" ]]; then
    if [[ "${pkg}" =~ ${SKIP_PKGS} ]]; then
      echo "SKIP: ${pkg} matched SKIP_PKGS pattern" >&2
      continue
    fi
  fi
  # Enforce presence of tests
  shopt -s nullglob
  test_files=("${dir}"/*_test.go)
  shopt -u nullglob
  if [[ ${#test_files[@]} -eq 0 ]]; then
    echo "FAIL: ${pkg} has no *_test.go files; add tests to satisfy CC-1." >&2
    fail=1
    continue
  fi
  prof="${tmpdir}/$(echo "${pkg}" | tr '/.' '__').out"
  # Run package tests with coverage
  if ! go test -count=1 -covermode=count -coverprofile "${prof}" "${pkg}" >/dev/null; then
    echo "FAIL: tests failed for ${pkg}" >&2
    fail=1
    continue
  fi
  # Parse coverage total
  total_line=$(go tool cover -func "${prof}" | tail -n1)
  cov_pct=$(echo "${total_line}" | awk '{print $NF}' | tr -d '%')
  cov_frac=$(awk -v n="${cov_pct}" 'BEGIN { printf "%.4f", n/100.0 }')
  printf "COVER: %s total=%.2f%%\n" "${pkg}" "${cov_pct}" >&2
  # Per-package threshold overrides
  pkg_threshold="${threshold}"
  if [[ -n "${PKG_THRESHOLDS:-}" ]]; then
    IFS=',' read -r -a rules <<< "${PKG_THRESHOLDS}"
    for rule in "${rules[@]}"; do
      [[ -z "${rule}" ]] && continue
      pat=${rule%=*}
      val=${rule#*=}
      if [[ -n "${pat}" && -n "${val}" ]] && [[ "${pkg}" =~ ${pat} ]]; then
        pkg_threshold="${val}"
        break
      fi
    done
  fi
  awk -v c="${cov_frac}" -v t="${pkg_threshold}" 'BEGIN { if (c+0.00001 < t) exit 1; else exit 0 }' || {
    echo "FAIL: ${pkg} coverage ${cov_frac} below threshold ${pkg_threshold}" >&2
    fail=1
  }
done

if [[ ${fail} -ne 0 ]]; then
  echo "Coverage gate failed." >&2
  exit 1
fi

echo "Coverage gate passed for changed packages."
exit 0
