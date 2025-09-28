# AMI Toolchain Specification (Chapter 3.0)
---
# General
## Authoritative Source
- docs/roadmap.md is the general guide to execute this plan
- Source of truth: `docs/Asynchronous Machine Interface.docx`
    - defines POP paradigm
    - defines AMI language syntax and semantics
    - defines toolchain program (ami)
- AGENTS.MD defines repo formatting and code styling guidelines
- This SPECIFICATION.md documents the 
  - implementation plan, 
  - constraints, and 
  - progress tracking for this repository.
- When conflicts arise, the `docs/Asynchronous Machine Interface.docx` governs.
> Goal: Implement a deterministic, scriptable CLI `ami` for workspace management, packages, linting, testing, and
>       building. All commands are non‑interactive and support machine‑parsable output via `--json`. Exit
>       codes and global flags are standardized.
## Architecture 
  - Language: Go 1.25+
  - Binary: `ami` (single CLI with subcommands)
  - CLI framework: `spf13/cobra` for root/subcommand structure and flag parsing
  - Workspace file: `ami.workspace` at repo root
  - Package cache: `${HOME}/.ami/pkg`
  - Build output: `./build`
  - Output formats: Human (default) and JSON (`--json`)
  - Exit codes: SUCCESS(0), USER_ERROR(1), SYSTEM_IO_ERROR(2), INTEGRITY_VIOLATION_ERROR(3), NETWORK_REGISTRY_ERROR(4)
## Language scope
- AMI is a Pipeline‑Oriented Programming (POP) language.
- Chapter 1 establishes semantics for events, nodes, edges, error pipelines, concurrency and observability. 
- Chapter 2 defines the AMI language needed to author those pipelines and companion imperative code (types, functions,
  state, memory). This specification tracks the concrete work items required to implement those language features in 
  the compiler and toolchain.
## I/O Streams and Formatting
- Human output goes to stdout; errors/diagnostics go to stderr.
- JSON output (when `--json`) is written to stdout; fatal errors still mirror a plain‑text summary to stderr.
- For streaming commands (e.g., `test`), JSON should be JSON Lines (one object per line) for events, plus a final summary object. Human output remains grouped and readable.
- Color:
  - Default is disabled unless `--color` is set. When `--json` is set, color must be disabled.
  - If auto‑detection is later added, it must never enable color when `--json` is present.
## Global Flags (all subcommands)

- `--help`: print help and exit
- `--json`: print machine‑parsable JSON (mutually exclusive with `--color`)
- `--verbose`: write verbose output
- `--color`: enable ANSI colors (human output only; mutually exclusive with `--json`)
- [X] `--redact`: redact exact field keys in debug logs
- [X] `--redact-prefix`: redact field keys by prefix in debug logs
- [X] `--allow-field`: allowlist of field keys to include in debug logs
- [X] `--deny-field`: denylist of field keys to exclude from debug logs

### Flag Interactions

- `--json` and `--color` cannot be used together.
  - Validation occurs during flag parsing; the process exits with `USER_ERROR (1)`.
  - Error message is written as plain text to stderr (not JSON, not colored) because rendering mode is undefined.
  - Tests cover: `--json --color` on root and subcommands, ensuring exit code 1 and a clear message.
## Coding Style Guidelines
- All builds should be in build/ or build/<time>/
  - DO NOT PUT logs, build artifacts or temp files in the repo root.
- Keep Go source files minimal: limit each file to a single top-level declaration—either one function, one type (including a struct), one const block, or one var block.
- For every .go file there will be a corresponding _test.go file containing at least one happy path and one sad path test.
- Use docstrings to describe what each top-level declaration is and what it does in clear language.
## Deliverables Checklist
- [X] Binary: `ami` with subcommand scaffold and flag plumbing
  - [x] Global flag handling (`--help`, `--json`, `--verbose`, `--color`)
  - [x] Logger with human/JSON renderers and verbosity levels
  - [x] Workspace loader/parser for `ami.workspace`
- [X] Manifest library `src/ami/manifest` with `Load()` and `Save()` for `ami.manifest`
  - [X] Package cache manager at `${HOME}/.ami/pkg`
- [X] Subcommands implemented (init, clean, mod clean/update/get/list, lint, test, build)
    - [x] `ami init` is completely implemented with >=80% test coverage and all tests passing.
    - [x] `ami clean` is completely implemented; tests passing; coverage in `cmd/ami` currently ~78% (≥75% minimum). Follow-up to raise ≥80%.
    - [x] CLI stubs registered: `build`, `test`, `lint`, `mod {get}`; inert and help-only for now.
    - [x] `ami mod clean` is completely implemented with >=80% test coverage and all tests passing.
    - [X] `ami mod update` is completely implemented with >=80% test coverage and all tests passing.
    - [X] `ami mod get` is completely implemented with >=80% test coverage and all tests passing.
    - [x] `ami mod list` implemented: lists cached packages with name, version, size, updated; JSON/human output; tests passing.
    - [x] `ami mod audit` implemented: audits workspace imports vs `ami.sum` and cache; JSON/human outputs; unit + e2e tests passing.
    - [x] `ami mod sum` enhanced: validates presence, JSON/scheme; verifies directory hashes against `${AMI_PACKAGE_CACHE}`; reports missing/mismatched; returns exit.Integrity on failure. Tests passing.
    - [x] `ami lint` Stage A implemented with >=80% coverage and tests passing (workspace presence, name style, import shape/order, local path checks, UNKNOWN_IDENT scan, strict mode, verbose debug file).
    - [X] `ami lint` Stage B (parser-backed rules): memory safety (`E_PTR_UNSUPPORTED_SYNTAX`, `E_MUT_BLOCK_UNSUPPORTED`), unmarked assignment (`E_MUT_ASSIGN_UNMARKED`), RAII hint (`W_RAII_OWNED_HINT`), duplicate import aliases and function decls, unused imports, pipeline position hints (ingress first/egress last), reachability, buffer policy smells (`drop` alias and tiny capacity with drop policies). Tests passing with ≥80% coverage in `cmd/ami`.
    - [X] `ami pipeline visualize` implemented: renders ASCII pipeline graphs to the terminal; JSON/human output; unit + e2e tests.
    - [X] `ami test` implemented:
      - [X] Go test wrapper that collects `_test.go`, streams `go test -json` events, prints human "test: OK" on success, and emits a final JSON summary in `--json` mode.
      - [X] `--verbose` writes `build/test/test.log` and `build/test/test.manifest` with `<package> <test>` entries in run order.
      - [X] Native AMI directive‑based assertions (parser/sem) integrated into harness: support `#pragma test:case <name>` and `#pragma test:assert parse_ok|parse_fail` with manifest entries and failures affecting exit code.
      - [X] Package‑level concurrency flag (`--packages`) and explicit per‑package summaries in human mode.
      - Notes: runtime execution of AMI code deferred; current scope wraps Go tests per roadmap M12.
    - [ ] `ami build` is completely implemented with >=80% test coverage and all tests passing.
  - [X] Deterministic behaviors (no prompts, stable outputs)
  - [X] CLI/toolchain tests run from `./build/test/` (per-test subdirs)
  - [x] Code quality guarantee
    - [x] `go vet ./...` and `go test -v ./...` pass
    - [x] Unit and integration tests (>=80% coverage target; ≥75% minimum met for changed packages)
    - [x] Build: `go build -o build/ami ./src/cmd/ami`
    - [x] Repository structure and testing conventions from AGENTS.md are enforced (source under `src/`, one type/function per file where practical, `_test.go` colocated, happy/sad path tests, ≥75% coverage with 80% target)
  - [ ] AMI compiler architecture
  - [X] `compiler/sem`: decomposed into modular files mirroring `scanner` pattern (one concept per file)
    - [X] `compiler/source`: decomposed into modular files (`position.go`, `file.go`, `fileset.go`) with tests split per concept
  - [X] `compiler/types`: verified modular split by concept; added concise docs and unit tests for composites, basics, and function rendering
    - [ ] `manifest`: decomposed into `types.go`, `validate.go`, `sumcheck.go`, and `io.go`; existing tests left intact and passing
- [X] Docs for user‑visible commands under `docs/` updated alongside features
- [X] `ami.sum` JSON summary file with package→version→sha256 mapping; 
  - [x] updated by `mod get/update`, verified by `build`
- [X] Examples. provide:
  - `examples/simple` and 
  - `examples/complex` 
  - workspaces with README; 
- [X] Makefile targets
  - [X] `make clean`
    - delete `build` directory and recreate it.
  - [X] `make lint`
    - Lint the entire repo without errors
  - [X] `make test`
    - Run all tests for the repo without errors
  - [X] `make build`
    - build `ami` binary without errors
  - [X] `make examples`
     - `examples` target stages builds under `build/examples/**`
## Remaining Work

- [X] CLI: scaffold and register additional subcommands as stubs in root (`mod list/sum/get`, `test`, `build`); add minimal help/tests; keep root stable.
- [X] Linter: expand Stage A to handle imports/naming/unknown identifiers; add JSON Lines streaming and final summary per SPEC.
- Tests: stabilize Cobra working-directory integration tests for root→subcommand invocations and unskip the pending tests once behavior is deterministic.
- Coverage: raise `src/cmd/ami` package test coverage to ≥80% (currently ~78–79%).
- [X] Scaffold src/ami/compiler/{token,scanner,parser,ast,source} with minimal types and tests (Phase 2 starter).
- [X] Add a test that checks rule mapping elevation (e.g., set W_IMPORT_ORDER=error makes non-zero exit in JSON mode).
- [ ] Stdlib logger pipeline: expose redaction/filters via CLI when pipeline mode replaces current logger. Wire `ami/stdlib/logger` pipeline config (JSONRedactKeys/Prefixes; future allow/deny) and add tests for batch/interval/backpressure, counters, and safety‑net redaction of `log.v1` lines.
---
# Details
## 1.0.0.0. Features and Work Breakdown
### 1.0.0.1 Workspace File Schema (`ami.workspace`)
- AMI projects are built around a YAML manifest called `ami.workspace`.
- [X] Create `docs/Workspace/README.md` to document the file format and schema, including the following required fields:
    - [X] `project`: `{ name, version }` (version follows SemVer)
    - [X] `packages`: list of import paths with version constraints (e.g., `^1.2`, `~1.2.3`, exact `1.2.3`)
        - [X] Accepted forms: exact `X.Y.Z` (with optional leading `v`), `^X.Y.Z`, `~X.Y.Z`, `>X.Y.Z`, `>=X.Y.Z`, and macro `==latest`.
        - [X] SemVer must be `MAJOR.MINOR.PATCH`; prereleases allowed when specified (e.g., `^1.0.0-rc.1`).
        - [X] Whitespace inside constraints is ignored (e.g., `>= 1.2.3` is accepted).
        - [X] Unsupported operators (e.g., `<=`) are rejected.
    - [X] Exact `toolchain.*` keys per Chapter 3.0 examples:
        - [X] `toolchain.compiler.concurrency`: integer ≥1 or the macro `NUM_CPU` (string). If `NUM_CPU`, detect host CPU count at runtime.
        - [X] `toolchain.compiler.target`: workspace‑relative output directory path (default `./build`). Must not be absolute; must not traverse outside workspace (reject `..`).
        - [X] `toolchain.compiler.env`: list of `os/arch` pairs; duplicates eliminated and order preserved.
            - [X] `os/arch` must match pattern `^[A-Za-z0-9._-]+/[A-Za-z0-9._-]+$`.
            - [X] Known examples (valid): windows/amd64, linux/amd64, linux/arm64, darwin/amd64, darwin/arm64. Extensible.
            - [X] Duplicates are eliminated; ordering is preserved as declared.
        - [X] `ami init` seeds `toolchain.compiler.env` with the current host OS/arch pair (e.g., `darwin/arm64`).
        - [X] `toolchain.linker`: object (reserved for future keys). For now must be an object; unknown types rejected.
        - [X] `toolchain.linter`: object (reserved for future keys). For now must be an object; unknown types rejected.
    - [X] Top‑level `version`: schema version in SemVer format (e.g., `1.0.0`).
    - [X] On `ami build`, enforce `toolchain.*` constraints; violations → `USER_ERROR` with clear message (and JSON diagnostic when `--json`).
        - JSON diagnostics: emits `diag.v1` with `level:"error"`, `code:"E_WS_SCHEMA"`, `message:"workspace validation failed: …"`, `file:"ami.workspace"`.
    - [X] Tests: invalid/missing keys, version constraint violations, JSON diagnostics on build
### 1.0.0.2 Workspace File Package (`src/workspace/`)
- [x] Implement a `workspace.Workspace` struct which can serialize/deserialize `ami.workspace`
- [x] `Workspace` struct should have a Load(path string) method to load `ami.workspace` YAML and deserialize it 
      into `Workspace`.
- [x] `Workspace` struct should have a Save(path string) method to write `ami.workspace` YAML using its internal 
      state.
- [x] `Workspace` struct should have a Create(path string) method to create `ami.workspace` YAML with defaults 
      (see 1.0.0.1).
### 1.0.1.0. AMI Toolchain (`ami`) command-line (spf13/cobra)
- [x] Implement `src/cmd/ami/main.go` using Cobra root command `src/cmd/ami/root.go`
- [x] Implement the following subcommands under `src/cmd/ami/cmd/`:
  - [x] `init.go`      : see 1.1.0.1
  - [x] `clean.go`     : see 1.1.0.2
  - [x] `mod/clean.go` : see 1.1.0.3
  - [x] `mod/update.go`: see 1.1.0.4
  - [x] `mod/get.go`   : see 1.1.0.5
  - [x] `mod/list.go`  : see 1.1.0.6
  - [x] `lint.go`      : see 1.1.0.7
  - [X] `test.go`      : see 1.1.0.8
  - [X] `build.go`     : see 1.1.0.9
  - [x] `version.go`   : see 1.1.0.10
  - [x] `help.go`      : see 1.1.0.11
- [x] Implement `--help` via Cobra as an alias for `ami help` command
- [x] Implement persistent flags 
  - `--json`, 
  - `--verbose`,
  - `--color` on root; 
  - bind to a shared 'options' struct
- [x] Implement exit code constants and mapping from error types
  - exit codes and exit functions will be in the `exit` package
