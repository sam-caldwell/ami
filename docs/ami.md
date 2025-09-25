# AMI CLI

A deterministic, scriptable CLI for managing AMI workspaces, dependencies, linting, testing, and building.

## Install / Build

- Build from source: `go build -o build/ami ./src/cmd/ami`
- Run: `./build/ami --help`

## Global Flags

- `--json`: emit machine‑parsable JSON. One object per line for streaming commands.
- `--verbose`: add timestamps to human output; enable extra details in some commands.
- `--color`: enable ANSI colors for human output (cannot be used with `--json`).
- `--help`: show help.

Rules:
- `--json` and `--color` are mutually exclusive. If both are provided, AMI exits with code 1 and prints a plain‑text error to stderr.
- Human output goes to stdout; errors to stderr. In `--json` mode, all records go to stdout with schema `diag.v1`.

## Exit Codes

- `0`: success
- `1`: user error (bad flags/args/inputs)
- `2`: system I/O error
- `3`: integrity violation (digest mismatch)
- `4`: network/registry error

## Common Examples

- Create a workspace: `./build/ami init`
- Clean build directory: `./build/ami clean`
- Fetch a module: `./build/ami mod get git+ssh://git@github.com/org/repo.git#v1.2.3`
- List cached modules: `./build/ami mod list` or `--json`
- Update dependencies from workspace: `./build/ami mod update`
- Verify dependency digests: `./build/ami mod verify`
- Build: `./build/ami build` (use `--verbose` for debug artifacts)
- Print version: `./build/ami version` (works with `--json`)

For details on each command, see the files in `docs/` like `mod.md`, `build.md`, and `workspace.md`.

