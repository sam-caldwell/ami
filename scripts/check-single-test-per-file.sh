#!/usr/bin/env bash
set -euo pipefail

# Thin wrapper: run the Go implementation.
exec go run "${BASH_SOURCE%/*}/check-single-test-per-file.go" "$@"