### 1.0.1.1. Structured Logging
- [x] Add a structured logger which logs to stdout (and to `build/debug` if `--verbose` is used.
- [x] Logging exists in a `logging` package.
- [X] All logs are JSON structured logs (even when --json is NOT used)
    - JSON renderer (stable schema: timestamp, level, module, msg, fields)
- [x] Human renderer (respects `--color`/`--verbose`)
- [x] Verbose timestamping (human):
- when `--verbose` is set, prefix every emitted line with an ISO‑8601 UTC timestamp
    - example: `{"t":"2025-09-24T17:05:06.123Z", "msg":"message"...}`
  - [x] Logs convert CRLF to LF
  - [x] Multi‑line messages must prefix every line unless wrapped as a multi-line string.
  - [x] Root logger wired from CLI flags; subcommands emit debug‑only messages when `--verbose`, writing to `build/debug/activity.log` without changing command stdout outputs.
  - [X] CLI exposes redaction/field filter controls for debug logs: `--redact`, `--redact-prefix`, `--allow-field`, `--deny-field`; wired to logger options; covered by CLI tests.
### 1.0.1.3. Basic CLI Testing
- [x] Tests:
    - Cobra command wiring,
    - flag parsing (persistent/local),
    - help text,
    - exit code paths,
    - JSON output schema
- [x] Validate flag interactions:
    - `--json` and `--color` are mutually exclusive
    - (exit `USER_ERROR` with plain‑text message)
- [x] Logger tests: with and without `--verbose` (human):
    - timestamp prefix presence/absence;
    - timestamp format is ISO‑8601 UTC with milliseconds;
    - multi‑line messages have per‑line prefix;
    - JSON output always contains a `timestamp` field.
- [x] Validate flag interactions:
    - `--json` and `--color` are mutually exclusive
    - (exit `USER_ERROR` with plain‑text message)
- [x] Module setup:
    - add `spf13/cobra` (and `spf13/pflag` via Cobra) to `go.mod`

### 1.0.1.4. CLI Benchmarks
- [X] Add microbenchmarks to measure core subcommand runtimes
  - Location: `src/cmd/ami/bench_subcommands_test.go`
  - Measures: `ami help`, `ami version`, `ami clean`, `ami lint`, `ami test`, `ami mod {update,list,sum,get,clean}`, `ami pipeline visualize`
  - Benchmarks run in isolated sandboxes with a temp workspace and cache (`AMI_PACKAGE_CACHE`), avoiding side effects.
  - Does not run by default; invoke with `go test -bench . ./src/cmd/ami`.
## 1.1.0.0. AMI Toolchain Command Details
### 1.1.1.0. Workspace Management (`ami init`)
- [X] `ami init` subcommand features complete:
    - [X] Create `ami.workspace` minimal schema if the file does not exist or if `--force` is used using the following
      template:
```yaml
---
version: 1.0.0
toolchain:
  compiler:
    concurrency: NUM_CPU
    target: ./build
    env:
      - darwin/arm64 # autodetect local machine environment.
    options:
      - verbose # same as --verbose
  linker:
    options:
      - Optimize: 0 # no optimization
  linter:
    options:
      - strict
packages:
  - main:
      name: newProject
      version: 0.0.1
      root: ./src
      import: []
```
- [x] Creates the build target directory (toolchain.compiler.target) if it doesn't exist
- [x] Creates the package source directory (packages.main.root) if it doesn't exist
- [x] Idempotent re‑run (no destructive changes if file exists; use `--force` to overwrite)
    - If `--force` is used, the tool will only add missing information.
- [x] Print error if local directory is not a git repo or execute `git init` if `--force` is used.
- [x] Tests:
    - create new workspace,
    - re‑init idempotence, `--json` output
    - re-init idempotence on `--force` with no missing fields (expect no change)
    - re-init idempotence on `--force` with missing fields (expect tool to re-add missing fields)
- [x] Add .gitignore with `./build` to the project root directory on `ami init`
### 1.1.2.0. Build Artifacts Cleanup (`ami clean`)
- [x] Command removes and recreates `./build` directory
- [x] Command emits actions via JSON/human format
- [x] Tests: fresh repo, with existing files, permissions edge cases
### 1.1.3.0. Package Cache Cleanup (`ami mod clean`)
- [X] Command removes `${AMI_PACKAGE_CACHE}`, then recreate `${AMI_PACKAGE_CACHE}` (empty)
- [X] Tests: add a dummy file to `${AMI_PACKAGE_CACHE}` then ensure that the clean function creates an empty directory
### 1.1.4.0. Package Cache Update (`ami mod sum`)
- [X] Command validates format of `ami.sum`
- [X] Command iterates over packages in `ami.sum` and verifies their hashes
- [X] Command uses SSH+GIT to pull down any packages in `ami.sum` which are not in `${AMI_PACKAGE_CACHE}`
- [X] Command compares `ami.workspace` to `ami.sum` to determine any missing packages and using that result the command
      downloads missing packages using SSH+GIT, updating `ami.sum` as they are downloaded
### 1.1.5.0. Package Fetch (`ami mod get <url>`)
- [X] `ami mod get <url>`: fetch a package into `${HOME}/.ami/pkg/<name>/<version>`
- [X] Supported sources (initial):
    - [X] `git+ssh://host/path#<semver-tag>` (key‑based authentication only; no interactive prompts).
      - Example: `git+ssh://git@github.com/org/repo.git#v1.2.3`
    - [X] Local path: `./subproject` (must be within workspace and declared in `ami.workspace`).
- [X] Sources are modular to allow later HTTPS implementation (internal runner separated by source type).
- [X] Command updates `ami.sum` as packages are downloaded (schema `ami.sum/v1`, object form with name → {version, sha256}).
  - [X] Audit pre‑checks integrated into `ami mod update`: non‑fatal audit summary in human mode and `audit` object in JSON output.
### 1.1.6.0. Package List (`ami mod list`)
- [x] Command lists all packages and versions in the package cache `${AMI_PACKAGE_CACHE}`
- [x] `ami mod list`: list cached packages (name, version, size, updated)
### 1.1.6.1. Package Audit (`ami mod audit`)
- [x] Command audits workspace dependency requirements against `ami.sum` and the package cache.
- [x] Outputs:
  - [x] JSON: single object with fields `requirements`, `missingInSum`, `unsatisfied`, `missingInCache`, `mismatched`, `parseErrors`, `sumFound`, `timestamp`.
  - [x] Human: concise summary lines and an `ok:` line when no issues.
- [x] Tests:
  - [x] Unit tests for command and workspace audit orchestration.
  - [x] End‑to‑end test under `tests/e2e/ami_mod_audit_test.go` launching the built CLI and validating stdout/stderr.
  - [x] E2E pattern documented in `docs/test-patterns/README.md`.
### 1.1.7.0. Project Linter (`ami lint`)
> Goal: Help users maintain code quality by providing a built-in linter as a first-class toolchain utility
- [x] Define lint entry points for AMI language sources (`.ami`) in workspace packages (Stage A scaffold)
    - [x] linter starts in 'main' package root (scaffold)
    - [X] linter processes imports in linear, recursive order (top-down, child first)
- [x] Implement basic rules aligned to language spec
    - naming,
        - [x] baseline: package name style (no snake_case)
        - [x] expand to identifiers (scanner‑backed underscore rule)
    - imports,
        - [x] verify import syntax (workspace)
        - [x] verify local import paths exist (workspace)
    - [X] verify package versioning rules are satisfied.
    - [x] unused,
    - [x] unknown identifiers (scaffold: sentinel UNKNOWN_IDENT in .ami)
    - [x] formatting markers
- [x] Enforce package versioning and import rules consistent with Chapter 3.0 (e.g., valid SemVer in package 
      declarations/imports, allowed characters in package names)
  - [X] Validate package declarations use valid SemVer (W_PKG_VERSION_INVALID)
  - [X] Validate local import constraints against local packages (E_IMPORT_CONSTRAINT; warn in non‑strict)
  - [X] Validate identifier naming rules beyond package names (scanner‑backed)
- [X] Output formats for lint:
    - [x] human summary to stdout
    - [x] `diag.v1` JSON lines with a final summary record; when `--verbose`, also stream per‑record NDJSON to `build/debug/lint.ndjson`
- [x] Expand lint rules to cover more of the language spec as it stabilizes
    - [x] Naming/style: package/id naming conventions, ban `_` identifier outside allowed sinks
        - [X] Warn on underscores in identifiers (W_IDENT_UNDERSCORE); pragma/config suppressible
    - [x] Imports: duplicate imports (W_IMPORT_DUPLICATE); stable ordering (W_IMPORT_ORDER);
          disallowed relative paths to parents (W_IMPORT_RELATIVE); invalid version constraints (W_IMPORT_CONSTRAINT_INVALID)
    - [X] Imports: duplicate alias (W_DUP_IMPORT_ALIAS), unused (W_UNUSED_IMPORT)
        - [X] Parser-backed unused imports for ident-form imports (W_UNUSED_IMPORT)
    - [x] Code hygiene: unreachable nodes/edges, duplicate function declarations across files (lint layer), 
          [X] TODO/FIXME policy
    - [x] Language‑specific:
        - [X] Reminders and detection: `W_LANG_NOT_GO` (info/warn), `W_GO_SYNTAX_DETECTED` (warn)
            - [X] Detect Go-like files starting with `package` (W_GO_SYNTAX_DETECTED)
- [x] Enforce/propagate AMI semantics via analyzer diagnostics surfaced in lint: 
      `E_MUT_BLOCK_UNSUPPORTED`, `E_MUT_ASSIGN_UNMARKED`, `E_PTR_UNSUPPORTED_SYNTAX`
  - [X] Integrate memory-safety analyzer for `E_PTR_UNSUPPORTED_SYNTAX` and `E_MUT_BLOCK_UNSUPPORTED` (Stage B)
  - [X] RAII hint: `W_RAII_OWNED_HINT` when `release(x)` not wrapped in `mutate(...)` (parser-backed)
  - [X] Collections: `W_MAP_*`, `W_SET_*`, `W_SLICE_ARITY_HINT` mirrored as warnings
  - [X] Pipelines: ingress/egress position hints (W_PIPELINE_INGRESS_POS, W_PIPELINE_EGRESS_POS)
- [X] Lint: severity configuration and rule suppression (pragma/config)
    - [x] Severities: error | warn | info (defaults per rule documented); `off` disables a rule
    - [x] Configuration: `ami.workspace` → `toolchain.linter.rules["RULE"] = "error|warn|info|off"`
    - [x] Inline suppression: `#pragma lint:disable RULE[,RULE2]` and `#pragma lint:enable RULE` (file‑wide scope, scaffold applied to source diags)
    - [X] File/package suppression via config; per‑directory overrides via `toolchain.linter.suppress` (path → codes)
    - [X] Strict mode preset: elevate warnings to errors (`--strict` or workspace config)
    - [X] Lint: include line/column positions in diagnostics where available
        - [X] Attach source file + `{line,column,offset}` to lint `diag.v1` records (when resolvable)
        - [X] Fall back to file‑only when exact positions are unavailable
        - [X] Tests validate position presence and formatting
- [X] Lint: cross‑package import/constraint consistency checks (strict mode)
    - [X] Validate conflicting exact versions across packages (E_IMPORT_CONSTRAINT_MULTI); promoted by strict
    - [X] Detect forbidden prerelease imports when constraints omit prereleases (E_IMPORT_PRERELEASE_FORBIDDEN)
    - [X] Ensure consistent import of the same package/version across the workspace (single version rule in strict mode)
        - Strict requires an exact pinned version per import path; ranges alone are flagged (W_IMPORT_SINGLE_VERSION) and promoted to error in strict.
    - [X] Report range incompatibilities conservatively (E_IMPORT_CONSTRAINT)
### 1.1.8.0. Project Test Runner (`ami test`)
- [X] Executing `ami test [path]` runs tests in the target directory
  - Collects `_test.go` via `go test -json ./...` and emits a human OK on success
  - On failure, writes errors to stderr and returns a non-zero exit
- [X] With `--verbose`, writes test results to `build/test/test.log`
- [X] With `--verbose`, writes `build/test/test.manifest` listing each test in execution order
 - [X] With `--json`, streams `go test -json` events to stdout and emits a final summary object including `ok`, `packages`, `tests`, and `failures` counts.
### 1.1.9.0. Project Builder (`ami build`)
#### 1.1.9.1. Configuration
- [X] The build / parse / compiler tool is configured by `ami.workspace`
- [X] If `toolchain.compiler.env` is empty, default to a single target `darwin/arm64` for this project phase.
#### 1.1.9.2. Build process:
  - [X] Compiler can produce consistent error messages when defects are identified in the sources
  - [X] Compiler can generate a json build plan if `--verbose` is used.
  - [X] Compiler can track token position to localize detected errors by line and position
  - [X] Using `ami.workspace` ensure that all local and remote packages are available on the local machine
  - [X] For every included source file (starting with `main.ami`)...
    - [X] Detect circular references and return an error/terminate
    - [X] Lex/Tokenize/Parse: Source → tokens → AST (per .ami file)
    - [X] Type-Checking: resolve names, perform inference/checks and const folding.
    - [X] If verbose is used, write AST and other required information to build/debug/ files.
    - [X] Variable declarations and local bindings (to enable broader local type inference in bodies).
- [X] Tuple/multi-value returns syntax and parsing.
- [X] Container literal syntax for `slice<T>`, `set<T>`, and `map<K,V>` (for container inference rules):
    - `slice<T>{e1, e2, ...}`
    - `set<T>{e1, e2, ...}`
    - `map<K,V>{k1: v1, k2: v2, ...}`
    - [X] Attach comments to function-body statements (top-level already covered).
- [X] Import lines with version constraints: accept `import <module> >= vX.Y.Z` (single and block forms), represent in AST (`ImportDecl.Constraint`) and surface in `sources.v1` (`importsDetailed`).
- [X] Function type parameters (scaffold): `FuncDecl.TypeParams []TypeParam{Name, Constraint}` and tolerant parser for `func F<T>(...)`/`func F<T any>(...)` (no semantics yet).
- [X] Types & Semantics
    - [X] Type inference M1 completion: inference for locals (idents), unary/binary expression inference for common cases; position-rich diagnostics on mismatches.
    - [X] Diagnostics: implement `E_TYPE_AMBIGUOUS` with source positions for ambiguous container literals (no type args and no elements, or any/any for maps without elements).
    - [X] Expand `E_TYPE_AMBIGUOUS` to returns/assignments/expr statements; ensure diagnostics include precise positions consistently.
    - [X] Type inference M2: container element/key inference; tuple return inference; propagation through `Event<T>` / `Error<E>`.
        - [X] Tuple return checks at return-sites (arity/type unification).
        - [X] Container element/key inference from literals; enforce consistent element/key/value types; diagnostics on mismatch.
    - [X] Type inference M3 (conservative): allow generic `Event<typevar>` to flow across pipeline steps without mismatch (conservative compatibility rule).
    - [X] Broader local scoping inference (shadowing, nested blocks) and return inference without annotations.
    - [X] Propagation of inferred container types across assignments, function calls, and returns.
    - [X] Cross‑package name resolution (multi‑file), constant evaluation, and additional validation rules per Phase 2.1 scope.
-  [ ] IR & Codegen
  - [X] IR (SSA) Construction (scaffold): emit `ssa.v1` debug per unit with straight-line SSA versioning of defs
  - [ ] Optimization and Analyses:
    - [ ] inlining
    - [ ] escape analysis
    - [ ] devirtualization
    - [ ] nil/bounds-check elim
    - [ ] CSE/DCE
    - [ ] raii violations / issues
    - [ ] loop opts
    - [ ] PGO (if enabled)
    - [ ] capability (I/O) violations (nodes exceeding permissions)
    - [X] Lower a minimal imperative subset with typed annotations into IR for debug (scaffold ops: VAR, ASSIGN, 
          RETURN, DEFER, EXPR).
    - [X] Enrich typed IR lowering: add SSA‑like temporaries and typed results; lower function calls 
          (callee + arg types) and container literals explicitly.
    - [X] Emit object stubs (scaffold) and per‑package object index in `build/obj/<pkg>/index.json`; build plan includes `objIndex` (verbose).
#### 1.1.9.3. Code generation / Linking:
- [ ] Instruction selection, register allocation, scheduling;
- [ ] emit object w/ relocations,
- [ ] export data, and tables (e.g., pclntab).
- [ ] Linking:
    - cmd/link resolves symbols,
    - does whole-program dead-code elim,
    - applies relocations (randomized layout for security,
    - lays out init order,
    - generates itabs/reflect data/GC tables/DWARF,
    - embeds assets,
    - and produces the binary (PIE/static as configured).
- [X] Invoke compiler driver to compile workspace packages into `toolchain.target/toolchain.env[]`
  directory (e.g. `./build/${env}`) (deterministic file layout).
- [X] Compiler can generate final LLVM object code as object artifacts
- [ ] Compiler can generate the LLVM runtime artifacts needed to produce the AMI binary.
- [ ] Compiler can link the generated LLVM runtime and object code into an executable on darwin/arm64.
- [ ] Failure modes:
  - syntax/type errors → USER_ERROR;
  - missing files → SYSTEM_IO_ERROR
- [X] JSON diagnostics (when `--json`):
  - [X] Workspace schema violation → `diag.v1` with `code:"E_WS_SCHEMA"` and descriptive message.
  - [X] Syntax errors: streams `diag.v1` records per error; exits with 1.
  - [X] Semantic errors (e.g., worker signature): streams `diag.v1` records; exits with 1.
  - [X] Cache vs `ami.sum` integrity mismatch → per‑item `diag.v1` records and a summary `diag.v1` with `code:"E_INTEGRITY"`; exits with 3.
  - [X] Existing `ami.manifest` vs `ami.sum` mismatch → `diag.v1` with `code:"E_INTEGRITY_MANIFEST"`; exits with 3.
- [X] Tests: minimal project build, multi‑package, missing deps, repeatability
  - [X] Multi‑package determinism for non‑debug obj indexes and asm
  - [X] Parser diagnostics stream multiple records in JSON; exit 1
  - [X] Missing file I/O emits `diag.v1` and exits 2 (JSON) and prints clear error (human)
  - [ ] End to End testing of compiled binaries to ensure ami compiler produces working binaries.
- [X] Directory layout is deterministic and mirrors the logical package/unit structure; all paths are relative to workspace.
 - [X] Do not emit debug artifacts without `--verbose`.
- [X] Ensure artifacts are reproducible across runs (given the same inputs) and contain ISO‑8601 UTC timestamps only where needed (e.g., top‑level metadata), never embedded in the core structures used by tests.
- [X] Include these paths in the build logs (JSON): `objIndex`, `buildManifest` (manifest also lists debug refs when verbose)
- [X] Rewrite `ami.manifest` in `build/ami.manifest`
    - [X] Contains `ami.manifest` content (packages map from ami.sum when present)
    - [X] Contains toolchain metadata (targetDir, targets)
    - [X] Contains evidence of built artifacts (objIndex entries)
    - [X] Contains evidence of all imported artifacts with build‑time integrity validation of `ami.sum` vs cache (verified/missing/mismatched)
    - [X] Contains cross references to ./build/debug artifacts when verbose
    - [X] Contains list of binaries produced in ./build/**/*
- [X] Tests:
    - [X] With `--verbose`, expected files exist and validate against schemas (AST/IR JSON), assembly is non‑empty.
    - [X] Without `--verbose`, no `build/debug/` directory is created.
    - [X] Contents are deterministic (AST/IR/ASM debug artifacts stable across runs).
- [X] Build plan emitted and validates in verbose and JSON modes.
  - Build Plan schema fields (stable):
    - `packages[].hasObjects` (bool): true if any `.o` exists under `build/obj/<pkg>/`. Stable and backward‑compatible.
    - `objects[]` (optional, array<string>): workspace‑relative paths to discovered `.o` files. Optional and additive.
  - Manifest additions (stable):
    - `objects[]` (optional, array<string>): workspace‑relative `.o` paths present under `build/obj/**`. Optional and additive.
#### 1.1.9.4. Debug Artifacts when `--verbose` is set
- [X] When `--verbose` is provided to `ami build`
  - generate debugging information in `./build/debug` for compiler debugging (not produced otherwise)
    - [X] A Resolved source stream (`build/debug/source/resolved.json`).
      - [X] An Abstract Syntax Trees (per package/unit)
          - [X] stored in .json format
          - [X] stored as `build/debug/ast/<package>/<unit>.ast.json`
          - [X] using stable field ordering and positions for nodes (imports, funcs, pipelines, steps, pragmas).
      - [X] Artifacts under `./build/debug/` to aid compiler debugging (not produced otherwise)
      - [X] Full timestamped activity logs for the compiler in `./build/debug/activity.log`
  - [X] Intermediate Representation (IR)
        - [X] Pipelines IR (debug): `build/debug/ir/<package>/<unit>.pipelines.json`
            - Captures pipeline steps and referenced workers, including generic payloads for inputs/outputs (T/U/E) to enable future type-compatibility checks.
        - [X] Final stage IR stored as `build/debug/ir/<package>/<unit>.ir.json`
            - Lowered IR in JSON capturing control/data flow before codegen; stable for golden tests.
        - [X] Edges summary (debug): `build/debug/asm/<package>/edges.json` (`edges.v1`)
            - Per‑package list of input edge initializations with derived semantics: `bounded` (true when `maxCapacity>0`) and `delivery` (`atLeastOnce` for `block`, `bestEffort` for `dropOldest`/`dropNewest`). Also embedded into `asm.v1` index under optional `edges`.
      - [X] Unlinked assembly artifacts: `build/debug/asm/<package>/<unit>.s` and index `asm.v1`
        - [X] Target assembly emitted prior to link steps; human‑readable text.
      - [X] Manifest: enumerate all debug artifacts across all packages as structured JSON:
        - including
          - resolved,
          - AST,
          - IR,
          - pipelines,
          - eventmeta,
          - ASM,
          - per‑package `index.json`,
          - and `edges.json` when present.
### 1.1.10.0. Pipeline Visualizer (`ami pipeline visualize`)
> Goal: Provide quick, in‑terminal ASCII visualizations of AMI pipelines to aid understanding and debugging without external tooling.

- Command hierarchy
    - Parent: `ami pipeline`
    - Subcommand: `visualize`
- Behavior
    - Reads `ami.workspace` to locate the main package (default `packages.main.root`) and source roots.
    - Discovers pipelines in sources and renders each as an ASCII graph to stdout in human mode.
    - When `--json` is present, emits a machine‑readable structure (e.g., `graph.v1`) describing nodes, edges, and attributes instead of ASCII.
    - Supports `--package <key>` and `--file <path>` to narrow scope (optional for initial milestone; default is main package and all `.ami` units).
    - Honors global flags; `--color` may colorize node kinds and edges in human mode (never with `--json`).
- Rendering (human)
    - ASCII only (no Unicode box‑drawing) for broad terminal compatibility.
    - Deterministic layout: stable ordering of nodes and edges; fixed width spacing; wrap long labels with ellipses.
    - One pipeline per block; separate blocks with a blank line; prefix with pipeline name and package.
    - Example (illustrative):
      ```
      package: main  pipeline: IngestToStore
      [ingress] --> (WorkerA) --> (WorkerB) --> [egress]
                    |                         
                    +--> (ErrorHandler) ------+
      ```
- JSON (`--json`)
    - Schema `graph.v1` (stable ordering): `{ package, unit, name, nodes:[{id, kind, label}], edges:[{from, to, attrs}] }`.
    - [X] Final summary record `{ schema:"graph.v1", type:"summary", pipelines:<n> }` emitted after graphs.
    - [X] Detect circular references; emit `diag.v1` error (`E_GRAPH_CYCLE`) and terminate with non‑zero exit. Human mode returns an error.
- Inputs
    - Source discovery aligns with build/lint: start at `packages.main.root` and include direct imports (workspace‑local only for initial phase).
    - Parser/AST extraction provided by compiler front‑end (Agent D). This command consumes the tolerant shape (no semantics beyond graph extraction).
- Errors
    - On missing workspace or packages, exit `USER_ERROR` and print a clear message; with `--json`, emit a `diag.v1` error record.
    - On parse/graph extraction failures, emit `diag.v1` stream and exit 1.
- Tests
    - Unit tests for renderer: given a minimal graph structure, assert ASCII layout (goldens) and JSON structure.
    - CLI tests: run command with/without `--json`, validate exit code and outputs.
    - E2E test: small sample pipeline sources under `build/test/e2e/pipeline_visualize/...` and assert the ASCII output contains expected lines.
- Future
    - Add `--focus <node>` to center graph; `--width` to control wrapping; `--legend` toggle; `--save <file>` to write ASCII/JSON to a file.
### 1.1.0.10. Version Subcommand (`ami version`)
- [X] Subcommand with build‑time injected version (ldflags)
### 1.1.0.11. Help Subcommand (`ami help`)
- [X] subcommand generated by converting `docs/help-guide/*.md` into compiled content
    - `docs/help-guide/README.md` and `docs/help-guide/**/*.md` provide end user content
    - `help.go` uses `go:embed` to consume `docs/help-guide/README.md` into the `ami` artifact
      to provide help content.
### 1.1.1.0. Dependency Management (packages/cache)
#### 1.1.1.1. Package Cache Directory
- [X] The environment variable AMI_PACKAGE_CACHE is used to locate the package cache directory when `ami` starts
- [X] If AMI_PACKAGE_CACHE is defined but does not exist, `ami` will create it
- [X] If AMI_PACKAGE_CACHE is not defined `ami` defaults to the `${HOME}/.ami/pkg` directory.
#### 1.1.1.2. Versioning Selection
  - [X] Implement SemVer parsing/validation per Chapter 3.0 (see “package versioning rules” and SemVer regex); reject invalid versions.
  - [X] If `<version>` is omitted, select the highest non‑prerelease SemVer tag by default (prereleases excluded unless explicitly requested in the constraint).
  - [X] Respect version constraints from `ami.workspace` when updating (`mod update`): evaluate existing `ami.sum` entries and report selected highest satisfying versions (non‑destructive); supports `^`, `~`, `>`, `>=`, exact `vX.Y.Z`; prereleases excluded unless explicitly requested.
  - [X] Tests: selection with/without prereleases, constraint satisfaction, invalid versions.
  - [ ] Integrity:
    - [X] Verify checksums in `ami.sum` (sha256) against cache contents; mismatches fail with INTEGRITY_VIOLATION_ERROR.
    - [ ] Verify signatures if provided (schema extension pending; failure results in INTEGRITY_VIOLATION_ERROR)
  - [X] Network errors return NETWORK_REGISTRY_ERROR

# Remaining Work
- No pending items for scanner; core features implemented and tested. Follow-up may refine diagnostics integration once the diag package lands.
  - Tests:
    - [X] cache clean/recreate
    - [X] get/list/update happy/sad paths
    - [X] integrity failure
    - [X] offline behavior (mod get/mod sum)
#### 1.1.1.3. Dependency Summary File: `ami.sum` (JSON)
- [X] Create the `workspace.Manifest` struct to represent `ami.sum` (JSON) in memory (`src/workspace/manifest.go`)
- [X] `workspace.Manifest` has a `Load(path string)` method to deserialize the file `ami.sum`.
- [X] `workspace.Manifest` has a `Save(path string)` method to serialize and write `ami.sum`.
- [X] `workspace.Manifest` has a `Validate()` method to verify `ami.sum` packages against AMI_PACKAGE_CACHE.
- [X] the file maps packages and their semver tags to the git commit hash (sha-256)
- [X] Format (canonical JSON, UTF‑8, LF newlines, stable key order on write):
  - Object form, nested by package then version:
    {
      "schema": "ami.sum/v1",
      "packages": {
        "github.com/example/foo": {
          "v1.2.3": "<sha256-commit-oid>",
          "v1.2.4": "<sha256-commit-oid>"
        },
        "git.example.com/org/bar": {
          "v0.9.0": "<sha256-commit-oid>"
        }
      }
    }
  - Implementation note: entries produced by `mod get`/`mod sum` in object form may include an optional `commit` field for traceability. This does not affect integrity checks (which continue to use directory `sha256`) and will be finalized alongside the Resolution rules.
- [ ] Resolution rules:
  - [X] `ami mod get <url>@<semver>` resolves the tag (e.g., `v1.2.3`) to a commit.
  - [X] If the remote repository supports Git SHA‑256 object format, record that commit OID directly.
  - [X] If the remote repository uses SHA‑1, derive a SHA‑256 identifier deterministically from the raw commit object content (Git‑canonical header `"commit <len>\0"` + body) and record the resulting SHA‑256 digest as `<sha256-commit-oid>`.
  - [X] Do not hash tarballs; the digest represents the commit object for the tag.
  - Note: current implementation records directory content `sha256` for integrity checks and attaches an optional `commit` field for traceability. Migration to commit‑digest as the canonical value will be completed with this section.
- [X] Write/update behavior:
  - [X] On verify, all required dependencies in `ami.workspace` have entries in `ami.sum` and local cache contents
    match the recorded digest; any mismatch or missing → `INTEGRITY_VIOLATION_ERROR (3)`.
- [ ] Tests:
  - [X] Create `ami.sum` from empty via update/get; ensure deterministic ordering (canonical key sort).
  - [X] SHA‑256 recorded from raw commit object for annotated and lightweight tags; deterministic digest. (Guarded by AMI_E2E_ENABLE_GIT=1)
  - [X] Detect and error on digest mismatch (cache tamper) with exit code 3 (build integrity test).
  - [X] `ami.sum` is not removed by `ami clean` and persists across builds.
##### CLI & Output
- [X] Flags: `--strict`, `--rules=<pattern>`, `--max-warn=<n>` (regex `/.../` and `re:<expr>`, glob `*?[]`, and substring supported). `--json` and `--color` are global flags.
- [ ] JSON: `diag.v1` codes use `LINT_*` namespace; include `file`, `pos`, and `data` fields where relevant
  - Note: current JSON uses existing rule codes (e.g., `W_*`, `E_*`). A compatibility alias `LINT_*` is available under `data.compatCode` when `--compat-codes` is set. Full code namespace migration remains pending.
- [X] Human: severity prefixes; counts summary; non‑zero exit on errors (and on warnings when `--strict`)
##### Tests & Docs
- [X] Unit tests for suppression and severity configuration; tests for import alias duplication and order; position assertions
- [X] Integration tests for cross‑package constraint checks using temporary multi‑package workspaces
- [X] Docs: `docs/lint.md` updated with rules list, severities/suppression, and CLI flags (`--strict`, `--rules`, `--max-warn`) with examples
## 1.2.0.0. Supporting Requirements
### 1.2.1.0. AMI Language (POP)
- [X] Parser and AST scaffold for AMI language (Chapter 2)
- [X] Lexical structure (2.1): UTF‑8, tokens, comments
- [X] Pipeline grammar: nodes, edges, chaining, config
- [X] Multiple entrypoints (multiple pipeline declarations)
- [X] Error pipeline parsing: `pipeline P { ... } error { ... }` captured in AST
 - [X] Concurrency and scheduling declarations (1.4, 2.3.6): collected via `#pragma concurrency` and exposed through IR attributes
 - [X] Compiler directives: 
   - [X] Backpressure collected via `#pragma backpressure` and mapped into IR (config) and pipelines.v1 default delivery
   - [ ] Capabilities/trust (deferred)
     - Deferred to a future milestone. No runtime semantics or enforcement yet; no tests. Scope will include trust boundaries, IO/capability annotations, and IR/codegen surface once enabled.
 - [X] Telemetry hooks (1.6): collected via `#pragma telemetry` and surfaced in ASM header
- [X] Parser/semantics diagnostics for package identifiers and import paths (happy/sad tests)
- [X] Basic node semantics: pipeline must start with `ingress` and end with `egress`; unknown nodes emit diagnostics
- [ ] Event typing, error typing, and contracts (1.7, 2.2)
  - [ ] Event schema (events.v1): id, timestamp, attempt, trace context; immutable payload typing and supported containers
  - [ ] Error schema (align with diag.v1): stable codes/levels; optional position and data fields
  - [ ] Contracts: node I/O shape declarations; buffering/order guarantees; backpressure policy; capability declarations (io.*)
  - [ ] Validation: schema validators for events/errors; CLI hooks where appropriate
  - [ ] Tests: unit + integration tests; JSON structural validation for schema conformance
  - Docs: see `docs/events.md` for beginner-friendly overview and plan.
- [X] Worker function signatures and factories (2.2.6, 2.2.13)
 - [ ] Node‑state tables and access (2.2.14, 2.3.5)
   - Deferred. No implementation yet in runtime/IR. Work will introduce a keyed ephemeral state store per node with scoped read/write APIs, deterministic serialization, and tests for visibility, lifetime, and isolation across nodes/pipelines.
 - [ ] Memory model: ownership & RAII (2.4)
- [ ] Observability: event IDs, telemetry hooks (1.6)
#### 1.2.1 Lexical Structure (Chapter 2.1)
- [X] Source text: UTF‑8, LF newlines; files end with newline
- [X] Identifiers: 
  - start `[A‑Za‑z_]`, 
  - continue `[A‑Za‑z0‑9_]`, 
  - case‑sensitive
  - max length: 255 characters
- [X] Keywords (tracked in token.Keywords): 
  - `_` (blank) recognized as IDENT;
  - `append`: KwAppend
  - `break`: KwBreak
  - `case`: KwCase
  - `const`: KwConst
  - `bool` : KwBool
  - `break`: KwBreak
  - `byte` : KwByte
  - `case`: KwCase
  - `close`: KwClose
  - `collect`: KwNodeCollect
  - `complex`: KwComplex
  - `complex64`: KwComplex64
  - `complex128`: KwComplex128
  - `const`: KwConst
  - `continue`: KwContinue
  - `copy`: KwCopy
  - `default`: KwDefault
  - `defer`: KwDefer
  - `delete`: KwDelete
  - `egress`: KwNodeEgress
  - `else`: KwElse
  - `enum`: KwEnum,
  - `error`: KwError,
  - `Error`: KwErrorEvent
  - `ErrorPipeline`: KwErrorPipeline
  - `Event`: KwEvent
  - `fallthrough`: KwFallthrough
  - `false`: KwFalse
  - `fanout`: KwNodeFanout
  - `float`: KwFloat
  - `float32`: KwFloat32
  - `float64`: KwFloat64
  - `for`: KwFor
  - `func`: KwFunc
  - `goto`: KwGoto
  - `if`: KwIf
  - `import`: KwImport
  - `ingress`: KwNodeIngress
  - `int`: KwInt
  - `int8`: KwInt8
  - `int16`: KwInt16
  - `int32`: KwInt32
  - `int64`: KwInt64
  - `int128`: KwInt128
  - `interface`: KwInterface
  - `label`: KwLabel
  - `latest`: KwLatest
  - `len`: KwLen
  - `make`: KwMake
  - `map`: KwMap
  - `mutable`: KwNodeMutable
  - `new`: KwNew
  - `nil`: KwNil
  - `package`: KwPackage
  - `panic`: KwPanic
  - `pipeline`: KwPipeline
  - `range`: KwRange
  - `real`: KwReal
  - `recover`: KwRecover
  - `return`: KwReturn
  - `rune`: KwRune
  - `select`: KwSelect
  - `set`: KwSet
  - `slice`: KwSlice
  - `state`: KwState
  - `string`: KwString
  - `struct`: KwStruct
  - `switch`: KwSwitch
  - `transform`: KwNodeTransform
  - `true`: KwTrue
  - `type`: KwType
  - `uint`: KwUint
  - `uint8`: KwUint8
  - `uint16`: KwUint16
  - `uint32`: KwUint32
  - `uint64`: KwUint64
  - `uint128`: KwUint128
  - `var`: KwVar  
- [X] Literals: 
  - integer,
  - float (decimal),
  - string with basic escapes
- [X] Operators, symbols, delimiters (tracked in token/symbols.go): 
  - `.`: SymPeriod 
  - `|`: SymPipe 
  - `,`: SymComma 
  - `;`: SymSemicolon
  - `:`: SymColon
  - `(`: SymLParen
  - `)`: SymRParen
  - `[`: SymLBracket
  - `]`: SymRBracket
  - `{`: SymLBrace
  - `}`: SymRBrace
  - `=`: SymEQ
  - `==`: SymBoolEQ
  - `!=`: SymBoolNE
  - `<`: SymBoolLT
  - `<=`: SymBoolLE
  - `>`: SymBoolGT
  - `>=`: SymBoolGE
  - `+`: SymPlus
  - `-`: SymHyphen
  - `*`: SymAsterisk
  - `/`: SymSlash
  - `%`: SymParenthesis
  - `\\`: SymBackSlash
  - `$`: SymDollarSign
  - `!`: SymExclamation
  - `\``: SymTick
  - `~`: SymTilde
  - `?`: SymQuestion
  - `@`: SymAt
  - `#`: SymPound
  - `^`: SymCaret
  - `"`: SymDoubleQuote
  - `'`: SymSingleQuote
- [X] Comments: 
  - `//` line comment, 
  - `/* ... */` block;
  - skipped by scanner
- [X] Special underscore 
  - `_` sink identifier semantics (semantic checks)
- [X] Package identifiers and import paths (validation rules)
- [X] Compiler directives (2.1.10): 
  - `#pragma`
    - style for concurrency,
    - used to restrict files to specific opsys/arch environments
### 1.2.2.0. Pipeline Grammar and Semantics (Ch. 1.1, 2.2)
- [X] Multiple entrypoints (1.1.5, 1.8.2): program may define multiple named ingress pipelines
- [X] Pipelines create graphs of nodes (e.g., Ingress) configured by attributes:
  ```
  pipeline <PipelineNameIdentifier> {
      <node>(
         <attribute>=<value>,
         <attribute>=<value>,
         <attribute>=<value>,
      ).<node>(
         <attribute>=<value>,
         <attribute>=<value>,
         <attribute>=<value>,
      ).<node>(
         <attribute>=<value>,
         <attribute>=<value>,
         <attribute>=<value>,
      ).<node>(
         <attribute>=<value>
         <attribute>=<value>,
         <attribute>=<value>,
         <attribute>=<value>,
      )
  }
  ```
  where valid nodes are `Ingress`,`Transform`,`Fanout`,`Collect`,`Egress`
- [X] Node‑chained notation (2.2.4): canonical attribute form 
  ```
    Ingress(
       ...<attribute list>...
    ).Transform(
       ...<attribute list>...
    ).Fanout(
       ...<attribute list>...
    ).Collect(
       ...<attribute list>...
    ).Egress(
       ...<attribute list>...
    )
  ```
- [ ] Nodes (2.2.5–2.2.11) — basic invariants implemented in semantics:
  - [X] Pipeline must start with `ingress` and end with `egress` (`E_PIPELINE_START_INGRESS`, `E_PIPELINE_END_EGRESS`).
  - [X] `ingress` only allowed at position 0 (`E_INGRESS_POSITION`); unique (`E_DUP_INGRESS`).
  - [X] `egress` only allowed at last position (`E_EGRESS_POSITION`); unique (`E_DUP_EGRESS`).
  - [X] Unknown node kinds emit `E_UNKNOWN_NODE`.
  - [X] Capabilities (IO permissions): only ingress/egress may perform input/output.
    - [X] Low-Level I/O operations are performed by the built-in `io` package.
      - This allows easy detection.  Only Ingress and Egress should use the `io` package.
      - Each `io` package feature maps to one or more `capabilities` (e.g., io.Read→`io.read`, io.Write→`io.write`, io.Open→`io.fs`, io.Connect→`network`) for compile-time/runtime enforcement; IR includes `capabilities`.
    - [X] Detection (scaffold):  
      - [X] flags any non-ingress/egress node using `io.*)` calls (`E_IO_PERMISSION`).
- [ ] Edges (2.2.7):
  - Exist as the `edge.*` package (e.g., `edge.FIFO` and `edge.LIFO`)
  - `edge.FIFO` is a dynamic first-in-first-out (queue) structure generated by the compiler
    - `edge.LIFO` is a dynamic last-in-first-out (stack) structure generated by the compiler.
    - The `edge` package may be extended in the future to add other edge styles (e.g., internetworked edges).
  - [X] Validation in semantics:
    - min/max capacity ordering, 
    - non-negative min, 
    - valid backpressure (`block`|`dropOldest`|`dropNewest`|`shuntNewest`|`shuntOldest`), 
    - pipeline name required. Emit a linter warning on legacy `drop` alias.
  - [X] Cross‑pipeline type safety: `edge.Pipeline(name=X, type=T)` must match the output payload type of pipeline `X`; emit `E_EDGE_PIPE_NOT_FOUND` for unknown `name` and `E_EDGE_PIPE_TYPE_MISMATCH` on mismatched types. Conservative when upstream type cannot be inferred.
  - [X] IR lowering attaches parsed `edge.*` specs to steps (scaffold via debug):
    - [X] `pipelines.v1`: step edge object with derived `bounded` and `delivery` fields (defaults scaffolded).
    - [X] `edges.v1`: per-package summary with `bounded` and `delivery` (defaults scaffolded).
  - [X] Map `#pragma backpressure` defaults into IR attributes (scaffold via debug edge object defaults).
  - [X] Stub `edge.*` specs (FIFO, LIFO, Pipeline) as compiler-generated artifacts for analysis/codegen; see `src/ami/compiler/edge` and `docs/edges.md`.
  - [X] Event payload flow type checking (scaffold): when both sides declare explicit step type, mismatches emit `E_EVENT_TYPE_FLOW`.
  - [X] Error pipelines (1.1.8): parsing supported in AST (`error { ... }`), semantics validated:
  - [X] Error pipeline must end with `egress` (`E_ERRPIPE_END_EGRESS`).
  - [X] Error pipeline cannot start with `ingress` (`E_ERRPIPE_START_INVALID`).
  - [X] Unknown nodes in error path flagged as `E_UNKNOWN_NODE`.
- [X] Event lifecycle and metadata (1.1.6–1.1.7): id, timestamp, attempt, trace context; immutable payload
  - [X] Debug contract emitted per unit as `build/debug/ir/<package>/<unit>.eventmeta.json` (`eventmeta.v1`) with fields: `id`, `timestamp` (ISO‑8601 UTC), `attempt` (int), and structured `trace` context (`trace.traceparent`, `trace.tracestate`); plus `immutablePayload: true`.
  - Runtime semantics are deferred; compiler enforces immutable event parameter shape (no pointers) and records generics; body‑level immutability checks are deferred until the imperative subset lands.
### 1.2.3.0. Worker Functions (Ch. 1.5, 2.2.6, 2.2.13)
- [X] Canonical signature parsed and enforced for worker references (ambient state, no raw pointers):
  `func(ev Event<T>) (Event<U>, error)`
  - Ambient state access via `state.get/set/...` per docx and Memory Safety (no `*State` parameter). Legacy explicit `State` parameter is not allowed for workers.
- [X] Purity and sandboxing: enforced at pipeline level for IO nodes (ingress first, egress last); deeper IO checks deferred with runtime integration.
- [X] Factories: `New*` worker factories recognized (existence check only for now); pipeline semantics resolve factory calls to top‑level functions; unknown references emit `E_WORKER_UNDEFINED`; invalid signatures emit `E_WORKER_SIGNATURE`.
- [X] Execution context/state (scaffold): placeholders defined in compiler types for later runtime/execution integration.
### 1.2.4.0. Merge Package (Collect + edge.MultiPath Attributes)
> Goal: Provide a stdlib `merge` package and `edge.MultiPath(...)` edge spec to configure merging behavior on `Collect` nodes when joining multiple upstreams. Attributes are declared as `merge.*(...)` inside the `edge.MultiPath(...)` argument list.
- [X] Edge spec: `edge.MultiPath(...)`
  - [X] Parser tolerant acceptance of `edge.MultiPath(<kv/attr list>)` alongside existing `edge.FIFO`, `edge.LIFO`, and `edge.Pipeline`.
  - [X] Grammar: keys/attrs may be specified via `k=v` pairs and/or `merge.*(...)` attribute calls (order-insensitive; last-write-wins for duplicates).
  - [X] Semantics (scaffold): `edge.MultiPath` is valid only on `Collect` nodes; analyzer emits `E_MP_ONLY_COLLECT` for non‑Collect usage (code name differs from earlier draft `E_EDGE_MULTIPATH_CONTEXT`).
- [ ] Stdlib: `merge` package API (attributes)
  - [X] `merge.Sort(field[, order])`: field selector string (e.g., "ts", "id"), optional order in {asc, desc}; stable ordering when combined with `merge.Stable()` (scaffold).
  - [X] `merge.Stable()`: requests stable sorting semantics (scaffold warning when used without Sort).
  - [X] `merge.Key(field)`: key selector used by other attributes (e.g., dedup/partition).
  - [X] `merge.Dedup([field])`: remove duplicates; optional `field` overrides default key; conflict with Key when fields differ.
  - [X] `merge.Window(size)`: bounded in-flight window (basic validation + smells).
  - [X] `merge.Watermark(field, lateness)`: watermarking; tolerant lateness validation (warn on non-positive) (scaffold).
  - [X] `merge.Timeout(ms)`: max waiting time (must be >0) (scaffold).
  - [X] `merge.Buffer(capacity[, backpressure])`: internal buffer size and backpressure policy {block, dropOldest, dropNewest}. Warn on ambiguous `drop` alias.
  - [X] `merge.PartitionBy(field)`: partition streams by key prior to merging (scaffold + conflicts with Key when fields differ).
- [ ] Semantics & Diagnostics
- [X] `E_EDGE_MULTIPATH_CONTEXT`: `edge.MultiPath` used outside a `Collect` node. (Emitted as `E_MP_ONLY_COLLECT` in code.)
  - [X] `E_MERGE_ATTR_UNKNOWN`: unknown `merge.*` attribute.
  - [X] `E_MERGE_ATTR_ARGS`: wrong arity/type for `merge.*` attribute args.
  - [X] `E_MERGE_ATTR_CONFLICT`: conflicting attributes (e.g., duplicate with different values).
  - [X] `E_MERGE_ATTR_REQUIRED`: missing required attributes (e.g., `merge.Sort` without a field).
- [ ] IR & Tooling (scaffold)
  - [X] `pipelines.v1` carries `edge.MultiPath` on `Collect` with tolerant `inputs` list and raw `merge` ops (name/args). Full normalization deferred.
  - [X] `edges.v1` summary includes per‑Collect MultiPath snapshots when present (debug parity with pipelines.v1).
- [ ] Lint & Smells
  - [X] `W_MERGE_SORT_NO_FIELD`: `merge.Sort` specified without a field.
- [X] `W_MERGE_TINY_BUFFER`: `merge.Buffer` set to very small capacity (<=1) with `dropOldest`/`dropNewest` policy.
  - [X] `W_MERGE_WATERMARK_MISSING_FIELD`: `merge.Watermark` without a field.
  - [X] `W_MERGE_WINDOW_ZERO_OR_NEGATIVE`: invalid window size.
  - [X] Ensure rules are suppressible and configurable.
- [ ] Tests (scaffold)
  - [X] Parser round‑trips `edge.MultiPath(...)` as raw attr on `Collect`.
  - [X] Semantics: context enforcement (Collect‑only) and basic merge‑op name/arity validation.
  - [X] IR/codegen: pipelines schema encodes MultiPath; ASM listings emit `mp_*` pseudo‑ops; edges.v1 includes MultiPath snapshot (build test added).
  - [ ] Merge attribute normalization and per‑attribute validation (deferred).
- [ ] Documentation
  - [X] Add `docs/merge.md` describing attribute semantics, precedence, examples.
  - [ ] Update `docs/edges.md` to include `edge.MultiPath` and cross-reference `merge`.
  - [X] Add examples under `examples/` demonstrating `Collect` with `edge.MultiPath` and various `merge.*` settings.
- [ ] Runtime/Planner (deferred)
  - [ ] Plan how `merge` attributes map to runtime merge operator configuration; leave unimplemented until runtime integration phase.

See also: `docs/merge.md` for a design summary and examples.
### 1.2.5.0. edge.MultiPath() Merge Facility (Collect)
> Create a first‑class `edge.MultiPath()` facility for `Collect` nodes to express multi‑upstream merge behavior. This section focuses on the end‑to‑end checklist. Align with expectations described in `docs/Asynchronous Machine Interface.docx` (merge/collect semantics, ordering, watermarks, and buffering) and reconcile naming with the existing `merge.*` attribute API.
- [X] Grammar & Parser
- [X] Parser: Accept `in=edge.MultiPath(<attr list>)` as a step argument on `Collect` nodes (tolerant inputs + raw `merge` ops).
  - [ ] Semantics/IR (scaffold): Minimal validations (Collect‑only, input type checks). IR/schema/codegen scaffolded; merge normalization and full lowering pending.
  - [ ] Attr list supports mixed `k=v` pairs and nested `merge.*(...)` attribute calls; order‑insensitive; last‑write‑wins.
  - [ ] Tolerant parsing (scaffold): capture raw string, normalized kv map, and a list of `merge.*` calls for later semantic validation.
- [ ] Semantics & Validation
  - [ ] Context rule: `edge.MultiPath` is only valid on `Collect` nodes; analyzer emits `E_MP_ONLY_COLLECT`.
  - [ ] Semantics: context checks (Collect‑only), attribute arity/type validation, conflicts, and required fields.
    - Status: Minimal checks complete (Collect‑only, inputs required, FIFO first, type compatibility across inputs when types provided, allowed merge operator names). Deeper arity/type validation of individual `merge.*` attributes is pending.
  - [ ] Recognize attributes (minimum viable set):
    - [ ] `merge.Sort(field[, order])` where `order ∈ {asc, desc}`; default asc.
    - [ ] `merge.Stable()` to request stable ordering when sort keys tie.
    - [ ] `merge.Key(field)` to define a key for other operations.
    - [ ] `merge.Dedup([field])` to remove duplicates; default to `merge.Key` when field is omitted.
    - [ ] `merge.Window(size)` bounded in‑flight window.
    - [ ] `merge.Watermark(field, lateness)` (lateness in time/units as defined by docx; tolerant scaffold accepts int/str until finalized).
    - [ ] `merge.Timeout(ms)` overall wait before forcing a merge decision.
  - [X] `merge.Buffer(capacity[, backpressure])` with `backpressure ∈ {block, dropOldest, dropNewest}` (parser accepts both explicit policies; linter warns on legacy `drop`).
    - [ ] `merge.PartitionBy(field)` partition upstreams by key before merging.
  - [ ] Diagnostics:
    - [X] `E_MERGE_ATTR_UNKNOWN` (unknown attribute), `E_MERGE_ATTR_ARGS` (invalid arity/type),
      `E_MERGE_ATTR_REQUIRED` (missing required field), `E_MERGE_ATTR_CONFLICT` (conflicting directives), and any docx‑specific constraints.
  - [X] Edge policy interaction: when `merge.Buffer(..., backpressure in {dropOldest, dropNewest})` and `capacity<=1`, emit lint smell (not a hard error). Also warn when using ambiguous `drop`.
 - [ ] IR & Schemas
  - [X] pipelines.v1 carries MultiPath scaffold (inputs + raw merge ops).
   - [X] edges.v1 includes per‑Collect MultiPath snapshots for debugging.
  - [X] edges.v1 includes per‑Collect MultiPath snapshots for debugging.
- [ ] Lint (Smells & Hints)
- [X] `W_MERGE_SORT_NO_FIELD` (Sort without a field), `W_MERGE_TINY_BUFFER` (Buffer capacity<=1 with `dropOldest`/`dropNewest`), `W_MERGE_WATERMARK_MISSING_FIELD`, `W_MERGE_WINDOW_ZERO_OR_NEGATIVE`.
  - [X] Ensure rules are suppressible via pragmas and configurable via workspace severities.
- [ ] Tests
  - [X] Parser: accept and round‑trip `edge.MultiPath(...)` with `merge.*` attributes.
  - [ ] Semantics: context enforcement (Collect‑only), per‑attribute arity/type validation, conflicts, and required fields.
  - [ ] IR: golden JSON snapshots verifying normalized merge config.
- [X] Lint: smells/hints coverage for tiny buffer with `dropOldest`/`dropNewest`, invalid windows, missing fields.
  - [ ] Determinism: verify `merge.Sort` + `merge.Stable` yields stable order across runs with identical inputs (scaffold level).
- [ ] Documentation
  - [X] `docs/merge.md`: detailed attribute semantics, precedence/resolution rules, examples.
  - [ ] `docs/edges.md`: add `edge.MultiPath`, link to `merge`, and demonstrate Collect with multiple upstreams.
  - [X] Examples under `examples/` illustrating common merge recipes from the docx (sorting, deduplication, watermarking, buffering).
- [ ] Runtime/Planner (Deferred)
  - [ ] Map normalized IR configuration to a merge operator plan; define contracts for buffering, ordering, and watermark handling per docx guidance.

See also: `docs/merge.md` for attribute semantics and pipeline examples.
### 1.2.6.0. Imperative Subset (Ch. 2.3)
- [ ] Functions: declaration forms, parameters, returns, overloading rules (if any)
  - Parser: supports `func name(params) [result| (tuple) ] { body }`, named or type-only params, single or tuple returns.
  - [X] Semantics: no overloading (duplicate function names rejected as `E_DUP_FUNC`); blank identifier `_` illegal for function and parameter names (`E_BLANK_IDENT_ILLEGAL`, `E_BLANK_PARAM_ILLEGAL`).
  - Tests: parser covers params/tuple returns; semantics covers duplicates, blank names.
- [ ] Data mutability: immutable-by-default; assignments require explicit `*` marker on LHS, and `mutate(expr)` for side-effectful expressions
  - Parser: captures function body tokens and builds a simple statement AST; legacy `mut {}` is not recognized.
  - Semantics: any assignment without `*` on the LHS emits `E_MUT_ASSIGN_UNMARKED`; `mut {}` usage emits `E_MUT_BLOCK_UNSUPPORTED`.
  - Tests: happy (assignment with `*`), and sad (unmarked assignment, `*` misused on RHS) cases.
 - [X] Defer statements: `defer <call-expr>` inside function bodies
   - Parser: produces `DeferStmt` with position; supports function and method calls.
   - Semantics: integrated with RAII; deferred releases/transfers are applied at function end for leak checks; intervening uses do not trigger use-after-release.
   - Docs: `docs/language-defer.md` with syntax and examples.
   - [ ] Types: `enum`, `struct`, `map`, `set`, `slice` with constraints
    - [ ] `enum` is implemented 100% with >=80% test coverage and all tests passing.
    - [ ] `struct` is implemented 100% with >=80% test coverage and all tests passing.
    - [ ] `map` is implemented 100% with >=80% test coverage and all tests passing.
    - [ ] `set` is implemented 100% with >=80% test coverage and all tests passing.
    - [ ] `slice` is implemented 100% with >=80% test coverage and all tests passing.
- [ ] Pointers and addresses (2.3.2)
  - AMI does not expose raw pointers or process addresses. The `&` operator is not allowed. Unary `*` is not a dereference; it is used on the left-hand side to mark mutating assignment.
  - Parser rejects pointer/address syntax with `E_PTR_UNSUPPORTED_SYNTAX`.
  - Semantics/linter must not model raw pointers (remaining work below).
- [ ] Concurrency declarations (2.3.6): worker pool sizes, scheduling hints
#### 6.5 Memory Model (Ch. 2.4)

- [ ] Allocation domains: event heap, node‑state, ephemeral stack
- [ ] Ownership & RAII: explicit ownership transfer; deterministic cleanup (scaffold semantics)
- [ ] Per‑VM memory management for runtime execution

Deliverables

- [ ] IR annotations for ownership
- [ ] Lints to prevent leaks and aliasing bugs (Owned<T> params, release/transfer/use‑after‑release; cross‑domain refs)
- [ ] Tests: ownership transfer, RAII lifetime, forbidden cross‑domain references

#### 6.6 Observability (Ch. 1.6)

- [ ] Event‑level telemetry: trace id (eventmeta), telemetry hooks via pragma captured in ASM header
- [ ] Pipeline/node metrics: queue depth, throughput, latency, errors
- [ ] Compiler/runtime hooks to emit diag.v1 compatible records

Deliverables

- [X] JSON debug artifacts (eventmeta), human/JSON logging with --verbose timestamps
- [ ] Tests: presence of telemetry pragma in ASM header; eventmeta schema validation
  - [ ] Metrics emission as diag.v1 JSON lines (pipeline/node)

#### 6.7 Composition and Versioning (Ch. 1.3)

- [ ] Type‑safe composition across packages; Collect() as composition points
- [ ] Compile‑time binding of package versions; conflict detection
- [ ] Capability/trust boundary declarations do not implicitly propagate

Deliverables

- [ ] Compiler checks and errors for cross‑package composition mismatches
- [ ] Tests: version conflict, trust boundary enforcement, explicit transforms for type bridging
- [ ] Config via `ami.workspace` (enable/disable rules, severity)
- [X] Output: human summary and `--json` with file/line/rule/message
- [ ] Exit codes: any error → USER_ERROR; IO issues → SYSTEM_IO_ERROR
- [X] Tests: rule triggers, suppression config, JSON schema, path globs

#### 6.8 Node‑State Table (Ephemeral Key Store)

- [ ] Ephemeral per‑node key/value store (in‑memory; cleared on restart)
- [ ] Namespacing: isolated per pipeline and per node instance
- [ ] API: `put(key,val[,ttl,maxReads])`, `get(key)`, `del(key)`, `has(key)`, `keys()`
- [ ] TTL policies: absolute expiration (duration), optional sliding refresh on access
- [ ] Delete‑on‑read‑count policies: remove after N reads (supports one‑time reads)
- [ ] Concurrency: atomic ops; consistent behavior under concurrent gets/puts
- [ ] Capacity limits: configurable memory cap and eviction policy (LRU when over limit)
 - [ ] Metrics: hits, misses, expirations, evictions, current size; diag.v1 emission hooks
 - [ ] Observability: debug dump guarded by `--verbose` in debug builds; optional emission during `ami test` via flags

Deliverables

 - [ ] Runtime package implementing the ephemeral store with TTL and read‑count deletion
 - [ ] CLI/runtime wiring (scaffold) to obtain per‑node store instances
 - [ ] CLI: `ami build --verbose` emits `kvstore.metrics` and guarded `kvstore.dump` records; `ami test` adds `--kv-metrics` and `--kv-dump` flags
 - [ ] Tester harness integration: kv registry usage, input meta directives (kv_pipeline/node/put/get/emit), helper builders, and auto‑emission toggle
 - [ ] Tests: TTL expiry (absolute and sliding), delete‑on‑read‑count, concurrency, capacity eviction, metrics counters, harness emissions
 - [ ] Test isolation utilities: `kvstore.ResetDefault()` and `tester.ResetDefaultKVForTest(t)`
 - [ ] Docs: usage, guarantees, limitations, and test isolation under `docs/`

#### 6.9 Secure Delete (Zeroization)

- [ ] Policy: define which data are “sensitive” and must be zeroized (e.g., credentials, tokens, PII, secrets)
- [ ] APIs: consistent zeroize hooks in runtime for buffers and values (e.g., `Zeroize([]byte)`; `Handle.Release()` zero‑fills tracked memory)
- [ ] Owned<T>: zeroize underlying memory upon `Close()/Release()` and after ownership transfer (zero original owner’s reference)
- [ ] Memory manager: zeroize any tracked byte buffers on `Release()`; verify idempotency and no panics on double release
- [ ] Node‑state table: zeroize values on `del()`, TTL expiration, capacity eviction, and delete‑on‑read‑count
- [ ] Copies and aliasing: document limitations; mitigate by storing sensitive data in dedicated zeroizable containers
- [ ] Concurrency: atomic zeroization; no data races when readers concurrent with delete/expiry
- [ ] Observability: redact sensitive fields in logs/diag; prevent accidental logging of raw secrets
- [ ] Configuration: opt‑in flags/annotations to mark sensitive fields; optional global policy in `ami.workspace`

Deliverables

- [ ] Runtime utilities for zeroization and safe disposal (byte slices and known container wrappers)
- [ ] Integration: hook zeroization into RAII/Owned<T>, memory manager `Handle.Release()`, node‑state store lifecycle
- [ ] Tests: unit tests verifying memory is overwritten before release (bytes set to 0), double release safety, concurrent delete; property tests for eviction/expiry paths
- [ ] Docs: guidance on limitations of zeroization under GC and compiler optimizations; recommended patterns for handling secrets

#### 6.10 Function Decorators (Python‑Style)

- [X] Syntax: one or more decorator lines immediately above a function
  - `@name` and `@name(arg1, arg2, ...)` forms
  - Args support identifiers, strings, numbers (no full expressions initially)
- [X] Scope: functions only (workers and helpers); not allowed on pipelines/structs/enums
- [X] Application order: bottom‑to‑top (Python semantics)
- [X] Resolution: decorator name resolves to built‑ins or top‑level functions
  - Unknown → `E_DECORATOR_UNDEFINED`
  - Non‑callable or invalid arity → `E_DECORATOR_INVALID`
- [ ] Semantics: decorators must not change a worker’s externally visible signature
  - Worker signature validation preserved; violations → `E_DECORATOR_SIGNATURE`
- [X] Determinism: decorator lists preserved in source order for AST
- [X] Interactions: decorators may add metadata (e.g., `@deprecated("msg")`, `@metrics`)
  - [X] `@deprecated` emits a compile‑time warning diag.v1 with stable fields
- [ ] Config: enable/disable specific decorators via `ami.workspace` (linter/compiler settings)
  - [X] Analyzer scaffold supports disabling decorators (E_DECORATOR_DISABLED); wiring to workspace deferred.
 - [X] AST: attach decorator metadata (name + arg list) to function nodes
 - [X] IR annotations: per‑function decorator list in `ir.v1`
- [X] Codegen (scaffold): allow no‑op or pass‑through; reserve hook points for wrappers

Deliverables

- [X] Parser support: capture decorators on function declarations, with args
- [X] Semantics: resolution, ordering, and error conditions (scaffold: unresolved and conflicting duplicates)
- [X] IR annotations: per‑function decorator list in `ir.v1`
- [X] Tests: parse (single/multiple), ordering, arg parsing
- [ ] Docs: syntax, ordering rules, built‑ins, configuration, and examples

#### 6.11 Enum Conformance (docs/*.docx)

- [ ] Spec alignment: implement enum semantics consistent with `docs/*.docx` (authoritative reference: `docs/Asynchronous Machine Interface.docx`)
- [ ] Canonical names: stable, case‑sensitive canonical string for each member
- [ ] Unique members: no duplicates, no blank identifiers; deterministic ordering
- [ ] Backing ordinal: stable ordinal for each member; reserved unknown sentinel optional
- [ ] Methods:
  - [ ] `String() string`: canonical name
  - [ ] `Ordinal() int`: zero‑based ordinal
  - [ ] `IsValid() bool`: validity check (non‑unknown and within range)
  - [ ] `Values() []<Enum>`: slice of all enum values in canonical order
  - [ ] `Names() []string`: slice of canonical names in canonical order
  - [ ] `Parse<Enum>(s string) (<Enum>, error)`: parse by name (case‑sensitive per docs)
  - [ ] `MustParse<Enum>(s string) <Enum>`: panics on invalid (for tests/tooling)
  - [ ] `FromOrdinal(i int) (<Enum>, error)`: map ordinal to value
  - [ ] JSON: `MarshalJSON()`/`UnmarshalJSON()` to/from canonical string
  - [ ] Text: `MarshalText()`/`UnmarshalText()` to/from canonical string
  - [ ] `GoString() string`: stable debug form `Enum(Name)`
- [ ] Error codes: parsing invalid input → `E_ENUM_PARSE`; invalid ordinal → `E_ENUM_ORDINAL`
- [ ] Determinism: generated tables/maps have stable ordering; no map iteration nondeterminism in public outputs
- [ ] Documentation: define case sensitivity and unknown handling consistent with docs

Deliverables

- [ ] Runtime/stdlib helpers (or codegen) providing the above methods for enums
- [ ] Compiler checks remain: unique members, valid literal values when assigned
- [ ] Tests: string round‑trip, JSON/text round‑trip, parse errors, ordinal mapping, `Values()`/`Names()` ordering, `IsValid()` behavior
- [ ] Docs: `docs/language-enum.md` with examples and guarantees

####  2.1.1) Wrapper and Output
  - [ ] Wraps `go test -json` for provided packages (default `./...`).
  - [ ] Streams `test.v1` JSON events: `run_start`, `test_start`, `test_output`, `test_end`, `run_end`.
  - [ ] Human output: concise per‑test lines and final summary; per‑package lines include cases.
  - [ ] Exit codes: failures → USER_ERROR (1); invocation/build/timeout → SYSTEM_IO_ERROR (2).
  - [ ] Flags: `--timeout`, `--parallel`, `--pkg-parallel`, `--failfast`, `--run`.
  - [ ] Package‑level concurrency: `--pkg-parallel` maps to `go test -p`.
  - [ ] JSON fields:
    - `run_start`: optional `timeout`, `parallel`, `pkg_parallel`, `failfast`, `run`.
    - `run_end`: includes `totals` and `packages[]` (per‑package `{package, pass, fail, skip, cases}`, sorted lexicographically).

#### 2.1.2) Native AMI tests (directive‑driven; parser+sem only)
  - [ ] Discover `*_test.ami` in workspace packages; default root `./src` when workspace missing.
  - [ ] Pragmas define cases and expectations:
    - `#pragma test:case <name>`; `#pragma test:skip <reason>`
    - `#pragma test:expect_no_errors`; `#pragma test:expect_no_warnings`
    - `#pragma test:expect_error <CODE> [count=N] [msg~="substr"]`
    - `#pragma test:expect_warn  <CODE> [count=N] [msg~="substr"]`
    - `#pragma test:expect_errors count=N`; `#pragma test:expect_warnings count=N`
  - [ ] Default case when no pragmas: assert no parser/semantic errors.
  - [ ] Emits `test_start`/`test_end` per case; attaches failing `diag.v1` into `test_end.diagnostics`.
  - [ ] Totals aggregated into global and per‑package summaries; human output prints per‑package counts and cases.

#### 2.1.3) Tests & Docs
  - [ ] CLI tests for flags, JSON stream, exit codes, per‑package summaries, default pattern, package‑level concurrency.
  - [ ] CLI tests for native AMI cases (pass/fail/skip) and rich assertions (count and substring filters).
  - [X] Coverage for `ami test` ≥80%.
  - [X] Docs: `docs/cmd/test.md` updated (flags, JSON fields, pragmas). `docs/runtime-tests.md` will be added during Phase 2 harness work.

Phase 2: Executable AMI tests (scaffolded)

- [ ] `ami/runtime/tester` provides a deterministic runtime harness (scaffold) with simulated execution:
  - Identity output over input payload by default; reserved input keys `sleep_ms` (delay) and `error_code` (force error) enable timeout/error scenarios.
  - Fixtures accepted via `#pragma test:fixture path=<rel> mode=<ro|rw>` (validated; enforcement deferred).
- [ ] Integrate harness with `ami test` via `#pragma test:runtime ...` directives (cases discovered, executed, and reported as `test.v1`).
- [ ] Support input/output JSON equality assertions, error assertions, and per‑case timeouts.
- [ ] Tests: CLI runtime cases (JSON equality, error, timeout) and unit tests for pragma parsing and harness behavior.

#### 2.1.3) Compiler (separate package; custom parser like Go)

- [ ] Create a standalone compiler library under `src/ami/compiler/` composed of cohesive subpackages (no CLI deps):
  - [X] `token`: token kinds, literals, operator precedences
  - [X] `scanner`: UTF‑8 reader, rune decoding, comment handling, tokenization (like Go’s scanner)
  - [X] `ast`: typed AST nodes, positions, comments (ImportDecl, FuncDecl scaffold)
  - [X] `parser`: parse package, imports (single-line), and empty function decls
  - [ ] `types`: symbol tables, scopes, basic type definitions (inference/checking deferred)
  - [ ] `sem`: semantic analysis (basic: duplicate function detection)
  - [ ] `ir`: lowered intermediate representation scaffold (stable orderig)
  - [ ] `codegen`: assembly-like text generation for debug artifacts
  - [ ] `diag`: diagnostics with position info and machine‑readable schema conversion
  - [X] `source`: file set management (file → offsets → line/col)
- [X] Public driver package `src/ami/compiler/driver` exposing `Compile(workspace, pkgs, opts) (artifacts, diagnostics)` (scaffold)
- [X] Grammar: document EBNF in `docs/` and unit‑test parser with fixtures
- [X] Error handling: Go‑style tolerant parsing (synchronize at semicolons/keywords), collect multiple errors
- [ ] Determinism: stable symbol iteration, stable IR output for golden tests
- [X] Tests (happy/sad path) for each subpackage; basic golden tests for parser/IR/codegen

Parser Enhancements (Positions & Comments)

- [X] Attach positions (`pos.line/column/offset`) to: directives, imports, funcs, enums, structs, pipelines, and node calls.
- [X] Capture and attach leading comments (`//`, `/*...*/`) to the following node; scanner preserves comment text and start position.
- [X] Tests cover presence of positions and attached comments on representative nodes (directive, import, enum, struct, func, pipeline step).
- [X] Extend positions to function-body statements and expressions: `ExprStmt`, `AssignStmt`, `DeferStmt`, and expression nodes (`CallExpr`, `SelectorExpr`, `UnaryExpr`, `Ident`, `BasicLit`).
- [X] Attach comments to function-body statements (top-level already covered).
- [X] Add binary expressions with precedence/associativity for arithmetic and comparisons (scaffold).
- [X] Add `return` statements to function bodies.
 - [X] Add local variable declarations: `var IDENT [Type] [= expr]` in function bodies.

Types & Semantics (incremental)

 - [X] Composite types and AST→types mapping: `Pointer`, `Slice`, `Map`, `Set`, `slice<T>`, `Event<T>`, `Error<E>`.
   - [X] Textual mapping scaffold: `types.Parse()` and `types.MustParse()` for primitives and generics (`slice<T>`, `set<T>`, `map<K,V>`, `Event<T>`, `Error<E>`); tests added.
- [X] Function signatures: `types.Function{Params, Results}` built from `FuncDecl` for scope/type introspection.
- [X] Import symbol scope: insert alias or last path segment into top‑level scope (kind `ObjType`, type `package`).
- [X] Tests: verify type mapping and inferred function signatures; scope contains imported package symbol.
- [X] Owned<T>: added to the type mapper; string rendering `Owned<…>`.
- [ ] RAII + Defer: semantic analyzer recognizes `defer`-scheduled releases/transfers and counts them toward required Owned<T> cleanup at function end; flags double-release when mixed with immediate release.
 - [X] Worker resolution across imports: dotted references like `pkg.Func()` accepted when `pkg` is imported; undefined worker diagnostics suppressed (signature checks across packages deferred).
- [ ] Type inference/unification across expressions and generic instantiation inside bodies (future)
  - Goals
    - [ ] Infer local expression types (identifiers, literals, unary/binary ops, calls) without explicit annotations where possible.
    - [ ] Instantiate generics from usage (e.g., infer `T` in `Event<T>`, `Error<E>`, `slice<T>`, `set<T>`, `map<K,V>` from call sites and argument/assignment contexts).
    - [ ] Preserve determinism and simple, predictable rules; avoid global inference surprises.
  - Scope (initial)
    - [ ] Intra‑function inference only (within a single body); cross‑function inference relies on declared signatures.
    - [ ] Unify arithmetic and comparison operands; check operator applicability and produce clear diagnostics on mismatch. (Implemented basic token-based checks for +,-,*,/,% and ==,!=,<,<=,>,>= with E_TYPE_MISMATCH.)
    - [ ] Function call instantiation: infer generic arguments from parameter→argument constraints; support tuple returns when present.
      - [ ] Local call instantiation for single-letter type variables (e.g., Event<T>, Owned<T>); no tuple returns yet.
      - [ ] Instantiate return type from call-site constraints for single-result functions (used in return/assignment checks).
      - [ ] Tuple returns: instantiate callee multi-result types at return sites when returning a call expression.
    - [ ] Container/generic propagation: infer element/key types for `slice<T>`, `set<T>`, and `map<K,V>` based on construction and assignment sites.
      - [ ] Literal-based inference for `slice{...}`, `set{...}`, `map{...}` without explicit type args; unify with assignment/return types.
      - [ ] Assignment unification for containers (element/key/value type checks).
    - [ ] Event/Error propagation: infer payload types for `Event<T>` / `Error<E>` flows through transforms and pipeline steps (within the file/local scope as a start).
    - [ ] Exclusions (initial): no overloading, no implicit conversions beyond documented numeric/string literal rules, no cross‑package global inference.
  - Diagnostics
    - [ ] `E_TYPE_MISMATCH`: incompatible operand or argument types (e.g., binary op or assignment mismatch). (Implemented for arithmetic/comparison; covered by tests.)
    - [ ] `E_TYPE_AMBIGUOUS`: insufficient constraints to determine a concrete type variable.
    - [ ] `E_TYPE_UNINFERRED`: remaining type variables at an emission point (e.g., return) that require explicit annotation.
    - [ ] `E_RETURN_TYPE_MISMATCH`: return expression type does not match declared result type (added).
    - [X] `E_CALL_ARG_TYPE_MISMATCH`: function call argument incompatible with parameter type (scaffold).
    - [X] `E_CALL_ARITY_MISMATCH`: function call arity differs from callee signature (scaffold).
  - Tests
    - [ ] Happy: infer `T` in `Event<T>` from a call site, propagate through assignment and return; infer container element types from literals/usages.
    - [ ] Happy: call-site instantiation for `Event<T>`/`Owned<T>` and assignment unification for generics (no return yet).
    - [ ] Happy: infer `K,V` for `map<K,V>` from put/get/assignment contexts; verify deterministic concrete types.
    - [ ] Happy: arithmetic/comparison type checks succeed with compatible operands and fail otherwise.
    - [ ] Happy: tuple-return instantiation at return site for local generic functions.
  - [ ] Sad: ambiguous inference produces `E_TYPE_AMBIGUOUS`; conflicting constraints produce `E_TYPE_MISMATCH`; missing concrete at return produces `E_TYPE_UNINFERRED`.

### Remaining work

- [X] Surface import version constraints into `sources.v1` (`importsDetailed`) during build planning/output.

-- [ ] Grammar extensions: unary/binary operators (tracking for backend ops already supported)
  - [X] Add tokens and precedence for bitwise (`&`, `|`, `^`), shifts (`<<`, `>>`).
  - [X] Introduce `ast.UnaryExpr` and extend parsing to support unary logical `!`.
  - [X] Lowering: map unary `!` to IR op `not`; tests added.
  - [X] Lowering: map `|`→`bor`, `&`→`band`, `^`→`xor`, `<<`→`shl`, `>>`→`shr`; tests added.
  - [X] LLVM emission: handle `bor` (→ `or` iN), `band` (→ `and` iN), and `xor`, `shl`, `shr` for integers; keep logical `and`/`or` for booleans only.

 

Type System and Semantics (Phase 2.1)

- [X] Types: add composite types and mapping from AST
  - [X] `types.Pointer`, `types.Slice`, `types.Map`, `types.Set`, `types.SliceTy`
  - [X] Mapper `types.FromAST(ast.TypeRef) Type` handles generics (`Event<T>`, `Error<E>`, `Owned<T>`, `map<K,V>`, `set<T>`, `slice<T>`) and pointer/slice forms (`*T`, `[]T`)
  - [X] Function signatures: `types.Function{Params, Results}` with string rendering; tests added
- [C] Semantics: function type inference and import symbol scope
  - [X] Build `types.Function` from `FuncDecl` params/results and insert into top scope as `ObjFunc` (scaffold via helper + tests)
  - [X] Insert imported package symbols into scope (alias or last path segment) as `ObjType` of `types.TPackage` (scaffold via helper + tests)
  - [X] Tests: verify inferred signature and import symbol resolution
- [ ] Types: unification/inference across expressions and generics within bodies
  - [ ] Call-argument type checking + unification for single-letter generics (e.g., `Event<T>`, `Owned<T>`), intra-function only.
  - [ ] Assignment unification for generics within bodies.
  - [ ] Arithmetic/comparison operand checks with `E_TYPE_MISMATCH` diagnostics.
  - [X] Return-site checks: `return` expressions unified against declared results; emits `E_RETURN_TYPE_MISMATCH` (scaffold).
  - [X] Local variable inference from initializer expressions; locals participate in call-arg checks.
  - [ ] Broader local expression inference (idents across scopes), return-site inference without annotations, and container propagation.
- [ ] Semantics: cross‑package name resolution (multi‑file), constant evaluation, additional validation rules
- [X] IR: lower imperative subset and full pipeline semantics with typed annotations
- [X] Codegen: emit real object files (beyond debug ASM) and integrate into build plan




#### 8.3) Examples Build Target

- [X] Makefile `examples` target builds all workspaces under `examples/*` using the CLI and stages outputs under `build/examples/<name>/`.
- [X] Each `examples/**/ami.workspace` includes a cross‑platform `toolchain.compiler.env` matrix (windows/amd64, linux/amd64, linux/arm64, darwin/amd64, darwin/arm64).

#### 8.2) Non‑Debug Build Artifacts

- [X] Emit per‑unit assembly under `build/obj/<package>/<unit>.s` for normal builds (no `--verbose`).
- [X] Per‑package index `build/obj/<package>/index.json` (`objindex.v1`) lists `{ unit, path, size, sha256 }`.
- [X] Determinism: indexes and obj asm are stable across runs; tests cover single and multi‑package scenarios (timestamp normalized in tests).
- [ ] `ami.manifest` includes these entries as `kind:"obj"` when present.

Edges Runtime Scaffolding (for harness/tests)

- [ ] Provide `push(e Event)` and `pop() Event` methods for `edge.FIFO`, `edge.LIFO`, and `edge.Pipeline` with bounded capacity and backpressure semantics:
  - [X] FIFO/LIFO runtime buffers with `Push/Pop` and policies: `block` → `ErrFull`; `dropOldest`/`dropNewest` implemented (shunt treated as drop).
  - [X] Pipeline runtime buffer (scaffold analogous to FIFO) with `Push/Pop`.
- [ ] Thread‑safe counters and tests for order (FIFO/LIFO), backpressure handling, and simple concurrency.
  - [X] Order/backpressure tests for FIFO/LIFO.
  - [X] Simple concurrency tests (single producer/consumer) for FIFO/LIFO.

### 9) Determinism and Non‑Interactive Operation

- [ ] All subcommands avoid prompts (use flags for destructive actions)
- [ ] Ensure stable ordering of listings/output (sorted)
- [ ] Ensure locale‑agnostic formatting (timestamps ISO‑8601 UTC)
- [ ] Tests: verify no TTY prompts, stable outputs in golden tests
 - [ ] Ensure reproducibility: fixed ordering for maps/sets, stable timestamps in logs (ISO‑8601 UTC), stable file layouts; golden tests for AST/IR serializers

### 10) Documentation and Help

- [ ] `ami help` and per‑command help with examples
 - [X] Examples README documents how to build/run sample workspaces and use `make examples`.
- [ ] Manpage/Markdown docs generated to `./build/docs` (optional)
- [ ] Reference Chapter 3.0 commands and flags in docs
- [ ] Tests: help text includes flags, exit codes, examples
- [ ] Add `version` command with semantically versioned CLI; `--json` returns `{ "version": "vX.Y.Z" }`
- [ ] Document output schemas in prose under `docs/` (reference only). Machine schemas are implemented as Go types in `src/schemas/` with unit tests (one struct per file with matching `_test.go`).
 - [ ] Language docs updated to canonical worker signature; removed legacy State‑parameter suggestions.
 - [ ] Defer examples updated to canonical return tuple `(Event<U>, error)`.
 - [ ] Lint docs updated: legacy worker signatures emit `E_WORKER_SIGNATURE` (no ambient‑suggest info).
## 1.3.0.0. JSON Schemas (Stable Output Contracts)
- All JSON outputs MUST conform to the following schemas. Emit objects with stable key ordering. 
  Unknown fields MUST NOT be added without a schema version bump. Schemas are implemented as Go types 
  under `src/schemas/` (see “Schema Implementation”).
### 1.3.1.0. Diagnostics Schema (errors, warnings, info)
- [x] Implemented `diag.v1` Go types under `src/schemas/diag` with unit tests (deterministic JSON ordering).
- Version: `diag.v1`
- Object fields:
  - `schema`: string, constant `"diag.v1"`
  - `timestamp`: string, ISO‑8601 UTC, millisecond precision
  - `level`: string, one of `"info" | "warn" | "error"`
  - `code`: string, machine code for the diagnostic (e.g., `"E1001"`)
  - `message`: string, human‑readable
  - `package`: string, optional (import path)
  - `file`: string, optional (workspace‑relative path)
  - `pos`: object, optional: `{ line: number, column: number, offset: number }`
  - `endPos`: object, optional, same shape as `pos`
  - `data`: object, optional (machine‑readable context)
> Example:
> { 
>   "schema":"diag.v1", 
>   "timestamp":"2025-09-24T17:05:06.123Z", 
>   "level":"error", 
>   "code":"E1001",
>   "message":"unexpected token ";"", 
>   "package":"example/app", 
>   "file":"pkg/main.ami",
>   "pos":{ 
>       "line":12, 
>       "column":17, 
>       "offset":214 
>   }, 
>   "data":{ 
>       "got":"SEMI", 
>       "want":["IDENT","RBRACE"] 
>   }
> }

Canonical diagnostic codes (this phase):
- `E_WS_SCHEMA`: workspace schema validation failed during `ami build`.
- `E_INTEGRITY`: cache vs `ami.sum` integrity mismatch during `ami build`.
- `E_INTEGRITY_MANIFEST`: existing `ami.manifest` vs `ami.sum` mismatch before build.
### 1.3.2.0. Test Stream Schema (JSON Lines)
- Version: `test.v1`
- Every line is a JSON object with `schema":"test.v1"` and a `type`:
  - `run_start`: `{ schema, type:"run_start", timestamp, workspace, packages:[string] }`
  - `test_start`: `{ schema, type:"test_start", timestamp, package, name }`
  - `test_output`: `{ schema, type:"test_output", timestamp, package, name, stream:"stdout|stderr", text }`
  - `test_end`: `{ schema, type:"test_end", timestamp, package, name, status:"pass|fail|skip", duration_ms:number, diagnostics?: [Diagnostics] }`
  - `run_end`: `{ schema, type:"run_end", timestamp, duration_ms:number, totals:{ pass:number, fail:number, skip:number, cases:number } }`
> Example line:
> { 
>   "schema":"test.v1", 
>   "type":"test_end", 
>   "timestamp":"2025-09-24T17:05:09.321Z", 
>   "package":"example/app", 
>   "name":"TestFoo", 
>   "status":"fail", 
>   "duration_ms":42, 
>   "diagnostics": [ 
>       {
>           "schema":"diag.v1", 
>           "level":"error", 
>           "code":"E2001", 
>           "message":"assertion failed"
>       } 
>   ] 
> }
### 1.3.3.0. Build Plan Schema
- Version: `buildplan.v1`
- Object fields:
  - `schema`: `"buildplan.v1"`
  - `timestamp`: ISO‑8601 UTC
  - `workspace`: string (absolute or workspace‑relative path)
  - `toolchain`: object `{ amiVersion:string, goVersion:string }`
  - `targets`: array of target objects:
    - `{ package:string, unit:string, inputs:[string], outputs:[string], steps:[string] }`
  - `graph`: object with dependency edges `{ nodes:[string], edges:[ [from:string, to:string] ] }` (optional)

>Example:
>{ 
>   "schema":"buildplan.v1", 
> "timestamp":"2025-09-24T17:05:07.000Z", 
> "workspace":".",
>   "toolchain":{
> "amiVersion":"v0.1.0",
> "goVersion":"1.25"
> },
>   "targets":[ 
>       {
>           "package":"example/app",
>           "unit":"pkg/main.ami",
>           "inputs":[
>               "pkg/main.ami"
>           ],
>           "outputs":[
>               "build/obj/example/app/main.o"
>           ],
>           "steps":[
>               "scan",
>               "parse",
>               "typecheck",
>               "ir",
>               "codegen"
>           ]
>       } 
>   ] 
> }
### 1.3.4.0. Debug Artifacts Schemas (build --verbose)
- Resolved Source Tree (`build/debug/source/resolved.json`): `sources.v1`
  > ```
  > {
  >         schema:"sources.v1", 
  >         timestamp, 
  >         units:[ 
  >             { 
  >                 package:string, 
  >                 file:string, 
  >                 imports:[string], 
  >                 source:string
  >             } 
  >         ] 
  > }
  > ```
- AST per unit (`*.ast.json`): `ast.v1`
  > ```
  > {
  >     schema:"ast.v1", 
  >     timestamp, 
  >     package:string, 
  >     file:string, 
  >     root: Node 
  > }
  > ```
  > `Node`:
  > ```
  > { 
  >     kind:string, 
  >     pos:{ line, column, offset }, 
  >     endPos?:{ ... }, 
  >     fields?:object, 
  >     children?:[Node] 
  > }
  > ```
- IR per unit (`*.ir.json`): `ir.v1`
  > ```
  > { 
  >     schema:"ir.v1", 
  >     timestamp: uint64, 
  >     package:string, 
  >     file:string,
  >     functions:[ 
  >        { 
  >           name:string, 
  >           blocks:[ 
  >               { 
  >                   label:string, 
  >                   instrs:[ 
  >                        { 
  >                             op:string, 
  >                             args:[string|number|object], 
  >                             result?:string
  >                         } 
  >                   ] 
  >                } 
  >           ] 
  >        } 
  >     ] 
  > }
  > ```
- [X] Assembly listing (`*.s`): text; additionally, emit an index JSON `asm.v1` per package
  > ```
  > { 
  >     schema:"asm.v1", 
  >     timestamp, 
  >     package:string, 
  >     files:[ 
  >         { 
  >             unit:string, 
  >             path:string, 
  >             size:number, 
  >             sha256:string 
  >         } 
  >     ] 
  > }
  > ```

Validation:
- Implement JSON Schema tests (or structural validation) to ensure outputs conform to these contracts.
- All timestamps in these artifacts must be ISO‑8601 UTC with milliseconds.
### 1.3.5.0. Non‑Debug Build Artifacts Schema
- Obj Index (`build/obj/<package>/index.json`): `objindex.v1`
  >```
  >{ 
  >    schema:"objindex.v1", 
  >    timestamp:uint64, 
  >    package:string, 
  >    files:[ 
  >        { 
  >            unit:string, 
  >            path:string,
  >            size:number,
  >            sha256:string
  >        } 
  >    ] 
  >}
  >```
  - Determinism: tests normalize `timestamp` and compare for equality across runs.
## 2.0.0.0 AMI Standard Library (stdlib)
> Goal: Provide a first‑class AMI standard library modeled after Go’s stdlib where it aligns with POP/AMI constraints 
> (determinism, explicit I/O, stable outputs). Code lives under `src/ami/stdlib/<pkg>` with `_test.go` colocated. 
> Docs in `docs/stdlib/`.
### 2.0.0.1. Integration with AMI
- [ ] Import path: `import "ami/stdlib/<pkg>"` from `.ami`
- [ ] Lints: forbid hidden I/O in restricted nodes; only explicit `os/*` calls allowed where policy permits
- [ ] Golden tests for determinism (JSON stable keys, ISO‑8601 UTC, filepath normalization)
### 2.1.0.0. Scope (Phase 1)
#### 2.1.1.0. Layout and Conventions
  - [ ] Create `src/ami/stdlib/` with subpackages below
  - [ ] Add scaffolding for all Phase 1 packages (`doc.go`, `ready.go`, `_test.go`); builds and vets clean
  - [ ] Per‑package docs in `docs/stdlib/<pkg>.md` (stubs)
  - [ ] Determinism policy documented (no ambient time/random; explicit injection) — see `docs/stdlib/determinism.md`
  - [ ] ≥80% coverage per package; happy/sad tests
#### 2.1.2.0. Core Text/Data
  - [ ] Package: `strings`: 
    - [ ] Contains()
    - [ ] HasPrefix()
    - [ ] HasSuffix()
    - [ ] Split()
    - [ ] Join()
    - [ ] Replace()
    - [ ] Trim()
    - [ ] TrimSpace()
    - [ ] ToLower()
    - [ ] ToUpper()
    - [ ] Fields()
    - [ ] Index()
    - [ ] LastIndex() 
    - [ ] EqualFold()
  - [ ] Package: `bytes`:
    - [ ] Contains()
    - [ ] Compare()
    - [ ] Index()
    - [ ] LastIndex()
    - [ ] Split()
    - [ ] Join()
    - [ ] Replace()
  - [ ] Package: `regexp` (subset): 
    - [ ] Compile()
    - [ ] MustCompile()
    - [ ] MatchString(); safe flags only; denial‑by‑default for catastrophic patterns (tests)
#### 2.1.3.0. Time 
  - [ ] `time`: Parse/Format ISO‑8601 UTC; Duration parse/format; Now via injected Clock only (no globals)
#### 2.1.4.0. Math 
> Goal: Provide a deterministic math package `ami/stdlib/math` with a well‑defined, cross‑platform set of numeric 
> constants and functions. Focus on correctness, stability across OS/arch, and clear NaN/Inf behaviors.
  - [ ] `math`: Baseline deterministic functions 
  - [ ] `rand`: Deterministic PRNG with explicit seed handle (no package‑level state)
  - [ ] Constants
    - [ ] `Pi`,
    - [ ] `E`,
    - [ ] `Phi`,
    - [ ] `Sqrt2`,
    - [ ] `SqrtE`,
    - [ ] `Ln2`,
    - [ ] `Ln10`,
    - [ ] `Log2E`,
    - [ ] `Log10E`
  - [ ] Predicates & sign
    - [ ] `IsNaN(x)`,
    - [ ] `IsInf(x, sign)`,
    - [ ] `Signbit(x)`,
    - [ ] `IsFinite(x)`
  - [ ] Basic ops
    - [ ] `Abs(x)`,
    - [ ] `Copysign(x, y)`,
    - [ ] `Max(x, y)`,
    - [ ] `Min(x, y)`
  - [ ] Rounding & decomposition
    - [ ] `Ceil(x)`,
    - [ ] `Floor(x)`,
    - [ ] `Trunc(x)`,
    - [ ] `Round(x)`,
    - [ ] `RoundToEven(x)`
    - [ ] `Frexp(x)`,
    - [ ] `Ldexp(frac, exp)`,
    - [ ] `Modf(x)`
  - [ ] Powers, roots, logs
    - [ ] `Sqrt(x)`,
    - [ ] `Pow(x, y)`
    - [ ] `Exp(x)`,
    - [ ] `Exp2(x)`,
    - [ ] `Expm1(x)`
    - [ ] `Log(x)`,
    - [ ] `Log10(x)`,
    - [ ] `Log2(x)`,
    - [ ] `Log1p(x)`
  - [ ] Trigonometry
    - [ ] `Sin(x)`,
    - [ ] `Cos(x)`,
    - [ ] `Tan(x)`
    - [ ] `Asin(x)`,
    - [ ] `Acos(x)`,
    - [ ] `Atan(x)`,
    - [ ] `Atan2(y, x)`
  - [ ] Hyperbolic (core subset)
    - [ ] `Sinh(x)`,
    - [ ] `Cosh(x)`,
    - [ ] `Tanh(x)`
  - [ ] Misc
    - [ ] `Hypot(x, y)`,
    - [ ] `Remainder(x, y)`,
    - [ ] `Mod(x, y)`
- Determinism & Semantics
  - [ ] Cross‑platform determinism policy documented (IEEE‑754 double, stable results across OS/arch)
  - [ ] Pure Go implementations or stabilized wrappers to ensure ≤1 ulp variance; tests enforce error bounds
  - [ ] Defined NaN/Inf propagation and domain behavior (e.g., `Sqrt(<0)` returns NaN)
  - [ ] No reliance on CPU‑specific math flags or non‑deterministic intrinsics
#### 2.1.5.0. Package io
##### 2.1.5.1. Filesystem/I/O (restricted, explicit)
  - [ ] `path/filepath`: Clean, Join, Base/Dir, Ext (no symlink eval)
  - [ ] `os`: ReadFile/WriteFile (explicit perms), Mkdir/MkdirAll, Stat; no env/process mutation
  - [ ] `io`: Bounded copy primitives (CopyN), Reader/Writer helpers
  - [ ] `bufio`: Reader/Writer with explicit buffer sizes; documented flush semantics
##### 2.1.5.2. Networking (stubs only)
###### package: `net/http` adapters deferred (external I/O; non‑deterministic)
> Goal: Provide deterministic HTTP/1.1 and HTTP/2 client/server adapters under `ami/stdlib/http` suitable 
>       for AMI pipelines, with explicit configuration, no global state, and stable behaviors across platforms.
- Types
  - [ ] `Request`:
    - Fields: 
      - `Method`, 
      - `URL`, 
      - `Headers map[string][]string` (canonicalized), 
      - `Body []byte`, 
      - `TimeoutMs`, 
      - `Proto` (`"h1"|"h2"`), 
      - `TLS *TLSConfig`, 
      - `FollowRedirects bool`, 
      - `MaxRedirects`, 
      - `Idempotent bool`
  - [ ] `Response`:
    - Fields: 
      - `StatusCode`, 
      - `Headers map[string][]string`, 
      - `Body []byte`, 
      - `Proto` (`"h1"|"h2"`), 
      - `DurationMs`, 
      - `TLS *TLSState`
  - [ ] `Client` 
    - config:
      - `MaxConnsPerHost`, 
      - `MaxIdleConns`, 
      - `IdleConnTimeoutMs`, 
      - `DisableCompression`,
      - `HTTP2 bool`,
      - `Retry Policy` (attempts/backoff), 
      - `Proxy` (deferred)
        - [ ] `Server` config (minimal):
            - `Addr`, `TLS *TLSConfig`, `ReadTimeoutMs`, `WriteTimeoutMs`, `IdleTimeoutMs`, `HTTP2 bool`
        - [ ] `TLSConfig` (subset): 
          > CA roots, serverName, client cert/key (mTLS), min/max version (≥1.2), cipher suites (sane defaults), cert pinning (SHA‑256 SPKI)
- Pipelines
  - [ ] Client pipeline: `Ingress(cfg).Transform(worker=encode).Egress(worker=httpClient)`
      - Input: `Event<Request>`; Output: `Event<Response>` or `Error<NetError>`
      - Features: batching disabled by default; backpressure policy selectable; retries for idempotent requests only (configurable)
  - [ ] Server pipeline: `Ingress(httpServer).Transform(worker=parse).Fanout(routes...).Egress(worker=writeResp)`
      - Ingress produces `Event<Request>`; handlers produce `Event<Response>`; errors route to error pipeline
      - Router: simple path/method matching; middleware hooks optional (future)
- Determinism & Policy
  - [ ] No global client/server; all state injected via config
  - [ ] Timeouts explicit; default values documented; clocks are system but only used for deadline enforcement
  - [ ] Redirects: disabled by default; when enabled, capped by `MaxRedirects`, headers/body forwarding rules documented
  - [ ] Compression: disabled by default (or deterministic gzip only); stable header ordering and canonicalization
  - [ ] Retries: exponential backoff without jitter (deterministic), idempotent methods only unless `Idempotent` set
  - [ ] HTTP/2: ALPN negotiation; server push disabled; header compression (HPACK) behavior is opaque but responses normalized to `Response`
  - [ ] I/O restrictions: only allowed in ingress/egress nodes per static checks; transforms remain pure
- Docs & Tests
  - [ ] `docs/stdlib/http.md` with configuration matrix and examples
  - [ ] Unit tests:
      - [ ] Client H1: `httptest.Server` exercising methods, headers, timeouts, redirects, TLS
      - [ ] Client H2: `httptest` with HTTP/2 (ALPN); verify upgrade and frame‑opaque behavior
      - [ ] Server H1: request parsing, routing, response writing, error pipeline wiring
      - [ ] TLS tests: self‑signed CA, mTLS, cert pinning
      - [ ] Retry policy tests: idempotent vs non‑idempotent, backoff schedule determinism
  - [ ] Coverage ≥80%
- Out of Scope (Phase 1)
  - [ ] HTTP/3/QUIC, WebSockets, HTTP/2 server push
  - [ ] Cookie jar, caching layers, proxy authentication
  - [ ] Transparent content decoding beyond optional gzip
###### package: `net/raw` - raw socket
###### package: `net/dns` - dns client / server
> Goal: Provide a deterministic DNS client under `ami/stdlib/dns` for issuing DNS queries over UDP/TCP (port 53) and 
> DNS‑over‑TLS (DoT, tcp/853). Focus on correctness, deterministic transaction IDs (configurable), 
> explicit timeouts/retries, and stable record ordering.
- Types
  - [ ] `Request`:
      - Fields: 
        - `Name` (FQDN), 
        - `Type` (`A|AAAA|CNAME|TXT|MX|SRV|NS|PTR|CAA`), 
        - `Class` (`IN`), 
        - `RecursionDesired bool`,
        - `EDNS0 bool` (with `UDPSize`), 
        - `DNSSECOK bool`, 
        - `Server string` (host:port),
        - `Network` (`udp|tcp|dot`),
        - `TimeoutMs`, 
        - `Retries`, 
        - `TXIDMode` (`hash|seeded`), 
        - `Seed uint64` (when `seeded`)
  - [ ] `Response`:
      - Fields: `RCode`,
        - `Truncated bool`,
        - `Answers []RR`, 
        - `Authority []RR`, 
        - `Additional []RR`, 
        - `Proto` (`udp|tcp|dot`), 
        - `DurationMs`
  - [ ] `RR` (resource record union):
      - Common: `Name`, `TTL`, `Type`, `Class`
      - Subtypes (oneOf): 
        - `A {IP string}`, 
        - `AAAA {IP string}`,
        - `CNAME {Target string}`,
        - `TXT {Text []string}`, 
        - `MX {Pref uint16, Host string}`, 
        - `SRV {Prio, Weight, Port uint16, Target string}`, 
        - `NS {Host string}`, 
        - `PTR {Host string}`, 
        - `CAA {Flag uint8, Tag string, Value string}`
  - Pipeline
    - [ ] Client pipeline: `Ingress(cfg).Transform(worker=encode).Egress(worker=dnsClient)`
        - Input: `Event<Request>`; Output: `Event<Response>` or `Error<DNSError>`
        - Behavior: 
          - UDP by default; 
          - auto‑fallback to TCP on truncation; 
          - DoT when `Network=dot` with TLS v1.2+, SNI, optional cert pinning
- Determinism & Policy
  - [ ] TXID derivation:
      - `hash`: 16‑bit TXID = low bits of SHA‑256(Name|Type|Class|server) (deterministic)
      - `seeded`: 16‑bit TXID from a PRNG seeded via `Seed` (no global RNG)
  - [ ] Stable ordering of RRs in each section (lexicographic by Name, Type, RDATA) to ensure reproducible output
  - [ ] Explicit timeouts and retry strategy (exponential backoff without jitter); retry only when safe (timeout, truncation after fallback)
  - [ ] I/O restrictions: only usable in ingress/egress nodes; transforms remain pure
- Docs & Tests
  - [ ] `docs/stdlib/dns.md` describing configuration, RR formats, and examples
  - [ ] Unit tests:
      - [ ] UDP/TCP echo servers generating canned DNS responses (A/AAAA/CNAME/TXT/MX/SRV/NS/PTR/CAA)
      - [ ] Truncation forcing TCP fallback
      - [ ] DoT with self‑signed CA and cert pinning
      - [ ] Deterministic TXID tests for `hash` and `seeded` modes
      - [ ] Stable RR ordering
  - [ ] Coverage ≥80%
- Out of Scope (Phase 1)
  - [ ] Recursive resolution, caching layers, negative caching
  - [ ] EDNS Client Subnet (ECS), DNS‑over‑HTTPS (DoH)
  - [ ] Zone transfers (AXFR/IXFR)
###### package: `net/smtp` - SMTP email client/server (interfaces + stubs initially)
- [ ] Package should support POP3 and POP3/s connections
  - Fetch mail from mailbox
- [ ] Package should support IMAP and IMAP/s connections
  - Fetch mail from mailbox
  - Synchronize local and remote mailboxes
- [ ] Package should support SMTP and SMTP/s connections
  - Send email
  - Receive email from peer SMTP servers
###### package: `net/amqp` - AMQP client/server (interfaces + stubs initially)
- [ ] Package should support connecting to AMQP/S server
- [ ] Package should support sending and receiving messages
###### package: `net/ssh` - SSH Client / server (interfaces + stubs initially)
- [ ] Package must run ssh server
- [ ] Package must run ssh client
- [ ] Package must support ssh key management functions
#### 2.1.6.0. Crypto
##### 2.1.6.1. Encoding/Hashing
  - [ ] `encoding/json`: Deterministic marshal (sorted keys), strict unmarshal option
  - [ ] `encoding/yaml`: Deterministic marshal (sorted keys optionally), strict unmarshal option
  - [ ] `encoding/bas64`:  (interfaces + stubs initially)
  - [ ] `crypto/sha256`: Sum256 + streaming helper
    - [ ] SHA‑256: `Sum(data)`, `New() io.Writer`, `HMAC(key, data)`; RFC test vectors

##### 2.1.6.2. Encryption / Key exchange
  - [ ] `crypto/aes` (interfaces + stubs initially)
  - [ ] `crypto/rsa` (interfaces + stubs initially)
  - [ ] `crypto/ecc` (interfaces + stubs initially)
  - [ ] `crypto/kyber` (interfaces + stubs initially)
  - [ ] `crypto/dilithium` (interfaces + stubs initially)
  - [ ] `crypto/gpg` (interfaces + stubs initially)
  - [ ] `crypto/diffie-hellman` (interfaces + stubs initially)
  - [ ] `crypto/key-creation` (interfaces + stubs initially)
  - [ ] `crypto/key-signing` (interfaces + stubs initially)
  - [ ] `crypto/hmac` (interfaces + stubs initially)
##### 2.1.6.3. Compression
  - [ ] `compress/gzip` (interfaces + stubs initially)
  - [ ] `compress/crsce` (interfaces + stubs initially)
#### 2.1.7.0. logger
> Goal: Provide a deterministic, configurable logging package `ami/stdlib/logger` that exposes composable AMI pipelines to write logs to multiple sinks. Each sink is implemented as an AMI pipeline with explicit configuration, bounded buffers, backpressure policy, retries, and observable counters.
##### 2.1.7.1. Core Concepts
- [x] Record format: `log.v1` JSON with fields `{ timestamp (ISO‑8601 UTC, ms), level, message, fields (map), package, pipeline?, node? }` (stable ordering on marshaling) — implemented under `src/schemas/log` with tests
- [ ] Formatter interface: text and JSON formatters; JSON is deterministic (sorted keys)
- [ ] Sink interface: Start() (establish resources), Write([]Record) error (batch), Close()
- [ ] Pipeline templates: Ingress(cfg).Transform(worker=format).Egress(worker=sink) with bounded edges and selectable backpressure (`dropOldest`/`dropNewest` or `block`).
- [ ] Levels: trace, debug, info, warn, error, fatal (string)
##### 2.1.8.0. Sinks (Pipelines)
- [x] Skeleton sinks present under `src/ami/stdlib/logger`: stdout, stderr, file (Phase 1; pipelines/batching/backpressure pending)
- [ ] `stdout` pipeline: writes to stdout; supports text and JSON; no colors when in JSON mode
- [ ] `stderr` pipeline: writes to stderr; same options as stdout
- [ ] `file` pipeline: appends to a file path; options: `path`, `perm` (octal), `maxSize` (bytes, optional noop in Phase 1), `flushInterval` ms; atomic writes per line
- [ ] `syslog` pipeline (tcp/514): RFC‑5424 emission; options: `addr`, `facility`, `appName`, `hostname`, `tls` (optional); reconnect with backoff
- [ ] `https` pipeline (tcp/443 via POST): options: `url`, `headers` (map), `timeout` ms, `batchMax`/`batchInterval` ms, TLS config; sends JSON array or NDJSON per request
- [ ] `amqps` pipeline: options: `url`, `exchange`, `routingKey`, `queue` (optional), `durable`, `confirm`, TLS config; publish with confirms and retry
##### 2.1.8.2 Configuration & Behavior
- [ ] Batching: configurable `batchMax` and `batchInterval`; batch flush on size/time
- [ ] Retries: exponential backoff with jitter disabled (deterministic sequence); max attempts configurable; dead‑letter counter when exceeded
- [ ] Buffering: bounded edge capacity per pipeline; backpressure `dropOldest`/`dropNewest`/`block` selectable; count dropped
- [ ] TLS: v1.2+ required, system roots by default; options for custom CA/mTLS; cert‑pinning optional
- [ ] Determinism: timestamps formatted in UTC ms; no random IDs; stable JSON output; all defaults explicit
- [ ] I/O restrictions: pipelines used from allowed nodes only (e.g., egress/ingress); transforms remain I/O‑restricted per static checks
#### 2.1.8.0 Observability
- [ ] Expose counters: enqueued, sent, failed, dropped, retries, open connections
- [ ] Optional metrics hook to `eventmeta.v1` for debug builds
#### 2.1.9.0 Docs & Tests
- [x] `docs/stdlib/logger.md` with examples for current logger options; sink docs pending
- [ ] Unit tests with fakes:
  - [x] stdout/stderr capture
  - [x] file writes (tmpdir), permissions, append behavior
  - [ ] syslog over localhost TCP with minimal RFC‑5424 verifier
  - [ ] https with httptest server; verify headers, batching, TLS settings (self‑signed CA in tests)
  - [ ] amqps via in‑process fake or interface mock; no external broker dependency
- [ ] Stress tests for backpressure and explicit `dropOldest`/`dropNewest` policies
- [ ] Coverage ≥80% across the package
#### 2.1.10.0 Out of Scope (Phase 1)
- [ ] UDP syslog (port 514/udp)
- [ ] File rotation/compression and log shipping agents
- [ ] Structured logging schema negotiation with remote sinks
#### 2.2.1.0 trigger (Ingress Triggers)
> Goal: Provide a `trigger` package usable in AMI pipelines to generate events at ingress based on time, 
> filesystem, network, or device signals. These are deterministic and policy‑controlled event sources.
##### 2.2.1.1. `trigger.Timer`
- [ ] `trigger.Timer(every=<duration>[, initial_delay=<duration>][, count=<n>])`
    - [ ] Fires periodically every `every`; optional `initial_delay` before first fire; optional finite `count`.
##### 2.2.1.2. `trigger.Schedule`
- [ ] `trigger.Schedule(at=<iso8601>|cron=<spec>[, tz=UTC][, skip_missed=true])`
    - [ ] Fires at a specific ISO‑8601 UTC timestamp or on a restricted, deterministic cron subset; default TZ is UTC.
##### 2.2.1.3. `trigger.FsNotify`
- [ ] `trigger.FsNotify(paths=[<path>...], events=[create,modify,delete,rename][, debounce_ms=<n>])`
  - [ ] Fires on filesystem changes under given paths; optional debounce; platform abstraction over inotify/fsnotify.
##### 2.2.1.4. `trigger.NetListen`
- [ ] `trigger.NetListen(proto=tcp|udp, addr=<ip>, port=<n>[, max_msg=<bytes>][, allow=[<cidr>...]])`
  - [ ] Fires on inbound network messages; payload is bytes + peer info; gated by `net` capability and allowlist.
#### 2.2.1.5. Determinism & Policy
- No jitter or ambient randomness; time‑based triggers use an injected clock for tests.
- All schedules evaluated in UTC unless explicit `tz` is provided (IANA TZ; default support limited to UTC in this phase).
- Capability gating: `fs` for FsNotify, `net` for NetListen, `device` for DeviceIrq. Deny by default.
- Backpressure: honor ingress edge policy (`dropOldest`/`dropNewest` → best‑effort; `block` → at‑least‑once with bounded queues).
- Ordering: per‑trigger instance preserves event order.
#### 2.2.1.6. Examples
- Timer: `Ingress(trigger.Timer(every=1s, initial_delay=500ms))`
- Schedule: `Ingress(trigger.Schedule(at="2025-12-31T23:59:50Z"))`
- FsNotify: `Ingress(trigger.FsNotify(paths=["./data"], events=[create,modify], debounce_ms=50))`
- NetListen: `Ingress(trigger.NetListen(proto=udp, addr="127.0.0.1", port=8125, max_msg=8192, allow=["127.0.0.0/8"]))`
- DeviceIrq: `Ingress(trigger.DeviceIrq(device="/dev/input0", event="gpio", filter="pin=7"))`
#### 2.2.1.7 Compiler/Runtime Integration
- Parser/sem: `trigger.*` valid only in ingress position; validate parameters and capabilities.
- Codegen: emit trigger initializers wired to pipeline edges with configured backpressure and capacities.
- Runtime: test harness provides simulated trigger sources for deterministic tests; real I/O behind capability flags.
#### 2.2.1.8 Docs & Tests
- `docs/stdlib/trigger.md` with config, capability requirements, and examples for each trigger.
- Unit tests:
  - Timer/Schedule (clock‑injected; periodic and one‑shot; skip_missed behavior)
  - FsNotify (simulated events; path filters and debounce)
  - NetListen (loopback UDP/TCP; allowlist; `max_msg` truncation policy)
  - DeviceIrq (simulated device notifications; filter evaluation)
- CLI tests: pipelines using `Ingress(trigger.*)` emit expected events under harness simulation.
- Coverage ≥80%.
- [ ] SFTP subsystem implementation (may be added later)
- [ ] Rekey negotiation tuning and advanced multiplexing
#### 2.3.0.0. Package: csv (reader/writer)
> Goal: Provide an RFC‑4180–compliant CSV reader/writer under `ami/stdlib/csv` with deterministic formatting, 
> strict/lenient parsing modes, and AMI‑friendly streaming pipelines. UTF‑8 only; explicit configuration for 
> delimiters, quoting, headers, and newlines.
##### Types & Config
- [ ] `ReaderConfig`:
  - `Delimiter rune` (default `,`), `Quote rune` (default `"`), `Comment *rune` (optional)
  - `TrimSpace bool`, `LazyQuotes bool` (off by default), `AllowCRLF bool` (normalize to LF when false)
  - `DetectHeaders bool` (use first row as headers), `Headers []string` (explicit header order)
  - `StrictFieldCount bool` (enforce equal fields per row when headers set)
  - `MaxRecordSize int` (bytes, 0 = unlimited), `MaxFields int` (0 = unlimited)
- [ ] `WriterConfig`:
  - `Delimiter rune`, `Quote rune`, `QuoteAll bool`, `UseCRLF bool` (default false)
  - `Headers []string` (controls column order for named records)
  - `BOM bool` (default false; only at start when enabled)
- [ ] `Record` = `[]string`
- [ ] `NamedRecord` = `map[string]string` (order governed by `Headers` when writing)
##### Pipelines
- [ ] File reader: `Ingress(worker=file(path)).Transform(worker=csv.Read(readerCfg)).Egress(worker=next)`
  - Produces `Event<Record>` or `Event<NamedRecord>` (when headers present)
- [ ] Stream reader: `Ingress(worker=stream(reader)).Transform(worker=csv.Read(readerCfg)).Egress(worker=next)`
- [ ] File writer: `Ingress(Event<Record|NamedRecord>).Transform(worker=csv.Write(writerCfg)).Egress(worker=file(path))`
- [ ] Stdout writer: `Ingress(Event<Record|NamedRecord>).Transform(worker=csv.Write(writerCfg)).Egress(worker=stdout)`
##### Determinism & Policy
- [ ] UTF‑8 only; optional BOM handling controlled by config
- [ ] Stable column ordering (from `Headers`) for named records; stable quoting rules
- [ ] Newline policy: default `\n`; `UseCRLF` when requested; reader can normalize to LF when `AllowCRLF=false`
- [ ] Streaming friendly: bounded buffers; backpressure via edge policy; 
- [ ] I/O restrictions: only ingress/egress nodes perform file/stream I/O; transforms remain pure
##### Docs & Tests
- [ ] `docs/stdlib/csv.md` with examples for reader/writer (headers and plain records)
- [ ] Golden tests for writer (LF/CRLF, QuoteAll on/off, BOM); reader tests for strict/lenient, quotes/escapes, comments, trimming
- [ ] Error cases: mismatched field counts, oversized records, invalid UTF‑8, malformed quotes (strict mode)
- [ ] Coverage ≥80%
##### Out of Scope (Phase 1)
- [ ] Dialect auto‑detection beyond `DetectHeaders`
- [ ] Multi‑character delimiters or quotes
- [ ] Non‑UTF‑8 encodings
#### 2.4.0.0 Merge.Sort() Semantics (Collect)
> Source of truth: Align with the behavioral details described in `docs/Asynchronous Machine Interface.docx` 
> for Collect/merge. The following checklist captures the normative expectations we implement and test against.
- [ ] Purpose
  - [ ] `merge.Sort(field[, order])` defines the output ordering for records merged at a `Collect` node. Sorting is 
        applied within the active merge window (see `merge.Window`, `merge.Watermark`, `merge.Timeout`) and, when 
        present, within each partition (`merge.PartitionBy`).
- [ ] Field selection
  - [ ] `field` is a selector string referencing a payload attribute; dotted selectors allowed (e.g., `event.meta.ts`, `user.id`).
  - [ ] The field MUST be resolvable on the payloads being merged; missing/unresolvable fields are handled per "Null/missing semantics".
  - [ ] The field type MUST be orderable. Supported orderable types: integer, floating point, boolean, timestamp, and string. Complex types (map/set/slice/struct) are not orderable.
- [ ] Order argument
  - [ ] Optional `order` ∈ {`asc`, `desc`}; default `asc`.

- [ ] Comparators (type‑specific ordering)
  - [ ] Integer/floating point: numeric comparison; IEEE‑754 NaN (if present) sorts after all numbers (null/missing rules may still apply).
  - [ ] Boolean: `false < true` for `asc`; reversed for `desc`.
  - [ ] Timestamp: compare by event time (RFC‑3339/ISO‑8601). If multiple encodings are permitted per docx, normalize to epoch nanoseconds for comparison.
  - [ ] String: binary (UTF‑8 code‑point) comparison; no locale/collation transforms. Case‑sensitive. Deterministic across platforms.

- [ ] Null/missing semantics
  - [ ] Records with missing or null `field` sort after records with present values by default (nulls‑last) for both `asc` and `desc` unless the docx specifies otherwise.
  - [ ] If a future attribute (e.g., `merge.NullsFirst()`) is introduced by the docx, adopt the documented precedence; until then, use nulls‑last deterministically.

- [ ] Windowing and watermarks
  - [ ] Sorting is applied to the current merge window only (bounded by `merge.Window`, `merge.Timeout`, and/or `merge.Watermark`).
  - [ ] `merge.Watermark(field, lateness)` defines when late/out‑of‑order records may be considered; watermarks trigger window closure and emission.
  - [ ] Records arriving after window closure are handled according to docx rules (e.g., late arrival policy); scaffolding may drop late or emit to the next window deterministically.

- [ ] Partitioning
  - [ ] When `merge.PartitionBy(field)` is present, sorting is applied independently within each partition key.

- [ ] Stability and tiebreaks
  - [ ] With `merge.Stable()`: the sort is stable; equal sort keys preserve their pre‑sort relative order as defined by deterministic arrival index within the active window/partition.
  - [ ] Without `merge.Stable()`: implementation may use an unstable sort, but MUST remain deterministic for identical inputs on the same platform/toolchain.
  - [ ] Secondary tiebreak: if `merge.Key(field)` is defined and its type is orderable, it MAY be used as a secondary key to enhance determinism; otherwise use (partition, upstreamIndex, arrivalIndex) within the window.

- [ ] Interaction with buffering/backpressure
  - [ ] `merge.Buffer(capacity[, backpressure])` constrains the internal buffer used for sorting within a window/partition.
  - [ ] `backpressure=block`: upstreams block on capacity; order is preserved within the window.
- [ ] `backpressure=dropOldest|dropNewest`: records may be dropped when capacity is exceeded; the output remains sorted among retained records, but global order may not reflect the hypothetical full set; emit `W_MERGE_TINY_BUFFER` when `capacity<=1`. Warn on ambiguous `drop`.

- [ ] Determinism & repeatability
  - [ ] For identical inputs, same attributes, and equal windowing, the output ordering MUST be deterministic across runs, architectures, and OSes.
  - [ ] No dependency on locale/timezone; all time and string comparisons are normalized and reproducible.

- [ ] Diagnostics (Sort‑specific)
  - [ ] `E_MERGE_SORT_FIELD_UNORDERABLE`: the selected field is not orderable (complex type).
  - [ ] `E_MERGE_SORT_FIELD_UNKNOWN`: field cannot be resolved in the input payload shape.
  - [ ] `E_MERGE_SORT_ORDER_INVALID`: `order` value not in {asc, desc}.
  - [ ] `W_MERGE_SORT_NO_FIELD`: attr provided without a field argument.

- [ ] Tests (Sort)
  - [ ] Type coverage: int/float/bool/time/string for both `asc` and `desc`.
  - [ ] Null/missing behavior: nulls‑last determinism; explicit docx behavior tested when specified.
  - [ ] Stability: equal keys with/without `merge.Stable()`; tiebreak determinism.
  - [ ] Partitioned sorting: independent order within each partition.
  - [ ] Windowing: emission boundaries via watermark/timeout; late arrival policy.
  - [ ] Buffer/backpressure interaction: block vs `dropOldest`/`dropNewest`; retained‑record sort correctness; lint smell for tiny buffers with `dropOldest`/`dropNewest`.

## Remaining Backend Effort
> Goal: `ami build` compiles AMI sources → AMI IR → LLVM → darwin/arm64 Mach‑O binary. This checklist captures the concrete backend tasks to reach that outcome while preserving existing front‑end behavior and debug artifacts.

- [ ] Backend Architecture Baseline
  - [ ] Freeze `ir` module surface for M1: ops, value model, function/block shapes, and determinism rules (stable JSON order).
  - [ ] Add concise package docs for `ir` describing lowering invariants used by codegen (no raw pointers exposed; SSA temps are internal only).

- [ ] LLVM Emitter (IR → LLVM `.ll`)
  - [ ] Introduce `src/ami/compiler/codegen/llvm` package (one concept per file):
    - [ ] `module.go`: create LLVM module, set target triple `arm64-apple-macosx` (exact min version detected from host or defaulted; deterministic fallback).
    - [ ] `types.go`: map AMI types to LLVM types (ints, bool, real, tuple as structs, containers as opaque runtime handles without exposing raw pointers across AMI surface).
    - [ ] `func.go`: declare/define functions from `ir.Function` with params/results; build entry block.
    - [ ] `instr_*.go`: lower `VAR`, `ASSIGN`, `RETURN`, `DEFER`, `EXPR(call,bin,unary)` to LLVM using a builder; ensure deterministic emission order.
    - [ ] `extern.go`: declare required runtime symbols (event, state, panic, alloc) as opaque functions; no user‑visible pointers in ABI.
    - [ ] `emit.go`: write textual LLVM IR; when `--verbose`, save as `build/debug/llvm/<pkg>/<unit>.ll`.
  - [ ] Tests (golden): emit `.ll` for minimal functions (var/assign/return, calls), container literals scaffold; verify stable output across runs.
  - [ ] Diagnostics: add `E_LLVM_EMIT` with file/line context when emission fails.

- [ ] Object Generation (LLVM → .o)
  - [ ] Tool detection: locate `clang` (preferred) and optionally `llc`; record versions in verbose logs.
  - [ ] Compile `.ll` to `.o` for `darwin/arm64` via `clang -target arm64-apple-macosx -c` into `build/obj/<pkg>/<unit>.o`.
  - [ ] Update object index to include real `.o` entries (retain `.s` only as debug when verbose).
  - [X] Tests: conditional E2E (skip if toolchain missing) that compiles a trivial unit to `.o` and updates `objindex.v1` deterministically.
  - [ ] Diagnostics: `E_TOOLCHAIN_MISSING` (no clang), `E_OBJ_COMPILE_FAIL` with captured stderr in JSON.

- [ ] Runtime ABI (minimal) for codegen
  - [ ] Define a small, deterministic runtime shim for event/state and process entry, built for `darwin/arm64`:
    - [ ] `rt/abi.h` and `rt/abi.ll` (or equivalent Go‑generated `.ll`) modeling `Event<T>` as a value/handle with no raw pointer exposure across AMI boundaries.
    - [ ] Expose only opaque handles/functions (`ami_rt_*`) used by the emitter; keep malloc/free encapsulated inside runtime.
  - [ ] Build or vendor a precompiled runtime object/bitcode for `darwin/arm64` checked into the repo or reproducibly built during tests (deterministic flags).
  - [ ] Tests: unit tests for symbol presence and linkability; skip when host toolchain absent.

- [ ] Linking (Mach‑O)
  - [ ] Link all package objects plus runtime into a single binary under `build/darwin/arm64/<project>` using `clang` with `-target arm64-apple-macosx`.
  - [ ] Deterministic flags: strip timestamps/non‑deterministic sections; enable dead‑strip for unused code where safe.
  - [ ] Update `ami.manifest` to list produced binaries and their relative paths; keep stable ordering.
  - [ ] Diagnostics: `E_LINK_FAIL` with tool stderr; include command/args in verbose logs only.
  - [ ] Tests: conditional E2E that links a hello‑pipeline binary and asserts it runs and exits 0 (or prints an expected line); skipped when clang is unavailable.

- [ ] Build Driver Integration
  - [ ] Extend `driver.Compile` (or add a `Build` orchestration) to run: lower → emit `.ll` (verbose) → compile `.o` → link (per env target).
  - [ ] Respect `toolchain.compiler.env`; for now, support only `darwin/arm64` and place outputs under `build/darwin/arm64/`.
  - [ ] Add flags (future‑compatible) for `--emit-llvm-only` and `--no-link`; wire but keep defaults to full build.
  - [ ] Ensure debug artifacts (AST/IR/LLVM/ASM) are only emitted with `--verbose` per existing policy.

- [ ] Memory Safety (AMI 2.3.2) Enforcement in Backend
  - [ ] Assert no public ABI uses raw pointers; internal `alloca` allowed within a function, never leaked.
  - [ ] For containers and events, pass by value or by opaque handle validated by runtime; add sad‑path tests that would previously expose pointers and ensure emitter rejects them.

- [ ] Determinism & Reproducibility
  - [ ] Normalize builder output ordering; stabilize `.ll` via sorted traversal and fixed attribute ordering.
  - [ ] Ensure indexes/manifests avoid embedding variable timestamps; tests compare across runs.

- [ ] CLI/UX and Diagnostics
  - [ ] Human summary: `built <n> objects; linked 1 binary → build/darwin/arm64/<name>`; JSON success record includes `binaries` and `objIndex`.
  - [ ] Stream diagnostics as `diag.v1` for tool errors; map to exit codes: emit (1), io (2), integrity (3), toolchain (2), link (2).

- [ ] Documentation
  - [ ] Add `docs/backend/README.md` covering: IR invariants, LLVM mapping, toolchain requirements, determinism flags, and how to reproduce a build.
  - [ ] Update examples to include a minimal program and `make examples` verification once the backend can link.

- [ ] Coverage & Vetting
  - [ ] ≥80% coverage for `codegen/llvm` and driver additions; include happy/sad tests.
  - [ ] `go vet ./...` clean; build checks updated to ensure `ami build` passes on hosts with/without LLVM (graceful skip with clear diagnostics).
