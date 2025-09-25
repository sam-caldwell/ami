# ami lint

Lints AMI source files declared in the current workspace. It discovers `*.ami` units per package, orders linting so that workspace‑local imports are checked before the `main` package, and emits diagnostics for common issues.

## Usage

- Human: `ami lint`
- JSON: `ami --json lint`

## Output

- Human: prints discovered units and a readable list of diagnostics, then a summary line: `lint: summary: <warns> warnings, <errs> errors`.
- JSON: emits a `sources.v1` record first, then one `diag.v1` per diagnostic, followed by a final `diag.v1` summary with code `LINT_SUMMARY`.

## Rules (initial set)

Default severities unless overridden by config/strict:

- Package naming:
  - `W_PKG_LOWERCASE` (warn): package name should be lowercase.

- Imports:
  - `W_DUP_IMPORT` (warn): duplicate import of the same path.
  - `W_DUP_IMPORT_ALIAS` (warn): duplicate alias used for multiple imports.
  - `W_UNUSED_IMPORT` (warn): imported alias not referenced.
  - `W_IMPORT_ORDER` (warn): imports not ordered lexicographically by path.
  - `W_REL_IMPORT` (warn): relative import paths (`./...`) disallowed; use workspace package names.

- File formatting:
  - `W_FILE_NO_NEWLINE` (warn): file does not end with newline.
  - `W_FILE_CRLF` (warn): CRLF line endings detected; use LF.

- Hygiene:
  - `W_TODO_COMMENT` (warn): TODO marker found in comment.
  - `W_FIXME_COMMENT` (warn): FIXME marker found in comment.
  - `W_BLANK_IDENT_USAGE` (warn/off by default): '_' identifier outside allowed sinks. Disabled by default to avoid noise; enable via config.

 - Pipeline hints:
  - `W_PIPELINE_INGRESS_POS` (info): ingress position hint when present.
  - `W_PIPELINE_EGRESS_POS` (info): egress position hint when present.

 - Language reminders:

 - RAII usage hints:
  - `W_RAII_OWNED_HINT` (info): Owned<T> in signatures/structs should be released using `mutate(release(x))` or an equivalent explicit release call.

- Collection constraints (mirrors as hints):
  - `W_MAP_ARITY_HINT` (warn): map requires two type parameters (`map<K,V>`).
  - `W_MAP_KEY_TYPE_HINT` (warn): map key should be scalar; slices or maps are not allowed.
  - `W_SET_ARITY_HINT` (warn): set requires one type parameter (`set<T>`).
  - `W_SET_ELEM_TYPE_HINT` (warn): set element should be scalar; slices or maps are not allowed.
  - `W_SLICE_ARITY_HINT` (warn): slice requires one type parameter (`slice<T>`).

- Workspace & cross‑package:
  - `E_WS_PKG_NAME` (error): workspace package name invalid.
  - `E_WS_PKG_VERSION` (error): workspace package version not valid SemVer.
  - `E_IMPORT_CONSTRAINT` (error): imported local package version doesn’t satisfy importer’s constraint.
  - `E_IMPORT_PRERELEASE_FORBIDDEN` (error): import resolves to a prerelease while constraint omits prerelease.
  - `E_IMPORT_CONSISTENCY` (error): conflicting constraints for the same package across importers.
  - Workspace schema/validation errors are surfaced as `diag.v1` with code `E_WS_SCHEMA` in JSON mode and as plain text in human mode.

## Configuration

- In `ami.workspace`, set per‑rule severities under `toolchain.linter.rules`:

  toolchain:
    linter:
      rules:
        W_PKG_LOWERCASE: error
        W_FILE_NO_NEWLINE: off

  - Allowed severities: `error`, `warn`, `info`, or `off` to disable the rule.
  - Applies to all files; inline pragmas can further suppress per‑file.

- Strict preset via workspace: set `toolchain.linter.strict: true` to elevate warnings to errors (same as `--strict`).

- Suppress rules via workspace:

  toolchain:
    linter:
      suppress:
        package:
          main: ["W_UNUSED_IMPORT"]
        paths:
          ./src/**: ["W_PKG_LOWERCASE"]

  - `package`: map of package name to list of rule codes to suppress.
  - `paths`: map of file or directory globs to list of rule codes. Use `/**` suffix to match directories recursively. Use `"*"` to suppress all rules for a scope.

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
- JSON namespace note: `diag.v1.code` currently uses existing rule names (e.g., `W_*`, `E_*`). A forward‑compat alias is also provided in `diag.v1.data.lint_code` prefixed with `LINT_` (e.g., `LINT_W_FILE_CRLF`). Once the migration is complete, `code` may switch to `LINT_*`; use `data.lint_code` for robust tooling.
- Future extensions continue in SPECIFICATION.md.
