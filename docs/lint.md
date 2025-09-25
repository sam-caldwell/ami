# ami lint

Lints AMI source files declared in the current workspace. It discovers `*.ami` units per package, orders linting so that workspace‑local imports are checked before the `main` package, and emits diagnostics for common issues.

## Usage

- Human: `ami lint`
- JSON: `ami --json lint`

## Output

- Human: prints discovered units and a readable list of diagnostics, then a summary line: `lint: summary: <warns> warnings, <errs> errors`.
- JSON: emits a `sources.v1` record first, then one `diag.v1` per diagnostic, followed by a final `diag.v1` summary with code `LINT_SUMMARY`.

## Rules (initial set)

- Package naming: `W_PKG_LOWERCASE` when the declared package name is not lowercase.
- Imports:
  - `W_DUP_IMPORT` for duplicate imports of the same path.
  - `W_UNUSED_IMPORT` when an import alias is not referenced.
- File formatting:
  - `W_FILE_NO_NEWLINE` when the file does not end with a newline (LF).
  - `W_FILE_CRLF` when CRLF line endings are detected; prefer LF.
- Workspace checks:
  - `E_WS_PKG_NAME` when a workspace package name is invalid.
  - `E_WS_PKG_VERSION` when a workspace package version is not a valid SemVer.
  - Workspace schema/validation errors are surfaced as `diag.v1` with code `E_WS_SCHEMA` in JSON mode and as a plain message in human mode.

## Configuration

- In `ami.workspace`, set per‑rule severities under `toolchain.linter.rules`:

  toolchain:
    linter:
      rules:
        W_PKG_LOWERCASE: error
        W_FILE_NO_NEWLINE: off

  - Allowed severities: `error`, `warn`, `info`, or `off` to disable the rule.
  - Applies to all files; inline pragmas can further suppress per‑file.

## Pragmas

- File‑level suppression toggles using `#pragma`:
  - `#pragma lint:disable RULE[,RULE2]` disables listed rules for the file.
  - `#pragma lint:enable RULE[,RULE2]` re‑enables listed rules.
  - Current scope is file‑wide (no block scoping yet).

## CLI Flags

- `--strict`: treat warnings as errors for reporting. In JSON output, warnings are emitted as `level:"error"`. The process exit code does not change (lint remains informative); builds/tests still govern exit behavior.
- `--rules=<pattern[,pattern...]>`: select which rules to include in output. Multiple patterns are comma‑separated. Supported forms:
  - Substring: case‑insensitive match. Example: `--rules=import` matches `W_UNUSED_IMPORT` and `W_DUP_IMPORT`.
  - Glob: `*`, `?`, `[]` wildcards, matched case‑insensitively. Example: `--rules=W_FILE_*` selects file formatting warnings.
  - Regex: prefix with `re:` or wrap with slashes `/.../`. Uses Go regex (case‑sensitive unless you add `(?i)`). Examples: `--rules=re:^W_FILE_` or `--rules=/^W_FILE_/`.
- `--max-warn=<n>`: maximum number of warnings to emit (0 = unlimited). Additional warnings are suppressed; the final summary still reflects the counts emitted.

Examples

- Show only file formatting warnings (human):
  - `ami --rules W_FILE_* lint`
- Show only file formatting warnings (JSON via regex):
  - `ami --json --rules re:^W_FILE_ lint`
- Strict mode with at most two warnings:
  - `ami --strict --max-warn 2 lint`

## Notes

- Linting order ensures workspace‑local packages are analyzed before `main` to improve signal for dependency issues.
- Future extensions will add configurable severities, suppressions, richer positions, and cross‑package checks as tracked in SPECIFICATION.md.
