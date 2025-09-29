# ami test

Runs your project’s tests and also checks AMI parser directives inside `.ami` files. Use this to verify both your Go tests and AMI syntax expectations.

Quick start:
- `ami test` — run tests in the current folder.
- `ami test path/to/project` — run tests in a different folder.
- `ami test --verbose` — also write logs and a manifest under `build/test/`.
- `ami test --json` — stream JSON events and a final summary (NDJSON).

Flags:
- `--json` — stream Go `-json` events, AMI directive events, and runtime events; then emit a final summary object.
- `--verbose` — write artifacts under `build/test/`:
  - `test.log` — raw `go test -json` output.
  - `test.manifest` — one line per Go test (`<pkg> <TestName>`), then AMI entries (`ami:<relpath> <case>`).
- `--packages N` — set Go test package concurrency (`go test -p N`). `0` uses the default.
- `--timeout MS` — default runtime test timeout in milliseconds (`0` disables).
- `--parallel N` — runtime test concurrency (`0` = serial).
- `--failfast` — stop after the first failing runtime test.
- `--run REGEX` — run only runtime tests whose names match the regex.

What runs:
- Go tests: `go test -json ./...` with optional package concurrency.
- AMI directive tests: for each `*.ami` file, look for pragmas that define parser expectations.
- Runtime tests: discover `*_test.ami` and run cases defined by runtime pragmas (see below).

Human output:
- Per‑package lines: `test: pkg <package> ok=N fail=M`.
- AMI summary when directives are present: `test: ami ok=X fail=Y`.
- `test: OK` when everything passes.

JSON output (NDJSON):
- Raw `go test -json` events.
- AMI directive events (`schema=ami.test.v1`):
  - `{ "schema":"ami.test.v1", "file":"<relpath>", "case":"<name>", "ok":true|false, "expect":"parse_ok|parse_fail", "count":<n|-1>, "gotErrs":<n>, "msg":"<substr>", "code":"E_PARSE", "line":<n>, "column":<n>, "offset":<n> }`
- Per‑package summaries (`schema=ami.test.pkg.v1`):
  - `{ "schema":"ami.test.pkg.v1", "package":"<pkg>", "ok":N, "fail":M }`
- Final summary object:
  - `{ "ok":bool, "packages":N, "tests":N, "failures":N, "ami_tests":N, "ami_failures":N, "runtime_tests":N, "runtime_failures":N, "runtime_skipped":N }`

Runtime JSON events (test.v1):
- `{"schema":"test.v1","type":"run_start","timeout_ms":MS,"parallel":N}`
- `{"schema":"test.v1","type":"test_start","file":"<rel>","case":"<name>"}`
- `{"schema":"test.v1","type":"test_end","file":"<rel>","case":"<name>","ok":true|false,"skipped":bool,"duration_ms":N,"error":"<msg>"}`
- `{"schema":"test.v1","type":"run_end","runtime_tests":N,"runtime_failures":N,"runtime_skipped":N,"duration_ms":N}`

AMI test pragmas inside `.ami` files:
- `#pragma test:case <name>` — declares a test case name.
- `#pragma test:assert parse_ok|parse_fail [count=N] [msg="substr"] [code=E_PARSE] [line=N] [column=N] [offset=N]` — sets expectations:
  - `parse_ok` — file should parse without errors.
  - `parse_fail` — file should have parse errors.
  - `count=N` — expect exactly N parser errors (optional).
  - `msg=substr` — at least one error message contains this text (optional).
  - `code=E_PARSE` — parser errors use `E_PARSE` (optional; reserved for future expansion).
  - `line/column/offset` — at least one error matches these positions (optional).

Runtime test pragmas in `*_test.ami`:
- `#pragma test:case <name>` — declares a runtime test case name.
- `#pragma test:skip <reason>` — marks all cases in the file as skipped.
- `#pragma test:fixture path=<rel> mode=<ro|rw>` — declares a fixture file; validated before run.
- `#pragma test:runtime input=<json> [output=<json>] [expect_error=<CODE>] [timeout=MS]`
  - Input/Output are JSON snippets (avoid spaces unless quoted).
  - Identity harness by default (output equals input).
  - Reserved input keys: `sleep_ms` (delay), `error_code` (force error for sad paths).
  - `timeout` overrides the default `--timeout` for the file’s cases.

Exit codes:
- Success: 0 when all Go tests pass and all AMI directive and runtime cases pass.
- Failure: non‑zero when any Go test fails, any AMI directive case fails, or any runtime case fails.

Troubleshooting:
- "no packages to test": there may be no Go packages. This is OK; AMI directive tests still run.
- Where are logs? With `--verbose`, check `build/test/test.log` and `build/test/test.manifest`.
- Nothing printed? In success cases, human output ends with `test: OK`. Use `--json` for machine parsing.

## See Also
- `docs/test-patterns/README.md` — end-to-end CLI testing patterns used by the e2e suite.
- `docs/events.md` — background on event and error schemas referenced by test outputs.
