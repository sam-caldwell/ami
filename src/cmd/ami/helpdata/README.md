AMI Help

Overview

- The `ami` CLI manages AMI workspaces, packages, linting, testing, and builds.
- Global flags: `--json`, `--verbose`, `--color` (mutually exclusive with `--json`).

Getting Started

- Initialize a workspace: run `ami init` in your project root to create `ami.workspace`.
- Lint your code: run `ami lint` (use `--strict` to treat warnings as errors).
- Run tests: run `ami test` (use `--verbose` to write logs under `build/test/`).
- Build the project: run `ami build` to produce artifacts in `build/`.

Commands

- `ami help`             : show this help guide
- `ami version`          : print CLI version
- `ami init`             : create or update `ami.workspace`
- `ami clean`            : clean and recreate `./build`
- `ami mod clean`        : clean and recreate the package cache
- `ami mod list`         : list cached packages
- `ami mod audit`        : audit workspace imports vs `ami.sum` and cache
- `ami mod sum`          : validate cache against `ami.sum`
- `ami mod update`       : copy workspace packages into the cache and refresh `ami.sum`
- `ami lint`             : run linter checks
- `ami test`             : run project tests
- `ami build`            : build the workspace
- `ami pipeline visualize` : render ASCII pipeline graphs (stub)

Examples

- Initialize a workspace:
  - `ami init`
  - `ami init --force`
- Build the project:
  - `ami build`
  - `ami build --verbose`
  - `ami build --json`
- Lint sources:
  - `ami lint --strict`
  - `ami lint --json`
- Run tests:
  - `ami test`
  - `ami test --verbose`
  - `ami test --json`
- Module operations:
  - `ami mod list --json`
  - `ami mod get ./vendor/alpha`
  - `ami mod update --json`
  - `ami mod sum --json`
  - `ami mod clean --json`
- Pipelines:
  - `ami pipeline visualize`
  - `ami pipeline visualize --json --no-summary`

Notes

- Human output goes to stdout; errors to stderr.
- JSON mode writes NDJSON for streaming commands.
- See docs/Asynchronous Machine Interface.docx for the language and runtime semantics.
