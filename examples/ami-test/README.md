## AMI Test Directives Example

This example shows how to use AMI test directives with `ami test`.

- `ok/`: contains a simple `.ami` file that parses successfully with a test case.
- `fail/`: contains a `.ami` file expected to fail parsing with a specific message/code.

Run:

- Human: run `ami test --verbose` inside each directory.
- JSON: run `ami test --json` to stream Go test events, AMI directive events, and a final summary.

Artifacts (with `--verbose`):

- `build/test/test.log`: raw `go test -json` output
- `build/test/test.manifest`: entries like `example.com/pkg TestX` and `ami:<relpath> <case>`

Directive examples:

- `#pragma test:case simple`
- `#pragma test:assert parse_ok`
- `#pragma test:assert parse_fail msg="expected 'package'" code=E_PARSE`

Positions (optional):

- Add `line`, `column`, or `offset` to `test:assert` to assert a specific error position.

