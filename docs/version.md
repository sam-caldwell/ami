# ami version

Prints the CLI version.

## Usage

- Human: `ami version`
- JSON: `ami --json version`

## Output

- Human: `version: vX.Y.Z`
- JSON: diagnostic record with `data.version = "vX.Y.Z"`.

Notes:
- Version is injected at build time via `-ldflags` in release builds. Defaults to `v0.0.0-dev` in development.
