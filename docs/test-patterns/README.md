# Test Patterns: End-to-End (E2E) CLI Tests

This guide documents the end-to-end test pattern used for the `ami` CLI. The goal is to validate commands using the built binary, interacting through standard streams in a deterministic and fully automated way.

## Principles
- Black-box: exercise the built `ami` binary via `exec.Command`.
- Deterministic: no prompts, no timing races, stable outputs (human or JSON).
- Hermetic: test inputs live under `build/test/...`; no global state is required beyond standard env vars.
- Scope-safe: avoid changing unrelated code; keep helpers local to tests.

## Pattern Overview
1. Build the CLI into `build/ami` using `go build -o build/ami ./src/cmd/ami`.
2. Stage a temporary workspace under `build/test/e2e/...` with the necessary files (e.g., `ami.workspace`, `ami.sum`).
3. Spawn the binary with `exec.Command`:
   - Set `cmd.Dir` to the staged workspace.
   - Pass flags (e.g., `--json`).
   - Wire `cmd.Stdin`, `cmd.Stdout`, and `cmd.Stderr`.
   - Override environment for isolation (e.g., `AMI_PACKAGE_CACHE`).
4. Run the process, assert exit success, and validate outputs:
   - JSON mode: unmarshal and assert fields.
   - Human mode: search for stable summary lines.

## Streams & Env
- `stdin`: Provide an empty reader (`io.NopCloser(bytes.NewReader(nil))`) to prove non-interactive behavior.
- `stdout`: Capture for assertions (JSON or human strings).
- `stderr`: Should be empty on success; assert emptiness in happy path tests.
- `AMI_PACKAGE_CACHE`: Use an absolute path inside the test’s workspace to avoid host-level side effects.

## File/Path Conventions
- Workspace root for a test: `build/test/e2e/<suite>/<name>`.
- Cache dir (when needed): `build/test/e2e/<suite>/<name>/cache`.
- Avoid writing outside `build/test/...` to keep tests self-contained.

## JSON vs Human Output
- Prefer JSON for machine checks: decode into a small struct and validate fields.
- For human output, assert on concise, stable summary text (e.g., `ok:` or `missing in sum:`), not full verbose logs.

## Example Skeleton (Go)
```go
bin := buildAmi(t) // builds ./src/cmd/ami into build/ami
cmd := exec.Command(bin, "mod", "audit", "--json")
cmd.Dir = wsDir
cmd.Env = append(os.Environ(), "AMI_PACKAGE_CACHE="+absCache)
cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
var stdout, stderr bytes.Buffer
cmd.Stdout, cmd.Stderr = &stdout, &stderr
if err := cmd.Run(); err != nil { t.Fatalf("run: %v\n%s", err, stderr.String()) }
// JSON assertion
var res struct { SumFound bool; MissingInSum []string }
if err := json.Unmarshal(stdout.Bytes(), &res); err != nil { t.Fatalf("json: %v", err) }
```

## Tips
- Make paths absolute when exporting via env (e.g., `AMI_PACKAGE_CACHE`) to avoid surprises from process `Dir` changes.
- Keep hashing or content-generation helpers inline in the e2e test when needed; do not import internal packages into e2e to preserve black-box behavior.
- Keep output assertions resilient to timestamps by focusing on stable fields/text.

## What Not To Do
- Do not rely on network access or external services.
- Do not depend on the developer’s global environment or HOME structure.
- Do not couple to internal packages for behavior under test; use the CLI only.

## Where To Look
- Example e2e: `tests/e2e/ami_mod_audit_test.go` — builds the binary, runs `ami mod audit` in JSON and human modes, and validates outputs.

