#!/usr/bin/env bash
set -euo pipefail

root="std/ami/stdlib"
if [[ ! -d "${root}" ]]; then
  echo "No stdlib directory at ${root}; skipping .ami-only check."
  exit 0
fi

# Find any regular files under stdlib that are not .ami
bad_files=$(find "${root}" -type f ! -name "*.ami" -print || true)
if [[ -n "${bad_files}" ]]; then
  echo "Non-.ami files found under ${root}:" >&2
  echo "${bad_files}" | sed 's/^/ - /' >&2
  exit 1
fi

echo "AMI stdlib check passed: only .ami files under ${root}."
exit 0

