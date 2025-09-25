# ami test

Runs Go tests for the given packages and streams structured results. This command wraps `go test -json` to provide stable human output and a machine‑readable JSON event stream (`test.v1`).

## Usage

- Human: `ami test ./...`
- JSON: `ami --json test ./...`
- Flags:
  - `--timeout <dur>`: per-package timeout (e.g., `50ms`, `2s`, `1m`)
  - `--parallel <n>`: limit parallelism within package (default Go behavior)
  - `--pkg-parallel <n>`: number of packages to test in parallel (maps to `go test -p`)
  - `--failfast`: stop after first test failure
  - `--run <regex>`: run only tests matching regex

If no packages are provided, `./...` is used by default.

## Output

- Human: concise lines per test and a summary at the end.
  - Includes per‑package lines: `test: package <name> — <pass> pass, <fail> fail, <skip> skip, <cases> cases`.
- JSON (`test.v1` JSON Lines):
  - `run_start`: overall run header
    - includes optional fields: `timeout`, `parallel`, `pkg_parallel`, `failfast`, `run`
  - `test_start`: individual test start
  - `test_output`: captured stdout from a test
  - `test_end`: individual test end with `status: pass|fail|skip`
  - `run_end`: totals, per‑package summaries, and duration
    - `totals`: global pass/fail/skip/cases
    - `packages`: array of `{ package, pass, fail, skip, cases }` sorted by package name

## Exit Codes

- 0: all tests pass
- 1: at least one test failed
- 2: test invocation/build error (system I/O)

## Notes

- Native AMI tests (Phase 1): the runner discovers `*_test.ami` under workspace packages and evaluates directive‑driven assertions without executing pipelines.
   - Supported pragmas at file scope:
     - `#pragma test:case <name>`: starts a new test case in this file
     - `#pragma test:expect_no_errors`: asserts parser/semantic analysis emit no errors
     - `#pragma test:expect_error <CODE>`: asserts an error diagnostic with code exists (e.g., `E_BAD_IMPORT`)
     - `#pragma test:expect_warn <CODE>`: asserts a warning diagnostic with code exists
     - `#pragma test:skip <reason>`: marks the current case as skipped
   - If no pragmas are present, a default case named after the file asserts no errors.
   - JSON mapping: each case emits `test_start` and `test_end` events with `package` = AMI package and `name` per case.
   - Diagnostics on failure are attached to `test_end.diagnostics` as `diag.v1` entries.

- Runtime AMI tests (Phase 2): executable cases via `#pragma test:runtime ...` with deterministic harness.
  - `#pragma test:runtime pipeline=<name> input=<json> expect_output=<json> [timeout=<ms>]`
  - `#pragma test:runtime expect_error=<CODE> [timeout=<ms>]`
  - `#pragma test:fixture path=<rel> mode=<ro|rw>` to attach fixtures (validated; enforcement deferred)
  - Reserved input keys interpreted by harness: `sleep_ms`, `error_code`.
