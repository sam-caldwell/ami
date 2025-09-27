# ami test

Runs Go tests in the workspace and evaluates AMI directive-based tests embedded in `.ami` files.

Usage:

- `ami test [path] [--json] [--verbose] [--packages N]`

Flags:

- `--json`: stream Go `-json` events and AMI test events (NDJSON), then emit a final summary object.
- `--verbose`: write artifacts under `build/test/`:
  - `test.log`: raw `go test -json` output
  - `test.manifest`: per-test entries (`<pkg> <TestName>`) and AMI entries (`ami:<relpath> <case>`)
- `--packages N`: set Go test package concurrency (`go test -p N`). `0` uses the default.

Human output:

- Prints per-package summaries: `test: pkg <package> ok=N fail=M`
- Prints AMI summary when present: `test: ami ok=X fail=Y`
- Prints `test: OK` when all tests (Go + AMI directives) pass.

JSON output:

- Streams Go `-json` events directly.
- Streams AMI directive events with schema `ami.test.v1`:
  - `{ "schema":"ami.test.v1", "file":"<relpath>", "case":"<name>", "ok":true|false, "expect":"parse_ok|parse_fail", "count":<n|-1>, "gotErrs":<n>, "msg":"<substr>", "code":"E_PARSE", "line":<n>, "column":<n>, "offset":<n> }`
- Streams per-package summary events with schema `ami.test.pkg.v1`:
  - `{ "schema":"ami.test.pkg.v1", "package":"<pkg>", "ok":N, "fail":M }`
- Emits a final summary object:
  - `{ "ok":bool, "packages":N, "tests":N, "failures":N, "ami_tests":N, "ami_failures":N }`

Directives in `.ami` files:

- `#pragma test:case <name>`: declares a test case.
- `#pragma test:assert parse_ok|parse_fail [count=N] [msg="substr"] [code=E_PARSE] [line=N] [column=N] [offset=N]`:
  - `parse_ok`: file should parse without errors.
  - `parse_fail`: file should have parse errors.
  - `count=N`: expect exactly N parser errors (optional).
  - `msg=substr`: at least one error message must contain `substr` (optional).
  - `code=E_PARSE`: all parser errors use code `E_PARSE` (optional; future codes may appear).
  - `line/column/offset`: at least one error must match the provided position(s) (optional).

Notes:

- AMI directives are parser-level checks (no runtime execution).
- Tests are deterministic and avoid interactive prompts.
